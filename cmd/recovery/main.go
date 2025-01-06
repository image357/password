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

	if filepath.Ext(abs) != ("." + pwd.DefaultFileEnding) {
		fmt.Println("Error: <file> must end with \"." + pwd.DefaultFileEnding + "\"")
		os.Exit(1)
	}

	abs = strings.ReplaceAll(abs, "\\", "/")
	dir, file := path.Split(abs)
	recoveryKey := args[1]

	// reconstruct filenames
	var recoveryFile string
	var passwordFile string

	suffix := filepath.Ext(strings.TrimSuffix(file, "."+pwd.DefaultFileEnding))
	if suffix == pwd.RecoveryIdSuffix {
		passwordFile = strings.TrimSuffix(strings.TrimSuffix(file, "."+pwd.DefaultFileEnding), pwd.RecoveryIdSuffix) + "." + pwd.DefaultFileEnding
		recoveryFile = file
	} else {
		passwordFile = file
		recoveryFile = strings.TrimSuffix(file, "."+pwd.DefaultFileEnding) + pwd.RecoveryIdSuffix + "." + pwd.DefaultFileEnding
	}

	// brute force to get storePath and id
	var storePath = ""
	var id = ""
	var storageKey = ""
	var password = ""
	err = nil

	pathParts := strings.Split(dir, "/")
	for i := 0; i < len(pathParts)-1; i++ {
		storePath = path.Clean(strings.Join(pathParts[:i+1], "/") + "/")

		id = strings.Join(pathParts[i+1:], "/")
		id = id + "/" + passwordFile
		id = strings.TrimSuffix(id, "."+pwd.DefaultFileEnding)
		id = pwd.NormalizeId(id)

		err = pwd.SetStorePath(storePath)
		if err != nil {
			continue
		}

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
