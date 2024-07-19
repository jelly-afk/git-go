package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"strings"
)

// Usage: your_program.sh <command> <arg1> <arg2> ...
func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}
	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}
		headFileContents := []byte("ref: refs/heads/main\n")
		err := os.WriteFile(".git/HEAD", headFileContents, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}
		fmt.Println("Initialized git directory")
	case "cat-file":
		blob_sha := os.Args[3]
		filePath := fmt.Sprintf("./.git/objects/%s/%s", blob_sha[:2], blob_sha[2:])
		
		f, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening blob file: %s\n", err)
		}
		defer f.Close()
		r, err := zlib.NewReader(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while reading through zlib: %s\n", err)
		}
		
		var out bytes.Buffer
    	if _, err := io.Copy(&out, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error while decompressing: %s\n", err)
		}
		dString := out.String()
		res := strings.SplitN(dString, "\x00", 2)[1]
		fmt.Print(res)
		

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
