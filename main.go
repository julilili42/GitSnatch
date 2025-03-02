// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/julilili42/GitSnatch/githubapi"
)

func newClient(token string) *githubapi.Client {
	return &githubapi.Client{
		BaseURL: "https://api.github.com/repos",
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 100 * time.Second,
		},
	}
}

func main() {
	repoOwner := "julilili42"
	repoName := "InventoryManager"
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable not set")
	}

	client := newClient(token)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	commitSHA, err := client.FetchLatestCommitSHA(ctx, repoOwner, repoName)
	if err != nil {
		log.Fatalf("Error fetching latest commit SHA: %v", err)
	}
	fmt.Println("Latest Commit SHA:", commitSHA)

	treeSHA, err := client.FetchTreeSHA(ctx, commitSHA, repoOwner, repoName)
	if err != nil {
		log.Fatalf("Error fetching tree SHA: %v", err)
	}
	fmt.Println("Tree SHA:", treeSHA)

	fileTree, err := client.FetchFileTree(ctx, treeSHA, repoOwner, repoName)
	if err != nil {
		log.Fatalf("Error fetching file tree: %v", err)
	}

	var wg sync.WaitGroup
	contentChannel := make(chan string, len(fileTree))

	for _, entry := range fileTree {
		if entry.Type == "blob" {
			wg.Add(1)
			go func(path, url string) {
				defer wg.Done()

				shorturl := strings.Replace(url, client.BaseURL, "", 1)
				content, err := client.FetchFileContent(ctx, shorturl)
				if err != nil {
					log.Printf("Error fetching content for %s: %v\n", path, err)
					return
				}

				contentChannel <- fmt.Sprintf("----- %s -----\n%s\n", path, content)
			}(entry.Path, entry.URL)
		}
	}

	go func() {
		wg.Wait()
		close(contentChannel)
	}()

	for content := range contentChannel {
		fmt.Println(content)
	}
}
