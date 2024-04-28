package models

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git-client/utils"
)

type GitRepo struct {
	Path     string
	Username string
	Email    string
}

type Config struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Init Initializes an empty repository
func (repo *GitRepo) Init() error {
	subdirs := []string{"objects", "refs/heads", "refs/tags"}
	for _, subdir := range subdirs {
		err := os.MkdirAll(filepath.Join(repo.Path, subdir), os.ModePerm)
		if err != nil {
			return err
		}
	}

	defaultFiles := map[string]string{
		"config":      "[core]\n\trepositoryformatversion = 0\n\tfilemode = true\n\tbare = false\n\tlogallrefupdates = true\n",
		"description": "Unnamed repository; edit this file 'description' to name the repository.\n",
		"HEAD":        "ref: refs/heads/master\n",
		"config.json": "",
	}

	for filename, content := range defaultFiles {
		err := os.WriteFile(filepath.Join(repo.Path, filename), []byte(content), os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetConfig Set user username and email
func (repo *GitRepo) SetConfig(username, email string) error {
	repo.Username = username
	repo.Email = email
	config := Config{Username: username, Email: email}
	configPath := filepath.Join(repo.Path, "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// LoadConfig Load user config
func (repo *GitRepo) LoadConfig() error {
	configPath := filepath.Join(repo.Path, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}
	repo.Username = config.Username
	repo.Email = config.Email
	return nil
}

// Add adds a file to the index
func (repo *GitRepo) Add(filePath string) error {
	err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// Only add regular files to the index
			fileContent, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			hash := sha1.New()
			hash.Write(fileContent)
			fileHash := hex.EncodeToString(hash.Sum(nil))

			indexPath := filepath.Join(repo.Path, "index")
			// Append the file path and its hash to the index file
			err = os.WriteFile(indexPath, []byte(fmt.Sprintf("%s %s\n", fileHash, path)), os.ModePerm)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Status Displays the repository status
func (repo *GitRepo) Status() (string, error) {
	// Get the list of files in the working directory
	workingDirFiles := make(map[string]string)
	err := filepath.Walk(repo.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relativePath, err := filepath.Rel(repo.Path, path)
			if err != nil {
				return err
			}
			if !utils.IsExcluded(relativePath) {
				workingDirFiles[relativePath] = utils.HashFile(path)
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	// Get the list of files in the index
	indexPath := filepath.Join(repo.Path, "index")
	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If index does not exist, return message indicating no commits yet
			return "No commits yet\nnothing to commit (create/copy files and use \"git add\" to track)\n", nil
		}
		return "", err
	}
	indexFiles := strings.Split(strings.TrimSpace(string(indexContent)), "\n")

	// Map to store hashes of files in the index
	indexHashes := make(map[string]string)
	for _, indexFile := range indexFiles {
		indexHashes[indexFile] = utils.HashFile(filepath.Join(repo.Path, indexFile))
	}

	var changesToBeCommitted []string
	var untrackedFiles []string

	// Compare files in the working directory with the index
	for file, workingDirHash := range workingDirFiles {
		indexHash, exists := indexHashes[file]
		if exists {
			// File exists in the index
			if workingDirHash != indexHash {
				// File has been modified
				changesToBeCommitted = append(changesToBeCommitted, fmt.Sprintf("\tmodified:   %s", file))
			}
			// Remove the file from indexHashes since it's accounted for
			delete(indexHashes, file)
		} else {
			// File is not in the index, it's untracked
			untrackedFiles = append(untrackedFiles, fmt.Sprintf("\t%s", file))
		}
	}

	// Check for files in the index not found in the working directory (deleted files)
	for indexFile := range indexHashes {
		changesToBeCommitted = append(changesToBeCommitted, fmt.Sprintf("\tdeleted:   %s", indexFile))
	}

	// Prepare the status message
	statusMessage := "On branch master\n\n"
	if len(changesToBeCommitted) > 0 {
		statusMessage += "Changes to be committed:\n"
		for _, change := range changesToBeCommitted {
			statusMessage += change + "\n"
		}
	}

	if len(untrackedFiles) > 0 {
		statusMessage += "\nUntracked files:\n"
		for _, untrackedFile := range untrackedFiles {
			statusMessage += untrackedFile + "\n"
		}
	}

	return statusMessage, nil
}

// Commit Creates a new commit
func (repo *GitRepo) Commit(message string) (string, error) {
	errLoadCfg := repo.LoadConfig()
	if errLoadCfg != nil {
		return "", errLoadCfg
	}
	commitContent := fmt.Sprintf("commit %s\nAuthor: %s <%s>\nDate: %s\n\n%s\n", time.Now().Format("Mon Jan 2 15:04:05 2006 -0700"), repo.Username, repo.Email, time.Now().Format(time.RFC1123), message)

	commitHash := utils.HashObject(commitContent)

	commitPath := filepath.Join(repo.Path, "objects", commitHash[:2], commitHash[2:])

	err := os.MkdirAll(filepath.Join(repo.Path, "objects", commitHash[:2]), os.ModePerm)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(commitPath, []byte(commitContent), os.ModePerm)
	if err != nil {
		return "", err
	}

	return commitHash, nil
}

// Diff calculates and displays the difference between files
func (repo *GitRepo) Diff() (string, error) {
	// Get the list of files in the working directory
	files, err := utils.ListFiles(repo.Path)
	if err != nil {
		return "", err
	}

	// Get the list of files in the index
	indexPath := filepath.Join(repo.Path, "index")
	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		return "", err
	}
	indexFiles := strings.Split(strings.TrimSpace(string(indexContent)), "\n")

	var diffOutput string

	// Loop through each file in the working directory
	for _, file := range files {
		// Check if the file is in the index
		found := false
		for _, indexFile := range indexFiles {
			if indexFile == file {
				found = true
				break
			}
		}

		// If the file is not in the index, it's untracked
		if !found {
			diffOutput += fmt.Sprintf("diff --git a/%s b/%s\n", file, file)
			diffOutput += fmt.Sprintf("new file mode 100644\n")
			diffOutput += fmt.Sprintf("--- /dev/null\n")
			diffOutput += fmt.Sprintf("+++ b/%s\n", file)

			// Read the content of the file
			filePath := filepath.Join(repo.Path, file)
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				return "", err
			}

			// Add the content to the diff output
			diffOutput += string(fileContent) + "\n"
		}
	}

	return diffOutput, nil
}

// Logs retrieves the history of commits
func (repo *GitRepo) Logs() (string, error) {
	var logsOutput string

	// Walk through all objects in the objects directory
	objectsDir := filepath.Join(repo.Path, "objects")
	err := filepath.Walk(objectsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// Read the content of each object
			objectContent, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Decode the object content
			decodedContent, err := utils.DecodeObject(objectContent)
			if err != nil {
				return err
			}

			// Check if the object is a commit
			if strings.HasPrefix(decodedContent, "commit") {
				// Extract commit information
				logsOutput += "commit " + info.Name() + "\n"
				logsOutput += decodedContent + "\n\n"
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return logsOutput, nil
}
