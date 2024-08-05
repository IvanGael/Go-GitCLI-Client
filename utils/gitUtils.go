package utils

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"git-client/models"
	"os"
	"path/filepath"
)

// NewGitRepo initializes a new Git repository
func NewGitRepo(path string) (*models.GitRepo, error) {
	repoPath := filepath.Join(path, ".git")
	err := os.MkdirAll(repoPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &models.GitRepo{Path: repoPath}, nil
}

// DecodeObject decodes the object content stored in the Git repository
// In a real Git implementation, you would handle various object types (blob, tree, commit, etc.)
func DecodeObject(content []byte) (string, error) {
	// Here, we assume that the content is already decoded (e.g., a commit object)
	return string(content), nil
}

// decodeObject decodes the object content stored in the Git repository
// func decodeObject(content []byte) (string, error) {
// 	// Split the object content into its type and data parts
// 	parts := bytes.SplitN(content, []byte("\n"), 2)
// 	if len(parts) != 2 {
// 		return "", errors.New("invalid object format")
// 	}
// 	objectType := string(parts[0])
// 	data := parts[1]

// 	switch objectType {
// 	case "blob":
// 		// For blob objects, return the data as is
// 		return string(data), nil
// 	case "tree":
// 		// For tree objects, parse and format the tree structure
// 		return parseTree(data)
// 	case "commit ":
// 		// For commit objects, return the commit information
// 		return string(data), nil
// 	default:
// 		return "", fmt.Errorf("unsupported object type: %s", objectType)
// 	}
// }

// ParseTree parses and formats the tree object
func ParseTree(data []byte) (string, error) {
	var treeOutput string
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := bytes.SplitN(line, []byte(" "), 2)
		if len(parts) != 2 {
			return "", errors.New("invalid tree entry format")
		}
		mode := string(parts[0])
		name := parts[1]
		treeOutput += fmt.Sprintf("%s %s\n", mode, name)
	}
	return treeOutput, nil
}

// ListFiles returns a list of files in a directory
func ListFiles(dirPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relativePath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}
			files = append(files, relativePath)
		}
		return nil
	})

	return files, err
}

// Contains checks if a slice contains a specific string
func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// HashObject calculates the SHA-1 hash of an object
func HashObject(content string) string {
	hash := sha1.New()
	hash.Write([]byte(content))
	return hex.EncodeToString(hash.Sum(nil))
}

// HashFile Calculates the SHA-1 hash of a file
func HashFile(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	hash := sha1.New()
	hash.Write(content)
	return hex.EncodeToString(hash.Sum(nil))
}

// IsExcluded returns excludedFiles
func IsExcluded(file string) bool {
	excludedFiles := map[string]bool{
		"HEAD":        true,
		"config":      true,
		"config.json": true,
		"description": true,
		"index":       true,
		"objects":     true,
	}
	return excludedFiles[file]
}
