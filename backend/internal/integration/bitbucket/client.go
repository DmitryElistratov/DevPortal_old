package bitbucket

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

func NewClient(username, password string) *Client {
	return &Client{
		baseURL:    "https://api.bitbucket.org/2.0",
		username:   username,
		password:   password,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type Repository struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	FullName  string `json:"full_name"`
	Links     Links  `json:"links"`
	IsPrivate bool   `json:"is_private"`
}

type Links struct {
	HTML struct {
		Href string `json:"href"`
	} `json:"html"`
}

type Branch struct {
	Name   string `json:"name"`
	Target struct {
		Hash string `json:"hash"`
	} `json:"target"`
}

type Tag struct {
	Name   string `json:"name"`
	Target struct {
		Hash string `json:"hash"`
	} `json:"target"`
}

func (c *Client) getAuth() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(c.username+":"+c.password))
}

// TestConnection tests the connection to Bitbucket
func (c *Client) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/user", c.baseURL), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.getAuth())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bitbucket API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListRepositories returns list of Bitbucket repositories
func (c *Client) ListRepositories(ctx context.Context) ([]Repository, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/repositories/%s", c.baseURL, c.username), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.getAuth())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bitbucket API error: %d - %s", resp.StatusCode, string(body))
	}

	var result struct {
		Values []Repository `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Values, nil
}

// ListBranches returns list of branches for a repository
func (c *Client) ListBranches(ctx context.Context, workspace, repo string) ([]Branch, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/repositories/%s/%s/refs/branches", c.baseURL, workspace, repo), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.getAuth())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bitbucket API error: %d - %s", resp.StatusCode, string(body))
	}

	var result struct {
		Values []Branch `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Values, nil
}

// ListTags returns list of tags for a repository
func (c *Client) ListTags(ctx context.Context, workspace, repo string) ([]Tag, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/repositories/%s/%s/refs/tags", c.baseURL, workspace, repo), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.getAuth())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bitbucket API error: %d - %s", resp.StatusCode, string(body))
	}

	var result struct {
		Values []Tag `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Values, nil
}

// CreateTag creates a new tag in the repository
func (c *Client) CreateTag(ctx context.Context, workspace, repo, tagName, hash string) error {
	payload := map[string]interface{}{
		"name": tagName,
		"target": map[string]string{
			"hash": hash,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/repositories/%s/%s/refs/tags", c.baseURL, workspace, repo), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.getAuth())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bitbucket API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}
