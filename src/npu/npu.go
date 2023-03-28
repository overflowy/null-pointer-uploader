package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"golang.design/x/clipboard"
)

func main() {
	expires := flag.Int("expires", 0, "set expiration time for the uploaded file in hours")
	secret := flag.Bool("secret", false, "enable secret mode")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: [-expires <hours>] [-secret] <filePath>\n")
		flag.PrintDefaults()
		fmt.Print("\n")
		os.Exit(1)
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
	}
	filePath := args[0]

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Failed to open file:", err)
		os.Exit(1)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		fmt.Println("Failed to create form file:", err)
		os.Exit(1)
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
		fmt.Println("Failed to close multipart writer:", err)
		os.Exit(1)
	}

	request, err := http.NewRequest("POST", "https://0x0.st", body)
	if err != nil {
		fmt.Println("Failed to create HTTP request:", err)
		os.Exit(1)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Failed to send HTTP request:", err)
		os.Exit(1)
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Failed to read response body:", err)
		os.Exit(1)
	}

	if responseBody != nil {
		fmt.Println(string(responseBody))
		clipboard.Write(clipboard.FmtText, []byte(string(responseBody)))
	}
}
