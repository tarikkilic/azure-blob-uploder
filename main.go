package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func main() {
	fmt.Printf("Azure Blob storage quick start sample\n")
	url := "https://jasonwhiteupwork.blob.core.windows.net/"
	ctx := context.Background()
	tid := os.Getenv("AZURE_TENANT_ID")
	fmt.Println(os.Getenv(tid))
	credential, err := azidentity.NewClientSecretCredential(os.Getenv("AZURE_TENANT_ID"), os.Getenv("AZURE_CLIENT_ID"), os.Getenv("AZURE_CLIENT_SECRET"), nil)

	if err != nil {
		log.Fatal("Invalid credentials with error: " + err.Error())
	}

	fmt.Printf("Creating a dummy file to test the upload and download\n")
	//serviceClient, err := azblob.NewServiceClient(url, credential, nil)
	data := []byte("\nhello world this is a blob\n")
	blobName := "jasonwhite"
	containerName := "jasonwhite"
	blobUrl := url + containerName + "/" + blobName
	blobClient, err := azblob.NewBlockBlobClient(blobUrl, credential, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created blob client")
	// Upload to data to blob storage
	_, err = blobClient.UploadBufferToBlockBlob(ctx, data, azblob.HighLevelUploadToBlockBlobOption{})
	if err != nil {
		log.Fatalf("Failure to upload to blob: %+v", err)
	}
}
