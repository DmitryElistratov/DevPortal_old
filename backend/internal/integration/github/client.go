package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		baseURL:    "https://api.github.com",
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// NewEnterpriseClient creates a client for GitHub Enterprise (or github.com).
// baseURL is normalized: trailing slashes are stripped and common UI suffixes
// (/login, /settings/apps, path segments) are removed so the API base never
// contains double slashes.
func NewEnterpriseClient(baseURL, token string) *Client {
	if baseURL == "" {
		return NewClient(token)
	}
	u := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	for _, suffix := range []string{"/login", "/settings", "/dashboard", "/api/v3", "/api"} {
		u = strings.TrimSuffix(u, suffix)
	}
	u = strings.TrimRight(u, "/")
	if u == "" || u == "https://github.com" || u == "http://github.com" || u == "https://www.github.com" {
		return NewClient(token)
	}
	if !strings.HasPrefix(u, "http") {
		u = "https://" + u
	}
	return &Client{
		baseURL:    u + "/api/v3",
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type Repository struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
	Private  bool   `json:"private"`
}

type Branch struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

type Tag struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

// TestConnection tests the connection to GitHub
func (c *Client) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/user", c.baseURL), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListRepositories returns list of GitHub repositories
func (c *Client) ListRepositories(ctx context.Context) ([]Repository, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/user/repos", c.baseURL), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API error: %d - %s", resp.StatusCode, string(body))
	}

	var repos []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// ListBranches returns list of branches for a repository
func (c *Client) ListBranches(ctx context.Context, owner, repo string) ([]Branch, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/repos/%s/%s/branches", c.baseURL, owner, repo), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API error: %d - %s", resp.StatusCode, string(body))
	}

	var branches []Branch
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}

	return branches, nil
}

// ListTags returns list of tags for a repository
func (c *Client) ListTags(ctx context.Context, owner, repo string) ([]Tag, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/repos/%s/%s/tags", c.baseURL, owner, repo), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API error: %d - %s", resp.StatusCode, string(body))
	}

	var tags []Tag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	return tags, nil
}

// CreateTag creates a new tag in the repository
func (c *Client) CreateTag(ctx context.Context, owner, repo, tagName, sha string) error {
	payload := map[string]string{
		"tag": tagName,
		"sha": sha,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/repos/%s/%s/git/refs", c.baseURL, owner, repo), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(body))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}
