package main

import (
	"context"
	"log"
	"net"

	merkleTree "go-merkle-file-transfer/merkle"
	pb "go-merkle-file-transfer/protos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileTransferServer struct {
	pb.UnimplementedFileTransferServer
	MerkleTree *merkleTree.MerkleTree
	FileMap    map[string][]byte
}

var merkletree = merkleTree.NewMerkleTree()

func (s *FileTransferServer) UploadFile(ctx context.Context, in *pb.FileData) (*pb.UploadStatus, error) {
	log.Printf("Received UploadFile request for file: %s\n", in.GetName())
	log.Printf("File content: %x", in.GetContent())

	// Storing file content in memory (in a production system, you'd probably write this to disk)
	s.FileMap[in.GetName()] = in.GetContent()
	
	// Update the Merkle tree
	s.MerkleTree.AddFile(in.GetContent())

	return &pb.UploadStatus{Success: true}, nil
}

func (s *FileTransferServer) DownloadFile(ctx context.Context, in *pb.FileName) (*pb.FileDownloadResponse, error) {
	log.Printf("Received DownloadFile request for file: %s\n", in.GetName())

	// Retrieve the file (again, in a production system, you'd probably read this from disk)
	fileContent, exists := s.FileMap[in.GetName()]
	log.Printf("Retrieved content: %x", fileContent)
	if !exists {
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
func NewFileTransferServer() *FileTransferServer {
	return &FileTransferServer{
		FileMap:    make(map[string][]byte),
		MerkleTree: merkleTree.NewMerkleTree(),
	}
}
func main() {
	// Initialize gRPC server
	grpcServer := grpc.NewServer()
	fileTransferServer := NewFileTransferServer()
	pb.RegisterFileTransferServer(grpcServer, fileTransferServer)

	// Start listening
	lis, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatalf("Failed to listen on port 5000: %v", err)
	}

	// Serve gRPC server
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
