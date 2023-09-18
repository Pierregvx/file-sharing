package main

import (
	"context"
	"log"

	pb "go-merkle-file-transfer/protos/lib"

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


