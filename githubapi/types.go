package githubapi

import "net/http"

type Client struct {
	BaseURL			string
	Token				string
	HTTPClient	*http.Client
}


type GitRefResponse struct {
    Object struct {
        SHA string `json:"sha"`
    } `json:"object"`
}

type CommitResponse struct {
    Tree struct {
        SHA string `json:"sha"`
    } `json:"tree"`
}

type FileTreeResponse struct {
    SHA       string      `json:"sha"`
    URL       string      `json:"url"`
    Tree      []TreeEntry `json:"tree"`
    Truncated bool        `json:"truncated"`
}

type TreeEntry struct {
    Path string `json:"path"`
    Mode string `json:"mode"`
    Type string `json:"type"`
    SHA  string `json:"sha"`
    URL  string `json:"url"`
}