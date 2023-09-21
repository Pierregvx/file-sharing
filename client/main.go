package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"strings"

	"database/sql"
	"time"

	_ "github.com/lib/pq"

	"log"

	merkleTree "go-merkle-file-transfer/merkle"
	pb "go-merkle-file-transfer/protos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var root *merkleTree.Node
var mt = merkleTree.NewMerkleTree()

func hashNodes(a, b []byte) []byte {
	data := append(a, b...)
	hash := sha256.Sum256(data)
	return hash[:]
}

func AddLeafToDB(db *sql.DB, leafHash []byte) error {
    sqlStatement := `
    INSERT INTO merkle_leaves (leaf_content)
    VALUES ($1)
    RETURNING id`
    id := 0
    err := db.QueryRow(sqlStatement, leafHash).Scan(&id)
    if err != nil {
        return fmt.Errorf("Failed to insert leaf: %v", err)
    }
    log.Printf("New leaf inserted with id: %d", id)
    return nil
}

// RestoreTree fetches all the leaves from the database and adds them to the MerkleTree
func RestoreTree(db *sql.DB) error {
	// Query the database for all leaves
	rows, err := db.Query("SELECT leaf_content FROM merkle_leaves")
	if err != nil {
		return err
	}
	defer rows.Close()

	// Read the rows and append to a slice
	var leaves [][]byte
	for rows.Next() {
		var leafHash string
		if err := rows.Scan(&leafHash); err != nil {
			return err
		}
		leaves = append(leaves, []byte(leafHash))
	}

	// Check for errors in iterating over rows.
	if err := rows.Err(); err != nil {
		return err
	}

	// Add leaves to MerkleTree
	mt.AddLeaves(leaves)
	return nil
}

func verifyMerkleProof(content []byte, merkleProof [][]byte, storedRoot []byte) bool {
	// Initialize with the leaf hash
	currentHash := sha256.Sum256(content)
	currentBytes := currentHash[:]

	// Traverse the proof, combining and hashing at each step
	for _, proofHash := range merkleProof {
		// Compare current hash and proof hash to determine ordering
		first, second := currentBytes, proofHash
		if bytes.Compare(first, second) > 0 {
			first, second = second, first
		}

		combined := append(first, second...)
		newHash := sha256.Sum256(combined)
		currentBytes = newHash[:]
	}
	// Compare the final calculated root with the stored root
	return bytes.Equal(currentBytes, storedRoot)
}

func uploadFile(client pb.FileTransferClient, fileName string, content []byte,db *sql.DB) error {
	log.Printf("Uploading file: %s\n", fileName)
	_, err := client.UploadFile(context.Background(), &pb.FileData{Name: fileName, Content: content})
	if err != nil {
		return err
	}

	err=mt.AddFile(content)
	if err != nil {
		log.Fatalf("Failed to add file to Merkle tree: %v", err)
	}

	err=AddLeafToDB(db,mt.Leaves[len(mt.Leaves)-1].Hash)
	if err != nil {
		log.Fatalf("Failed to add leaf to database: %v", err)
	}

	return nil
}

func downloadFile(client pb.FileTransferClient, fileName string,db *sql.DB) ([]byte, error) {
	response, err := client.DownloadFile(context.Background(), &pb.FileName{Name: fileName})
	if err != nil {
		return nil, err
	}
	err=RestoreTree(db)
	if err != nil {
		log.Fatalf("Failed to restore Merkle tree: %v", err)
	}
	root, err = mt.ComputeRoot()
	if err != nil {
		log.Fatalf("Failed to compute Merkle root: %v", err)
	}
	if !verifyMerkleProof(response.Content, response.MerkleProof, root.Hash) {
		return nil, fmt.Errorf("Merkle proof verification failed")
	}

	return response.Content, nil
}
func getFileNameFromPath(filePath string) string {
	segments := strings.Split(filePath, "/")
	return segments[len(segments)-1]
}

// Reads file from a given location and returns its content
func getFileFromLocation(filePath string) ([]byte, error) {

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Could not read file: %v", err)
	}
	return content, nil
}

// Uploads multiple files based on a list of file paths
func uploadFiles(client pb.FileTransferClient, filePaths []string,db *sql.DB) {
	for _, filePath := range filePaths {
		content, err := getFileFromLocation(filePath)
		if err != nil {
			log.Printf("Could not read file %s: %v", filePath, err)
			continue // Skip to the next file
		}
		fileName := getFileNameFromPath(filePath)
		err = uploadFile(client, fileName, content,db)
		if err != nil {
			log.Printf("Upload failed for file %s: %v", filePath, err)
		}
	}
}
func initDB(connStr string) *sql.DB {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatalf("Failed to connect to the database: %v", err)
    }
    return db
}

func initGRPCClient(serverAddr string) *grpc.ClientConn {
    const maxRetries = 5
    
    for i := 0; i < maxRetries; i++ {
        conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
        if err == nil {
            log.Println("Successfully connected to server")
            return conn
        }
        log.Printf("Failed to connect, retrying... (Attempt %d)\n", i+1)
        time.Sleep(2 * time.Second) // Wait before retrying
    }
    log.Fatalf("Failed to connect after %d attempts", maxRetries)
    return nil
}
func handleOperation(operation string, filePathList []string, client pb.FileTransferClient, db *sql.DB) {
    switch operation {
    case "upload":
        uploadFiles(client, filePathList, db)
    case "download":
        fileName := getFileNameFromPath(filePathList[0])
        content, err := downloadFile(client, fileName, db)
        if err != nil {
            log.Fatalf("Download failed: %v", err)
        }
        log.Printf("Downloaded content: %s", string(content))
    default:
        log.Fatalf("Invalid operation: %s", operation)
    }
}

func main() {
    connStr := "host=client-db port=5432 user=clientuser password=clientpassword dbname=clientdb sslmode=disable"
    db := initDB(connStr)
    defer db.Close()

    operation := flag.String("operation", "", "Operation to perform: upload or download")
    filePaths := flag.String("filePaths", "", "Comma-separated list of paths to the files to upload")
    flag.Parse()

    if *operation == "" || *filePaths == "" {
        log.Fatalf("Both 'operation' and 'filePaths' must be specified.")
    }

    filePathList := strings.Split(*filePaths, ",")
    
    serverAddr, ok := os.LookupEnv("SERVER_ADDR")
    if !ok {
        serverAddr = "server1:5001" // default address
    }
    conn := initGRPCClient(serverAddr)
    defer conn.Close()

    client := pb.NewFileTransferClient(conn)
    handleOperation(*operation, filePathList, client, db)
}
