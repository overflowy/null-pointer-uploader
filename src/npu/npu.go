package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rodaine/table"
	"golang.design/x/clipboard"
)

type uploadedFile struct {
	fileName string
	upUrl    string
}

func main() {
	expires := flag.Int("expires", 0, "set expiration time for the uploaded file in hours")
	secret := flag.Bool("secret", false, "enable secret mode")
	dirMode := flag.Bool("dir", false, "enable directory mode (non-recursive)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: [-expires <hours>] [-secret] [-dir] <directory/filePath>\n")
		flag.PrintDefaults()
		fmt.Print("\n")
		os.Exit(1)
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
	}

	// construct list of files to upload
	filePaths := []string{}
	dir := "./"
	if *dirMode {
		dir = strings.TrimSuffix(strings.TrimSuffix(args[0], "/"), "\\") // trim trailing slashes

		dirEntries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Println("Failed to read files in directory:", err)
			os.Exit(1)
		}

		for _, entry := range dirEntries {
			if !entry.IsDir() {
				filePaths = append(filePaths, fmt.Sprintf("%s/%s", dir, entry.Name()))
			}
		}
	} else {
		for _, argPath := range args {
			filePaths = append(filePaths, argPath)
		}
	}

	// upload files
	uploadedFiles := []uploadedFile{}
	for _, filePath := range filePaths {
		responseBody, err := uploadFile(filePath, expires, secret)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if responseBody != nil {
			upUrl := string(responseBody)
			upUrl = strings.TrimSuffix(strings.TrimSuffix(upUrl, "\n"), "\r")

			uploadedFiles = append(uploadedFiles, uploadedFile{
				fileName: filepath.Base(filePath),
				upUrl:    upUrl,
			})
		}
	}

	if len(uploadedFiles) == 1 {
		// original, single file upload
		upUrl := uploadedFiles[0].upUrl + "\n"
		fmt.Println(string(upUrl))
		clipboard.Write(clipboard.FmtText, []byte(upUrl))
	} else {
		// for multiple file upload
		printTblOutput(&uploadedFiles)
		logFile, err := writeOutputToFile(dir, &uploadedFiles)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("\nLog file saved:", logFile)
	}
}

func printTblOutput(uploadedFiles *[]uploadedFile) {
	tbl := table.New("File", "| 0x0 Url")
	for _, upFile := range *uploadedFiles {
		tbl.AddRow(upFile.fileName, "| "+upFile.upUrl)
	}
	tbl.Print()
}

func writeOutputToFile(dir string, uploadedFiles *[]uploadedFile) (logFile string, err error) {
	logFile = fmt.Sprintf("%s/%s.%d.log", dir, "upload", time.Now().Unix())
	f, err := os.Create(logFile)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to save output log file: %v", err))
	}
	defer f.Close()

	for _, upFile := range *uploadedFiles {
		_, err = f.WriteString(fmt.Sprintf("%s: %s\r\n", upFile.fileName, upFile.upUrl))
		if err != nil {
			return "", errors.New(fmt.Sprintf("Failed to write to log file: %v", err))
		}
	}

	f.Sync()

	return logFile, nil
}

func uploadFile(filePath string, expires *int, secret *bool) (responseBody []byte, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to open file: %v", err))
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to create form file: %v", err))
	}
	io.Copy(part, file)

	if *expires > 0 {
		expiration := time.Now().Add(time.Duration(*expires) * time.Hour)
		fmt.Println("Expires:", expiration.Format(time.RFC3339))
		fmt.Println()
		writer.WriteField("expires", strconv.Itoa(*expires))
	}

	if *secret {
		writer.WriteField("secret", "true")
	}

	err = writer.Close()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to close multipart writer: %v", err))
	}

	request, err := http.NewRequest("POST", "https://0x0.st", body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to create HTTP request: %v", err))
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to send HTTP request: %v", err))
	}
	defer response.Body.Close()

	responseBody, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to read response body: %v", err))
	}

	return responseBody, nil
}
