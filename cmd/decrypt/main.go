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
		fmt.Println("Usage: decrypt <file> <key>")
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
	encryptedData := string(fileContents)

	// decrypt data
	packedData, err := password.Decrypt(encryptedData, storageKey)
	if err != nil {
		fmt.Println("Error decrypting file:", err)
		os.Exit(1)
	}

	// print data
	fmt.Printf("%v", packedData)
}
