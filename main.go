package main

import (
	"fmt"
	"log"
	"os"

	"github.com/julilili42/GitSnatch/githubapi"
)

func main() {
	repoOwner := "julilili42"
	repoName := "InventoryManager"
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable not set")
	}
	
	headers := map[string]string{
		"Accept":        "application/vnd.github.v3.raw",
		"Authorization": "token " + token,
	}

	commitSHA, err := githubapi.FetchLatestCommitSHA(repoOwner, repoName, headers)
	if err != nil {
		log.Fatalf("Error fetching latest commit SHA: %v", err)
	}
	fmt.Println("Latest Commit SHA:", commitSHA)

	treeSHA, err := githubapi.FetchTreeSHA(commitSHA, repoOwner, repoName, headers)
	if err != nil {
		log.Fatalf("Error fetching tree SHA: %v", err)
	}
	fmt.Println("Tree SHA:", treeSHA)

	fileTree, err := githubapi.FetchFileTree(treeSHA, repoOwner, repoName, headers)
	if err != nil {
		log.Fatalf("Error fetching file tree: %v", err)
	}

	for _, entry := range fileTree {
		if entry.Type == "blob" {
			content, err := githubapi.FetchFileContent(entry.URL, headers)
			if err != nil {
				log.Printf("Error fetching content for %s: %v\n", entry.Path, err)
				continue
			}
			fmt.Printf("----- %s -----\n%s\n\n", entry.Path, content)
		}
	}
}
