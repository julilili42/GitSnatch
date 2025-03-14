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

func fetchRepositoryData(ctx context.Context, client *pkg.Client, selectedFiles []string, repoOwner, repoName string) ([]pkg.TreeEntry, []string, error) {
	commitSHA, err := client.FetchLatestCommitSHA(ctx, repoOwner, repoName)
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching latest commit SHA: %v", err)
	}

	treeSHA, err := client.FetchTreeSHA(ctx, commitSHA, repoOwner, repoName)
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching tree SHA: %v", err)
	}

	fileTree, err := client.FetchFileTree(ctx, selectedFiles, treeSHA, repoOwner, repoName)
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching file tree: %v", err)
	}

	paths := getPaths(fileTree)

	return fileTree, paths, nil
}

func getPaths(fileTree []pkg.TreeEntry) []string {
	paths := make([]string, 0, len(fileTree))
	for _, entry := range fileTree {
		path := entry.Path
		paths = append(paths, path)
	}

	return paths
}

func filterFileTree(fullTree []pkg.TreeEntry, selectedPaths []string) []pkg.TreeEntry {
	selectedSet := make(map[string]struct{})
	for _, p := range selectedPaths {
		selectedSet[p] = struct{}{}
	}

	var filtered []pkg.TreeEntry
	for _, entry := range fullTree {
		if _, ok := selectedSet[entry.Path]; ok {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func processFileTree(ctx context.Context, client *pkg.Client, fileTree []pkg.TreeEntry) string {
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

	var collectedLines []string
	for content := range contentChannel {
		collectedLines = append(collectedLines, content)
	}

	return strings.Join(collectedLines, "")
}

var fetchCmd = &cobra.Command{
	Use:   "fetch [repoOwner] [repoName]",
	Short: "Fetch all file contents from a GitHub repository",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		var repoOwner, repoName string

		if len(args) >= 1 {
			repoOwner = args[0]
		} else {
			repoOwner = pkg.AskQuestion("Who is the owner of the GitHub repo?")
		}

		if len(args) >= 2 {
			repoName = args[1]
		} else {
			repoName = pkg.AskQuestion("What is the name of the GitHub repo?")
		}

		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			log.Fatal("GITHUB_TOKEN environment variable not set")
		}

		client := newClient(token)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		fileTree, paths, err := fetchRepositoryData(ctx, client, nil, repoOwner, repoName)
		if err != nil {
			log.Fatal(err)
		}

		selectedFiles := pkg.MultiSelect("Select the files you want to copy.", paths)

		selectedTree := filterFileTree(fileTree, selectedFiles)

		finalContent := processFileTree(ctx, client, selectedTree)

		if err := clipboard.WriteAll(finalContent); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		pkg.InfoText("Copied to clipboard successfully.")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
