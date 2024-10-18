package main

import (
	"fmt"
	pwd "github.com/image357/password"
	"github.com/image357/password/log"
	"os"
	"path"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Println("Usage: recovery <file> <key>")
		os.Exit(1)
	}

	// reconstruct filenames
	var recoveryFile string
	var passwordFile string

	parts := strings.Split(args[0], ".")
	if len(parts) < 2 {
		fmt.Println("Error: non password or recovery key file")
		os.Exit(1)
	}

	if parts[len(parts)-2] == "recovery" {
		passwordFile = strings.Join(parts[:len(parts)-2], ".") + "." + parts[len(parts)-1]
		recoveryFile = args[0]
	} else {
		passwordFile = args[0]
		recoveryFile = strings.Join(parts[:len(parts)-1], ".") + ".recovery." + parts[len(parts)-1]
	}

	passwordFile = strings.ReplaceAll(passwordFile, "\\", "/")
	recoveryFile = strings.ReplaceAll(recoveryFile, "\\", "/")

	passwordFile = path.Clean(passwordFile)
	recoveryFile = path.Clean(recoveryFile)

	println(passwordFile)
	println(recoveryFile)

	// brute force to get storePath and id
	pathParts := strings.Split(passwordFile, "/")
	var key string = ""
	var err error = nil
	for i := 0; i < len(pathParts); i++ {
		storePath := strings.Join(pathParts[:i+1], "/")
		if storePath == "" {
			storePath = "/"
		}

		id := strings.Join(pathParts[i+1:], "/")
		idParts := strings.Split(id, ".")
		id = pwd.NormalizeId(strings.Join(idParts[:len(idParts)-1], "/"))

		fileEnding := idParts[len(idParts)-1]

		if id == "." {
			break
		}

		println(storePath)
		println(id)
		println(fileEnding)

		pwd.SetStorePath(storePath)
		pwd.SetFileEnding(fileEnding)

		key, err = pwd.Get(id+".recovery", args[1])
		if err != nil {
			log.Error(err.Error())
		}

		password, err := pwd.Get(id, key)
		if err != nil {
			log.Error(err.Error())
			return
		}
		println(password)
	}
}
