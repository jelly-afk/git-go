package git

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

// InitRepo initializes a new Git repository
func InitRepo() error {
	dirs := []string{".git", ".git/objects", ".git/refs"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dir, err)
		}
	}

	headFileContents := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		return fmt.Errorf("error writing HEAD file: %w", err)
	}

	return nil
}

// CloneRepo clones a remote Git repository
func CloneRepo(repoURL string) error {
	refsURL := repoURL + "/info/refs?service=git-upload-pack"
	res, err := http.Get(refsURL)
	if err != nil {
		return fmt.Errorf("error fetching repository: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	lines := bytes.Split(body, []byte("\n"))
	if len(lines) < 2 {
		return fmt.Errorf("invalid response format")
	}

	parts := bytes.SplitN(lines[1], []byte(" "), 2)
	if len(parts) < 2 {
		return fmt.Errorf("invalid reference format")
	}

	want := parts[0][8:]
	readURL := repoURL + "/git-upload-pack"
	bodyContent := fmt.Sprintf("0032want %s\n00000009done\n", want)

	res, err = http.Post(readURL, "application/x-git-upload-pack-request", bytes.NewBufferString(bodyContent))
	if err != nil {
		return fmt.Errorf("error sending POST request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// TODO: Process the response body to extract and write objects
	// This part needs to be implemented based on the Git protocol specification

	return nil
}
