package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
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
	case "hash-object":
		filePath := os.Args[3]
		byteContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file's content : %s\n", err)
		}
		strContent := string(byteContent)
		blobFileContent := fmt.Sprintf("blob %d\x00%s", len(strContent), strContent)
		hasher := sha1.New()
		hasher.Write([]byte(blobFileContent))
		hashBytes := hasher.Sum(nil)
		hashString := hex.EncodeToString(hashBytes)
		fmt.Print(hashString)
		newDirPath := fmt.Sprintf(".git/objects/%s", hashString[:2])
		newFilePath := fmt.Sprintf(".git/objects/%s/%s", hashString[:2], hashString[2:])
		// fmt.Print(newFilePath)
		var b bytes.Buffer
		w := zlib.NewWriter(&b)
		_, err = w.Write([]byte(blobFileContent))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error compressing file content: %s\n", err)

		}
		err = w.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error closing zlib writer: %s\n", err)

		}
		err = os.MkdirAll(newDirPath, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while making new directory: %s\n", err)

		}
		err = os.WriteFile(newFilePath, b.Bytes(), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file's content from me: %s\n", err)
		}
	case "ls-tree":
		treeHash := os.Args[3]
		filePath := fmt.Sprintf(".git/objects/%s/%s", treeHash[:2], treeHash[2:])
		f, _ := os.Open(filePath)
		r, _ := zlib.NewReader(f)
		con, _ := io.ReadAll(r)
		split := bytes.Split(con, []byte("\x00"))
		use := split[1 : len(split)-1]
		for _, a := range use {
			b := bytes.Split(a, []byte(" "))
			if len(b) > 1 {
				fmt.Println(string(b[1]))
			}

		}
		// for _,arr := range split {
		// 	x := bytes.Split(arr, []byte(" "))
		// 	for _,i := range x {
		// 		println(string(i))
		// 	}
		// }

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
