package main

import (
	// "context"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/schollz/progressbar/v3"
)

func main() {
	fmt.Printf("---Azure Blob Storage Uploader---\n")

	ctx := context.Background()

	containerName := flag.String("container", "", "Container name")
	storageAccount := flag.String("account", "", "Storage account")
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		dirname += "\\upload"
	} else {
		dirname += "/upload"
	}

	yourDirectoryPath := flag.String("path", dirname, "Directory of files")
	flag.Parse()
	url := "https://" + *storageAccount + ".blob.core.windows.net/"

	fmt.Printf("Get all files from directory...\n")
	var files []string
	err = filepath.Walk(*yourDirectoryPath, func(path string, info os.FileInfo, err error) error {
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
		total_size := int64(len(data))
		if err != nil {
			log.Fatalf("Failure to read file: %+v", err)
		}
		fileName := file[strings.LastIndex(file, "\\")+1:]
		blobName := fileName

		blobUrl := url + *containerName + "/" + blobName
		blobClient, err := azblob.NewBlockBlobClient(blobUrl, credential, nil)
		if err != nil {
			log.Fatal(err)
		}
		bar := progressbar.Default(100)
		fmt.Println("Starting upload file: " + fileName)
		_, err = blobClient.UploadBufferToBlockBlob(ctx, data, azblob.HighLevelUploadToBlockBlobOption{
			Progress: func(bytesTransferred int64) {
				percByteTransfered := bytesTransferred * 100
				totalPerc := percByteTransfered / int64(total_size)
				bar.Set(int(totalPerc))
			}})
		if err != nil {
			log.Fatalf("Failure to upload to blob: %+v", err)
		}
		fmt.Println("Finished upload file: " + fileName)
		bar.Reset()
	}

}
