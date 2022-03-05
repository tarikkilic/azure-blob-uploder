package main

import (
	// "context"
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/schollz/progressbar/v3"
)

const MAX_SIZE int64 = 268435456000

func getFilesFromDir(yourDirectoryPath string) map[string]string {
	files := make(map[string]string)
	file_counter := 1
	err := filepath.Walk(yourDirectoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if !info.IsDir() {
			if info.Size() > MAX_SIZE {
				log.Fatalf("The file exceeded 250 GB: %s\n", path)
				os.Exit(1)
			}
			files[strconv.Itoa(file_counter)] = path
			file_counter++
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return files
}

func listAllFiles(files map[string]string) {
	keys := make([]string, 0)
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Println("Select files: ")
	for _, k := range keys {
		fmt.Printf("%s: %s\n", k, files[k])
	}
}

func askWhichFiles() []string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	selectedFiles := strings.Split(text, " ")
	// Trimming last \n or \r char element.
	lastElement := selectedFiles[len(selectedFiles)-1]
	lastElement = strings.TrimSuffix(lastElement, "\n")
	lastElement = strings.TrimSuffix(lastElement, "\r")
	lastElement = strings.TrimSuffix(lastElement, "\n")
	selectedFiles[len(selectedFiles)-1] = lastElement
	return selectedFiles
}

func getSelectFilesAndCheck(selectedFiles []string, files map[string]string) []string {
	var selectedFilesArr []string
	for _, elem := range selectedFiles {
		if value, ok := files[elem]; ok {
			selectedFilesArr = append(selectedFilesArr, value)
		} else {
			log.Fatalf("Index number didn't found: %s\n", elem)
			os.Exit(1)
		}
	}
	return selectedFilesArr
}

func setCredentials() *azidentity.ClientSecretCredential {
	credential, err := azidentity.NewClientSecretCredential(os.Getenv("AZURE_TENANT_ID"), os.Getenv("AZURE_CLIENT_ID"), os.Getenv("AZURE_CLIENT_SECRET"), nil)
	if err != nil {
		log.Fatal("Invalid credentials with error: " + err.Error())
		os.Exit(1)
	}
	return credential
}

func uploadFiles(file string, containerUrl string, ctx context.Context, credential *azidentity.ClientSecretCredential) {

	// Upload to data to blob storage
	data, err := os.ReadFile(file)
	total_size := int64(len(data))
	if err != nil {
		log.Fatalf("Failure to read file: %+v", err)
	}
	fileName := file[strings.LastIndex(file, "\\")+1:]
	blobName := fileName

	blobUrl := containerUrl + "/" + blobName
	blobClient, err := azblob.NewBlockBlobClient(blobUrl, credential, nil)
	if err != nil {
		log.Fatal(err)
	}
	bar := progressbar.Default(100)
	fmt.Println("-------------------------------")
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
	fmt.Println("-------------------------------")
	bar.Reset()
}

func main() {
	fmt.Printf("---Azure Blob Storage Uploader---\n")
	ctx := context.Background()

	containerName := flag.String("container", "", "Container name")
	storageAccount := flag.String("account", "", "Storage account")
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	if runtime.GOOS == "windows" {
		dirname += "\\upload"
	} else {
		dirname += "/upload"
	}
	yourDirectoryPath := flag.String("path", dirname, "Directory of files")
	flag.Parse()

	//set your blob azure url
	url := "https://" + *storageAccount + ".blob.core.windows.net/"
	containerUrl := url + *containerName
	fmt.Printf("Get all files from directory...\n")
	files := getFilesFromDir(*yourDirectoryPath)

	//list all files of directory
	listAllFiles(files)

	//ask which files upload
	selectedFiles := askWhichFiles()

	//get file path and check if exist
	selectedFilesArr := getSelectFilesAndCheck(selectedFiles, files)

	//Set credentials
	credential := setCredentials()

	fmt.Println("-------------------------------")
	fmt.Println("Starting upload all files...")
	for _, file := range selectedFilesArr {
		uploadFiles(file, containerUrl, ctx, credential)
	}
	fmt.Println("Finished upload all files.")

}
