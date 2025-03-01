package githubapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func newClient(token string) *Client {
	return &Client {
		BaseURL: "https://api.github.com",
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func sendGETRequest(url string, headers map[string]string) ([]byte, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("error while preparing the request: %v", err)
    }

    for key, value := range headers {
        req.Header.Add(key, value)
    }

    client := &http.Client{}
    res, err := client.Do(req)
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

func FetchLatestCommitSHA(repoOwner, repoName string, headers map[string]string) (string, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/heads/main", repoOwner, repoName)

    body, err := sendGETRequest(url, headers)
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

func FetchTreeSHA(commitSHA, repoOwner, repoName string, headers map[string]string) (string, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/commits/%s", repoOwner, repoName, commitSHA)

    body, err := sendGETRequest(url, headers)
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

func FetchFileTree(treeSHA, repoOwner, repoName string, headers map[string]string) ([]TreeEntry, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", repoOwner, repoName, treeSHA)

    body, err := sendGETRequest(url, headers)
    if err != nil {
        return nil, fmt.Errorf("error in sendGETRequest: %v", err)
    }

    var response FileTreeResponse
    if err := json.Unmarshal(body, &response); err != nil {
        return nil, fmt.Errorf("error while parsing JSON: %v", err)
    }

    return response.Tree, nil
}

func FetchFileContent(fileURL string, headers map[string]string) (string, error) {
    body, err := sendGETRequest(fileURL, headers)
    if err != nil {
        return "", fmt.Errorf("error while fetching file content: %v", err)
    }
    return string(body), nil
}
