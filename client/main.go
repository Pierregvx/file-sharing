package main

import (
	"context"
	"log"

	pb "go-merkle-file-transfer/protos/lib"

	"google.golang.org/grpc"
)

func uploadFile(client pb.FileTransferClient, fileName string, content []byte) error {
	log.Printf("Uploading file: %s\n", fileName)
	_, err := client.UploadFile(context.Background(), &pb.FileData{Name: fileName, Content: content})
	if err != nil {
		return err
	}
	return nil
}

func downloadFile(client pb.FileTransferClient, fileName string) ([]byte, error) {
	response, err := client.DownloadFile(context.Background(), &pb.FileName{Name: fileName})
	if err != nil {
		return nil, err
	}

	return response.Content, nil
}


func main() {
	conn, err := grpc.Dial(":5000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileTransferClient(conn)

	err = uploadFile(client, "example.txt", []byte("This is an example"))
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}

	content, err := downloadFile(client, "example.txt")
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	log.Printf("Downloaded content: %s", string(content))

}
