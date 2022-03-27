# azure-blob-uploder

azure-uploader.exe -account <account_name> -container <container_name> -path <directory>

---Azure Blob Storage Uploader---
Usage of C:\Users\tarik\Desktop\Development\azure-blob-uploder\bin\windows\azure-uploader.exe:
-account string
Storage account
-container string
Container name
-path string
Directory of files (default "C:\\Users\\<user>\\upload")

# Run from main.go

go run main.go -account <account_name> -container <container_name> -path <directory>

# Build on windows machine

# For macos

GOOS=darwin GOARCH=amd64 go build -o .\bin\macos\<file_name>

# For windows

go build -o .\bin\windows\<file_name>.exe
