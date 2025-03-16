// github_client.go
package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"
)

func NewClient(token string) *Client {
	return &Client{
		BaseURL: "https://api.github.com/repos",
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 100 * time.Second,
		},
	}
}

func (c *Client) sendGETRequest(ctx context.Context, endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error while preparing the request: %v", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3.raw")
	req.Header.Set("Authorization", "token "+c.Token)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while sending the request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading the response: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API Error (%d): %s", res.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) FetchLatestCommitSHA(ctx context.Context, repoOwner, repoName string) (string, error) {
	branches := []string{"main", "master"}
	var err error
	var body []byte

	for _, branch := range branches {
		url := fmt.Sprintf("/%s/%s/git/refs/heads/%s", repoOwner, repoName, branch)
		body, err = c.sendGETRequest(ctx, url)
		if err == nil {
			break
		}
	}

	if err != nil {
		return "", fmt.Errorf("error in sendGETRequest: %v", err)
	}

	var response GitRefResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("error while parsing JSON: %v", err)
	}

	if response.Object.SHA == "" {
		return "", fmt.Errorf("SHA not found in response")
	}

	return response.Object.SHA, nil
}

func (c *Client) FetchTreeSHA(ctx context.Context, commitSHA, repoOwner, repoName string) (string, error) {
	url := fmt.Sprintf("/%s/%s/git/commits/%s", repoOwner, repoName, commitSHA)

	body, err := c.sendGETRequest(ctx, url)
	if err != nil {
		return "", fmt.Errorf("error in sendGETRequest: %v", err)
	}

	var response CommitResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("error while parsing JSON: %v", err)
	}

	if response.Tree.SHA == "" {
		return "", fmt.Errorf("tree SHA not found in response")
	}

	return response.Tree.SHA, nil
}

func (c *Client) FetchFileTree(ctx context.Context, selectedFiles []string, treeSHA, repoOwner, repoName string) ([]TreeEntry, error) {
	url := fmt.Sprintf("/%s/%s/git/trees/%s?recursive=1", repoOwner, repoName, treeSHA)

	body, err := c.sendGETRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("error in send GET Request: %v", err)
	}

	var response FileTreeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error while parsing JSON: %w", err)
	}

	responseTree := make([]TreeEntry, 0, len(response.Tree))

	for _, treeEntry := range response.Tree {
		// only files not folders
		if treeEntry.Type != "blob" {
			continue
		}

		if len(selectedFiles) == 0 || slices.Contains(selectedFiles, treeEntry.Path) {
			responseTree = append(responseTree, treeEntry)
		}
	}

	return responseTree, nil
}

func (c *Client) FetchFileContent(ctx context.Context, fileURL string) (string, error) {
	body, err := c.sendGETRequest(ctx, fileURL)
	if err != nil {
		return "", fmt.Errorf("error while fetching file content: %v", err)
	}
	return string(body), nil
}
