package main

import (
	"fmt"
	pwd "github.com/image357/password"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Println("Usage: recovery <file> <key>")
		os.Exit(1)
	}

	// basic operation on args
	abs, err := filepath.Abs(args[0])
	if err != nil {
		fmt.Println("Error: <file> is not a proper path")
		os.Exit(1)
	}
	_, err = os.Stat(abs)
	if err != nil {
		fmt.Println("Error: <file> does not exist")
		os.Exit(1)
	}

	abs = strings.ReplaceAll(abs, "\\", "/")
	dir, file := path.Split(abs)
	recoveryKey := args[1]

	// reconstruct filenames
	var recoveryFile string
	var passwordFile string

	fileParts := strings.Split(file, ".")
	if len(fileParts) < 2 {
		fmt.Println("Error: <file> is not a password or recovery key file")
		os.Exit(1)
	}

	if fileParts[len(fileParts)-2] == pwd.RecoveryIdSuffix[1:] && len(fileParts) >= 3 {
		passwordFile = strings.Join(fileParts[:len(fileParts)-2], ".") + "." + fileParts[len(fileParts)-1]
		recoveryFile = file
	} else {
		passwordFile = file
		recoveryFile = strings.Join(fileParts[:len(fileParts)-1], ".") + pwd.RecoveryIdSuffix + "." + fileParts[len(fileParts)-1]
	}

	// brute force to get storePath and id
	var storePath string = ""
	var id string = ""
	var storageKey string = ""
	var password string = ""
	err = nil

	pathParts := strings.Split(dir, "/")
	for i := 0; i < len(pathParts)-1; i++ {
		storePath = path.Clean(strings.Join(pathParts[:i+1], "/") + "/")

		id = strings.Join(pathParts[i+1:], "/")
		idParts := strings.Split(passwordFile, ".")
		id = pwd.NormalizeId(path.Join(id, strings.Join(idParts[:len(idParts)-1], ".")))

		fileEnding := idParts[len(idParts)-1]

		pwd.SetStorePath(storePath)
		pwd.SetFileEnding(fileEnding)

		recoveryId := id + pwd.RecoveryIdSuffix
		storageKey, err = pwd.Get(recoveryId, recoveryKey)
		if err != nil {
			continue
		}

		password, err = pwd.Get(id, storageKey)
		if err == nil {
			break
		}
	}

	// print failure
	if err != nil {
		fmt.Println("Error: cannot recover password")
		os.Exit(1)
	}

	// print success
	fmt.Printf("storage path:     %v\n", filepath.FromSlash(storePath))
	fmt.Printf("password file:    %v\n", passwordFile)
	fmt.Printf("recovery file:    %v\n", recoveryFile)
	fmt.Printf("storage key:      %v\n", storageKey)
	fmt.Printf("recovery key:     %v\n", recoveryKey)
	fmt.Printf("id:               %v\n", id)
	fmt.Printf("password:         %v\n", password)
}
