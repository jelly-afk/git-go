package main

import (
	"fmt"
	"git-go/internal/git"
	"os"
	"strings"
)

// Usage: run.sh <command> <arg1> <arg2> ...
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		if err := git.InitRepo(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing repository: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("Initialized git directory")

	case "cat-file":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "usage: mygit cat-file -p <sha>\n")
			os.Exit(1)
		}
		blobSha := os.Args[3]
		content, err := git.ReadObject(blobSha)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading object: %s\n", err)
			os.Exit(1)
		}
		fmt.Print(content)

	case "hash-object":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "usage: mygit hash-object -w <file>\n")
			os.Exit(1)
		}
		filePath := os.Args[3]
		hashBytes, err := git.CreateBlob(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating blob: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("%x", hashBytes)

	case "ls-tree":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "usage: mygit ls-tree --name-only <sha>\n")
			os.Exit(1)
		}
		treeHash := os.Args[3]
		content, err := git.ReadObject(treeHash)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading tree: %s\n", err)
			os.Exit(1)
		}
		// Parse and print only the names from the tree content
		entries := strings.Split(content, "\x00")
		for _, entry := range entries {
			if entry == "" {
				continue
			}
			parts := strings.Split(entry, " ")
			if len(parts) > 1 {
				fmt.Println(parts[1])
			}
		}

	case "write-tree":
		dirPath, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting working directory: %s\n", err)
			os.Exit(1)
		}
		hashBytes, err := git.CreateTree(dirPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating tree: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("%x", hashBytes)

	case "commit-tree":
		if len(os.Args) < 7 {
			fmt.Fprintf(os.Stderr, "usage: mygit commit-tree <tree-sha> -p <parent-sha> -m <message>\n")
			os.Exit(1)
		}
		treeSha := os.Args[2]
		pCommitSha := os.Args[4]
		commitMsg := os.Args[6]
		hashBytes, err := git.CreateCommit(treeSha, pCommitSha, commitMsg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating commit: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("%x", hashBytes)

	case "clone":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "usage: mygit clone <repository-url>\n")
			os.Exit(1)
		}
		repoURL := os.Args[2]
		if err := git.CloneRepo(repoURL); err != nil {
			fmt.Fprintf(os.Stderr, "Error cloning repository: %s\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
