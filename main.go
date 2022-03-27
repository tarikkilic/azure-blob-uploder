package main

import (
	// "context"
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/joho/godotenv"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
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

func setCredentials(key string) *azidentity.ClientSecretCredential {
	var myEnv map[string]string
	myEnv, err := godotenv.Read("config.env")
	if err != nil {
		log.Fatalf("Failure to read config: %+v", err)
		os.Exit(1)
	}
	tenantIdE := myEnv["tenantId"]
	clientIdE := myEnv["clientId"]
	clientSecretE := myEnv["clientSecret"]

	tenantId := decrypt([]byte(tenantIdE), key)
	clientId := decrypt([]byte(clientIdE), key)
	clientSecret := decrypt([]byte(clientSecretE), key)

	// credential, err := azidentity.NewClientSecretCredential(os.Getenv("AZURE_TENANT_ID"), os.Getenv("AZURE_CLIENT_ID"), os.Getenv("AZURE_CLIENT_SECRET"), nil)
	credential, err := azidentity.NewClientSecretCredential(string(tenantId), string(clientId), string(clientSecret), nil)
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

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func decrypt(data []byte, passphrase string) []byte {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

func main() {
	fmt.Printf("---Azure Blob Storage Uploader---\n")
	ctx := context.Background()

	if os.Args[1] == "encrypt" {
		err := godotenv.Load("config.env")
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		fmt.Print("Enter tenant id: ")
		tenantId, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		fmt.Println()

		fmt.Print("Enter client id: ")
		clientId, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		fmt.Println()

		fmt.Print("Enter secret id: ")
		clientSecret, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		fmt.Println()

		fmt.Print("Enter your key: ")
		key, err := term.ReadPassword(int(syscall.Stdin))

		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		fmt.Println()
		defer os.Stdin.Close()
		cipherTenantId := encrypt(tenantId, string(key))
		cipherClientId := encrypt(clientId, string(key))
		cipherClientSecret := encrypt(clientSecret, string(key))

		myMap := make(map[string]string)
		myMap["tenantId"] = string(cipherTenantId)
		myMap["clientId"] = string(cipherClientId)
		myMap["clientSecret"] = string(cipherClientSecret)
		err = godotenv.Write(myMap, "config.env")
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// tenantID := encryptCmd.String("tenant_id", "", "AZURE_TENANT_ID")
	// clientID := encryptCmd.String("tenant_id", "", "AZURE_TENANT_ID")
	// clientSecret := encryptCmd.String("tenant_id", "", "AZURE_TENANT_ID")
	uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)

	containerName := uploadCmd.String("container", "", "Container name")
	storageAccount := uploadCmd.String("account", "", "Storage account")
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
	yourDirectoryPath := uploadCmd.String("path", dirname, "Directory of files")
	uploadCmd.Parse(os.Args[2:])

	//set your blob azure url
	url := "https://" + *storageAccount + ".blob.core.windows.net/"
	containerUrl := url + *containerName

	fmt.Print("Enter your key: ")
	key, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	fmt.Println()

	fmt.Printf("Get all files from directory...\n")
	files := getFilesFromDir(*yourDirectoryPath)

	//list all files of directory
	listAllFiles(files)

	//ask which files upload
	selectedFiles := askWhichFiles()

	//get file path and check if exist
	selectedFilesArr := getSelectFilesAndCheck(selectedFiles, files)

	//Set credentials
	credential := setCredentials(string(key))

	fmt.Println("-------------------------------")
	fmt.Println("Starting upload all files...")
	for _, file := range selectedFilesArr {
		uploadFiles(file, containerUrl, ctx, credential)
	}
	fmt.Println("Finished upload all files.")

}
