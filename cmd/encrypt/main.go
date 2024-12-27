package main

import (
	"fmt"
	"github.com/image357/password"
	"os"
	"unicode/utf8"
)

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Println("Usage: encrypt <file> <key>")
		os.Exit(1)
	}

	filePath := args[0]
	storageKey := args[1]

	// read file
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}
	if !utf8.Valid(fileContents) {
		fmt.Println("invalid utf8 character after file reading")
		os.Exit(1)
	}
	data := string(fileContents)

	// encrypt data
	encryptedData, err := password.Encrypt(data, storageKey)
	if err != nil {
		fmt.Println("Error encrypting file:", err)
		os.Exit(1)
	}

	// print data
	fmt.Printf("%v", encryptedData)
}
