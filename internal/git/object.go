package git

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// CreateBlob creates a new blob object from a file
func CreateBlob(filePath string) ([]byte, error) {
	byteContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file content: %w", err)
	}

	blobContent := fmt.Sprintf("blob %d\x00%s", len(byteContent), string(byteContent))
	hashBytes := hashString(blobContent)
	hashHex := fmt.Sprintf("%x", hashBytes)

	if err := writeObject(hashHex, blobContent); err != nil {
		return nil, err
	}

	return hashBytes, nil
}

// CreateTree creates a new tree object from a directory
func CreateTree(dir string) ([]byte, error) {
	dirInfo, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("error getting file stats: %w", err)
	}

	if dirInfo.Name() == ".git" {
		return []byte{}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	var treeContent []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(dir, entry.Name())
			mode := getGitFileMode(dirPath)
			subDirHash, err := CreateTree(dirPath)
			if err != nil {
				return nil, err
			}
			if len(subDirHash) > 0 {
				subDirEntry := fmt.Sprintf("%s %s\x00%s", mode, entry.Name(), subDirHash)
				treeContent = append(treeContent, subDirEntry)
			}
		} else {
			filePath := filepath.Join(dir, entry.Name())
			hashBytes, err := CreateBlob(filePath)
			if err != nil {
				return nil, err
			}
			mode := getGitFileMode(filePath)
			treeEntry := fmt.Sprintf("%s %s\x00%s", mode, entry.Name(), hashBytes)
			treeContent = append(treeContent, treeEntry)
		}
	}

	sortTreeEntries(treeContent)
	treeString := strings.Join(treeContent, "")
	treeHeader := fmt.Sprintf("tree %d\x00", len(treeString))
	treeObjContent := treeHeader + treeString
	hashBytes := hashString(treeObjContent)
	hashHex := fmt.Sprintf("%x", hashBytes)

	if err := writeObject(hashHex, treeObjContent); err != nil {
		return nil, err
	}

	return hashBytes, nil
}

// CreateCommit creates a new commit object
func CreateCommit(treeSha, parentSha, commitMsg string) ([]byte, error) {
	now := time.Now()
	gitTimestamp := fmt.Sprintf("%d %s", now.Unix(), now.Format("-0700"))
	commiterString := "John Doe johndoe@example.com" + " " + gitTimestamp
	commitContent := fmt.Sprintf("tree %s\nparent %s\nauthor %s\ncommitter %s\n\n%s\n",
		treeSha, parentSha, commiterString, commiterString, commitMsg)

	commitHeader := fmt.Sprintf("commit %d\x00", len(commitContent))
	commitObjContent := commitHeader + commitContent
	hashBytes := hashString(commitObjContent)
	hashHex := fmt.Sprintf("%x", hashBytes)

	if err := writeObject(hashHex, commitObjContent); err != nil {
		return nil, err
	}

	return hashBytes, nil
}

// ReadObject reads a Git object from the repository
func ReadObject(sha string) (string, error) {
	filePath := fmt.Sprintf(".git/objects/%s/%s", sha[:2], sha[2:])
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening object file: %w", err)
	}
	defer f.Close()

	r, err := zlib.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("error creating zlib reader: %w", err)
	}

	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		return "", fmt.Errorf("error decompressing object: %w", err)
	}

	dString := out.String()
	parts := strings.SplitN(dString, "\x00", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid object format")
	}

	return parts[1], nil
}

// Helper functions
func hashString(content string) []byte {
	hasher := sha1.New()
	hasher.Write([]byte(content))
	return hasher.Sum(nil)
}

func writeObject(hashHex, content string) error {
	newDirPath := fmt.Sprintf(".git/objects/%s", hashHex[:2])
	newFilePath := fmt.Sprintf("%s/%s", newDirPath, hashHex[2:])

	out := zlibCompress(content)
	if err := os.MkdirAll(newDirPath, 0755); err != nil {
		return fmt.Errorf("error creating object directory: %w", err)
	}

	if err := os.WriteFile(newFilePath, out.Bytes(), 0644); err != nil {
		return fmt.Errorf("error writing object file: %w", err)
	}

	return nil
}

func zlibCompress(content string) bytes.Buffer {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	if _, err := w.Write([]byte(content)); err != nil {
		panic(fmt.Sprintf("error compressing content: %v", err))
	}
	if err := w.Close(); err != nil {
		panic(fmt.Sprintf("error closing zlib writer: %v", err))
	}
	return b
}

func getGitFileMode(path string) string {
	info, err := os.Lstat(path)
	if err != nil {
		panic(fmt.Sprintf("error getting file stats: %v", err))
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
		jName := strings.Split(entries[j], " ")[1]
		return iName > jName
	})
}
