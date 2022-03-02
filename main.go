package main

import (
	// "context"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func main() {
	fmt.Printf("---Azure Blob Storage Uploader---\n")
	url := "https://jasonwhiteupwork.blob.core.windows.net/"
	ctx := context.Background()

	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dirname += "\\upload"
	fmt.Printf("Get all files from directory...\n")
	var files []string
	err = filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	credential, err := azidentity.NewClientSecretCredential(os.Getenv("AZURE_TENANT_ID"), os.Getenv("AZURE_CLIENT_ID"), os.Getenv("AZURE_CLIENT_SECRET"), nil)

	if err != nil {
		log.Fatal("Invalid credentials with error: " + err.Error())
	}

	fmt.Println("Starting upload files...")
	for _, file := range files {
		// Upload to data to blob storage
		data, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failure to read file: %+v", err)
		}
		fileName := file[strings.LastIndex(file, "\\")+1:]
		blobName := fileName
		containerName := "jasonwhite"
		blobUrl := url + containerName + "/" + blobName
		blobClient, err := azblob.NewBlockBlobClient(blobUrl, credential, nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Starting upload file: " + fileName)
		_, err = blobClient.UploadBufferToBlockBlob(ctx, data, azblob.HighLevelUploadToBlockBlobOption{})
		if err != nil {
			log.Fatalf("Failure to upload to blob: %+v", err)
		}
		fmt.Println("Finished upload file: " + fileName)
	}

}
