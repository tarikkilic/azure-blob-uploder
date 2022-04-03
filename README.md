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

# Encrypted Version

## Encrypt env variables

If you set env variable(tenant id, secret id, secret client), you should run this command:

azure-uploader.exe encrypt

Program will ask tenant id, client id, secret id and key. Variables are masked, don't worry. You should keep 'key'. You will use to upload files.
After run program, created config.env files. The file used by when upload. It includes tenant id, client id, secret id as encrypted.

## Upload files with env variables

Run this command as usual.

azure-uploader.exe -account <account_name> -container <container_name> -path <directory>

Program will ask your key to decrypt your tenant id, secret id and client id from config.env. Key input is masked.
