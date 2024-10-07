package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var removalString = `#ifdef _MSC_VER
#include <complex.h>
typedef _Fcomplex GoComplex64;
typedef _Dcomplex GoComplex128;
#else
typedef float _Complex GoComplex64;
typedef double _Complex GoComplex128;
#endif
`

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Usage: patchheader <file>")
		os.Exit(1)
	}

	file, err := os.OpenFile(args[0], os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}

	b, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}
	content := string(b)

	// remove strings
	content = strings.Replace(content, removalString, "", 1)

	// write file
	err = file.Truncate(0)
	if err != nil {
		fmt.Printf("Error truncating file: %v\n", err)
		os.Exit(1)
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		fmt.Printf("Error seeking file: %v\n", err)
		os.Exit(1)
	}
	_, err = file.WriteString(content)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		os.Exit(1)
	}
	err = file.Close()
	if err != nil {
		fmt.Printf("Error closing file: %v\n", err)
		os.Exit(1)
	}
}
