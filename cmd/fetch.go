// fetch.go
package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	pkg "github.com/julilili42/GitSnatch/pkg"
	"github.com/spf13/cobra"
)

func newClient(token string) *pkg.Client {
	return &pkg.Client{
		BaseURL: "https://api.github.com/repos",
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 100 * time.Second,
		},
	}
}

var fetchCmd = &cobra.Command{
	Use:   "fetch [repoOwner] [repoName]",
	Short: "Fetch all file contents from a GitHub repository",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		repoOwner := args[0]
		if repoOwner == "" {
			repoOwner = pkg.AskQuestion("Who is the owner of the GitHub repo?")
		}

		repoName := args[1]
		if repoName == "" {
			repoName = pkg.AskQuestion("What is the name of the GitHub repo?")
		}

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

		treeSHA, err := client.FetchTreeSHA(ctx, commitSHA, repoOwner, repoName)
		if err != nil {
			log.Fatalf("Error fetching tree SHA: %v", err)
		}

		fileTree, err := client.FetchFileTree(ctx, treeSHA, repoOwner, repoName)
		if err != nil {
			log.Fatalf("Error fetching file tree: %v", err)
		}

		var wg sync.WaitGroup
		contentChannel := make(chan string, len(fileTree))

		var paths []string

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
					paths = append(paths, fmt.Sprintf("%s \n", path))
				}(entry.Path, entry.URL)
			}
		}

		go func() {
			wg.Wait()
			close(contentChannel)
		}()

		var collectedLines []string

		for content := range contentChannel {
			collectedLines = append(collectedLines, content)
		}

		finalContent := strings.Join(collectedLines, "")

		test := pkg.MultiSelect("Select the files you want to copy.", paths)

		fmt.Println(test)

		if err := clipboard.WriteAll(finalContent); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
