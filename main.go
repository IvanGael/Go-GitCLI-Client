package main

import (
	"fmt"
	"git-client/utils"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: git <command> [args]")
		os.Exit(1)
	}

	command := os.Args[1]
	repo, err := utils.NewGitRepo(".")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	switch command {
	case "config":
		if len(os.Args) != 4 {
			fmt.Println("Usage: git config <username> <email>")
			os.Exit(1)
		}
		username := os.Args[2]
		email := os.Args[3]
		err := repo.SetConfig(username, email)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("Config set successfully:", username, email)
	case "init":
		err := repo.Init()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("Initialized empty repository:", repo.Path)

	case "add":
		if len(os.Args) != 3 {
			fmt.Println("Usage: git add <file>")
			os.Exit(1)
		}
		filePath := os.Args[2]
		err := repo.Add(filePath)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("Added to index:", filePath)

	case "status":
		status, err := repo.Status()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Print(status)

	case "commit":
		if len(os.Args) < 3 {
			fmt.Println("Usage: git commit <message>")
			os.Exit(1)
		}
		message := os.Args[2]
		commitHash, err := repo.Commit(message)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("Committed to master:", commitHash)

	case "diff":
		diff, err := repo.Diff()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Print(diff)

	case "logs":
		diff, err := repo.Logs()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Print(diff)

	default:
		fmt.Println("Unknown command:", command)
		os.Exit(1)
	}
}
