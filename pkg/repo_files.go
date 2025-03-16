package pkg

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
)

func FetchRepositoryData(ctx context.Context, client *Client, selectedFiles []string, repoOwner, repoName, commitSHA string) ([]TreeEntry, error) {

	commit := commitSHA

	if commitSHA == "" {
		var err error
		commit, err = client.FetchLatestCommitSHA(ctx, repoOwner, repoName)
		if err != nil {
			return nil, fmt.Errorf("error fetching latest commit SHA: %v", err)
		}
	}

	treeSHA, err := client.FetchTreeSHA(ctx, commit, repoOwner, repoName)
	if err != nil {
		return nil, fmt.Errorf("error fetching tree SHA: %v", err)
	}

	fileTree, err := client.FetchFileTree(ctx, selectedFiles, treeSHA, repoOwner, repoName)
	if err != nil {
		return nil, fmt.Errorf("error fetching file tree: %v", err)
	}

	return fileTree, nil
}

func GetSortedFilePaths(fileTree []TreeEntry) []string {
	paths := make([]string, 0, len(fileTree))
	for _, entry := range fileTree {
		path := entry.Path
		paths = append(paths, path)
	}
	sort.Slice(paths, func(i, j int) bool {
		return strings.ToLower(paths[i]) < strings.ToLower(paths[j])
	})

	return paths
}

func FilterFileTree(fullTree []TreeEntry, selectedPaths []string) []TreeEntry {
	selectedSet := make(map[string]struct{})
	for _, p := range selectedPaths {
		selectedSet[p] = struct{}{}
	}

	var filtered []TreeEntry
	for _, entry := range fullTree {
		if _, ok := selectedSet[entry.Path]; ok {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func ProcessFileTree(ctx context.Context, client *Client, fileTree []TreeEntry) string {
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
