package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"os"

	// "os"
	"time"

	"log"

	merkleTree "go-merkle-file-transfer/merkle"
	pb "go-merkle-file-transfer/protos"

	"google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure"
)

var root *merkleTree.Node
var mt = merkleTree.NewMerkleTree()

func hashNodes(a, b []byte) []byte {
	data := append(a, b...)
	hash := sha256.Sum256(data)
	return hash[:]
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

func uploadFile(client pb.FileTransferClient, fileName string, content []byte) error {
	log.Printf("Uploading file: %s\n", fileName)
	_, err := client.UploadFile(context.Background(), &pb.FileData{Name: fileName, Content: content})
	if err != nil {
		return err
	}

	mt.AddFile(content)
	root, err = mt.ComputeRoot()
	if err != nil {
		log.Fatalf("Failed to compute Merkle root: %v", err)
	}

	return nil
}

func downloadFile(client pb.FileTransferClient, fileName string) ([]byte, error) {
	response, err := client.DownloadFile(context.Background(), &pb.FileName{Name: fileName})
	if err != nil {
		return nil, err
	}

	if !verifyMerkleProof(response.Content, response.MerkleProof, root.Hash) {
		return nil, errors.New("File integrity check failed")
	}

	return response.Content, nil
}

// Reads file from a given location and returns its content
func getFileFromLocation(filePath string) ([]byte, error) {
    return os.ReadFile(filePath)
}

func main() {
	const maxRetries = 5
	var conn *grpc.ClientConn

	serverAddr, ok := os.LookupEnv("SERVER_ADDR")
	if !ok {
		serverAddr = "server1:5001" // default address
	}

	for i := 0; i < maxRetries; i++ {
		connec, err := grpc.Dial(serverAddr, grpc.WithInsecure())
		if err == nil {
			conn = connec
			log.Println("Successfully connected to server")
			break
		}
		log.Printf("Failed to connect, retrying... (Attempt %d)\n", i+1)
		time.Sleep(2 * time.Second) // Wait before retrying
	}

	if conn == nil {
		log.Fatalf("Failed to connect after %d attempts", maxRetries)
	}

	defer conn.Close()

	client := pb.NewFileTransferClient(conn)

	err := uploadFile(client, "example.txt", []byte("This is an example"))
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}
	err = uploadFile(client, "example2.txt", []byte("This is an example"))
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}
	err = uploadFile(client, "example3.txt", []byte("This is an exampl,e2"))
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}
	err = uploadFile(client, "example3s.txt", []byte("This is an exasmpl,e2"))
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}
	log.Printf("\n\n")

	content, err := downloadFile(client, "example3.txt")
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	log.Printf("Downloaded content: %s", string(content))

}
