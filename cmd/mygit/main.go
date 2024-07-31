package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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
		hashByte := createBlob(filePath)
		fmt.Printf("%x", hashByte)

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
	case "write-tree":
		dirPath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get working directory %s\n", command)
		}
		hashBytes := generateTreesAndBlobs(dirPath)
		hashString := fmt.Sprintf("%x", hashBytes)
		fmt.Print(hashString)
	case "commit-tree":
		treeSha := os.Args[2]
		pCommitSha := os.Args[4]
		commitMsg := os.Args[6]
		now := time.Now()
		gitTimestamp := fmt.Sprintf("%d %s", now.Unix(), now.Format("-0700"))
		commiterString := "John Doe johndoe@example.com" + " " + gitTimestamp
		commitContent := fmt.Sprintf("tree %s\nparent %s\nauthor %s\ncommitter %s\n\n%s\n", treeSha, pCommitSha, commiterString, commiterString, commitMsg)

		commitHeader := fmt.Sprintf("commit %d\x00", len(commitContent))
		commitObjContent := commitHeader + commitContent
		hashBytes := hashString(commitObjContent)
		hashHex := fmt.Sprintf("%x", hashBytes)
		newDirPath := fmt.Sprintf(".git/objects/%s", hashHex[:2])
		newFilePath := fmt.Sprintf("%s/%s", newDirPath, hashHex[2:])

		out := zlibCompress(commitObjContent)
		err := os.MkdirAll(newDirPath, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while making new directory: %s\n", err)

		}
		err = os.WriteFile(newFilePath, out.Bytes(), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file's content from me: %s\n", err)
		}
		fmt.Print(hashHex)
	case "clone":
		repo_url := os.Args[2]
		res, err := http.Get(repo_url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while fetching repo: %s\n", err)
		}
		println(res)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)

		os.Exit(1)
	}
}

func generateTreesAndBlobs(dir string) []byte {
	dirInfo, err := os.Stat(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting file stats: %s\n", err)
	}
	if dirInfo.Name() == ".git" {
		return []byte{}
	}

	entries, err := os.ReadDir(dir)
	var treeContent []string
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %s\n", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(dir, entry.Name())
			mode := getGitFileMode(dirPath)
			subDirHash := generateTreesAndBlobs(entry.Name())
			if len(subDirHash) > 0 {
				subDirEntry := fmt.Sprintf("%s %s\x00%s", mode, entry.Name(), subDirHash)
				treeContent = append(treeContent, subDirEntry)
			}

		} else {
			filePath := filepath.Join(dir, entry.Name())
			hashBytes := createBlob(filePath)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting file stats: %s\n", err)
			}
			mode := getGitFileMode(filePath)
			treeEntry := fmt.Sprintf("%s %s\x00%s", mode, entry.Name(), hashBytes)
			treeContent = append(treeContent, treeEntry)
		}
	}
	sortTreeEntries(treeContent)
	treeString := strings.Join(treeContent, "")
	treeHeader := fmt.Sprintf("tree %d\x00", len(treeString))
	treeFileContent := treeHeader + treeString
	hashBytes := hashString(treeFileContent)
	hashHex := fmt.Sprintf("%x", hashBytes)
	newDirPath := fmt.Sprintf(".git/objects/%s", hashHex[:2])
	newFilePath := fmt.Sprintf(".git/objects/%s/%s", hashHex[:2], hashHex[2:])

	out := zlibCompress(treeFileContent)
	err = os.MkdirAll(newDirPath, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while making new directory: %s\n", err)

	}
	err = os.WriteFile(newFilePath, out.Bytes(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file's content from me: %s\n", err)
	}
	return hashBytes

}

func createBlob(filePath string) []byte {
	byteContent, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file's content : %s\n", err)
	}
	strContent := string(byteContent)
	blobFileContent := fmt.Sprintf("blob %d\x00%s", len(strContent), strContent)
	hashBytes := hashString(blobFileContent)
	hashhex := fmt.Sprintf("%x", hashBytes)
	newDirPath := fmt.Sprintf(".git/objects/%s", hashhex[:2])
	newFilePath := fmt.Sprintf(".git/objects/%s/%s", hashhex[:2], hashhex[2:])

	out := zlibCompress(blobFileContent)
	err = os.MkdirAll(newDirPath, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while making new directory: %s\n", err)

	}
	err = os.WriteFile(newFilePath, out.Bytes(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file's content from me: %s\n", err)
	}
	return hashBytes
}

func hashString(content string) []byte {
	hasher := sha1.New()
	hasher.Write([]byte(content))
	hashBytes := hasher.Sum(nil)
	return hashBytes
}

func zlibCompress(content string) bytes.Buffer {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	_, err := w.Write([]byte(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error compressing file content: %s\n", err)

	}
	err = w.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error closing zlib writer: %s\n", err)

	}
	return b
}

func getGitFileMode(path string) string {
	info, err := os.Lstat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting file,s stat: %s\n", err)

	}

	mode := info.Mode()

	switch {
	case mode.IsDir():
		return "40000"
	case mode&os.ModeSymlink != 0:
		return "120000"
	default:
		if mode&0111 != 0 {
			return "100755"
		}
		return "100644"
	}
}

func sortTreeEntries(entries []string) {
	sort.Slice(entries, func(i, j int) bool {
		iName := strings.Split(entries[i], " ")[1]
		jName := strings.Split(entries[i], " ")[1]
		return iName > jName

	})
}
