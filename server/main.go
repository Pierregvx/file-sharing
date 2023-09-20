package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	_ "github.com/lib/pq" // The underscore is important
	merkleTree "go-merkle-file-transfer/merkle"
	pb "go-merkle-file-transfer/protos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileTransferServer struct {
	pb.UnimplementedFileTransferServer
	MerkleTree *merkleTree.MerkleTree
	DB         *sql.DB  // Add this line
}


var merkletree = merkleTree.NewMerkleTree()

func (s *FileTransferServer) UploadFile(ctx context.Context, in *pb.FileData) (*pb.UploadStatus, error) {
    log.Printf("Received UploadFile request for file: %s\n", in.GetName())
    log.Printf("File content: %x", in.GetContent())

    
    // Update the Merkle tree
    s.MerkleTree.AddFile(in.GetContent())

    // Prepare SQL statement to insert file content and metadata into the database
    stmt, err := s.DB.Prepare("INSERT INTO file_storage(file_name, file_content) VALUES($1, $2)")
    if err != nil {
        log.Printf("Failed to prepare SQL statement: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal Server Error")
    }
    defer stmt.Close()

    // Execute the SQL statement
    _, err = stmt.Exec(in.GetName(), in.GetContent())
    if err != nil {
        log.Printf("Failed to execute SQL statement: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal Server Error")
    }

    return &pb.UploadStatus{Success: true}, nil
}


func (s *FileTransferServer) DownloadFile(ctx context.Context, in *pb.FileName) (*pb.FileDownloadResponse, error) {
	log.Printf("Received DownloadFile request for file: %s\n", in.GetName())

	// Prepare SQL statement to fetch file content and metadata
	stmt, err := s.DB.Prepare("SELECT file_content FROM file_storage WHERE file_name=$1")  
    if err != nil {
        log.Printf("Failed to prepare SQL statement: %v", err)
        return nil, status.Errorf(codes.Internal, "Internal Server Error")
    }
    defer stmt.Close()

    var fileContent []byte
    err = stmt.QueryRow(in.GetName()).Scan(&fileContent)
    if err != nil {
        log.Printf("Failed to execute SQL statement: %v", err)
        return nil, status.Errorf(codes.NotFound, "File not found")
    }

	log.Printf("Retrieved content: %x", fileContent)
	if fileContent == nil{
		return nil, status.Errorf(codes.NotFound, "File not found")
	}

	leafIndex := s.MerkleTree.GetIndexFromContent(fileContent)
	if leafIndex == -1 {
		return nil, status.Errorf(codes.NotFound, "File not found in Merkle Tree")
	}

	proof, err := s.MerkleTree.GenerateProof(leafIndex)
	if err != nil {
		log.Printf("Error generating Merkle proof: %v", err)
		return nil, status.Errorf(codes.Internal, "Could not generate Merkle proof")
	}
	return &pb.FileDownloadResponse{
		Content:     fileContent,
		MerkleProof: proof,
	}, nil

}
func NewFileTransferServer(db *sql.DB) *FileTransferServer {
	return &FileTransferServer{
		MerkleTree: merkleTree.NewMerkleTree(),
		DB:         db,  // Add this line
	}
}

func main() {
	// Read the port from the environment variables
	port, ok := os.LookupEnv("SERVER_PORT")
	if !ok {
		port = "5000" // default port
	}

	db, err := sql.Open("postgres", "host=db port=5432 user=user password=password dbname=mydb sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
	fileTransferServer := NewFileTransferServer(db)
	pb.RegisterFileTransferServer(grpcServer, fileTransferServer)

	// Start listening
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	log.Printf("Server is listening on port %s...", port)


	// Serve gRPC server
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
