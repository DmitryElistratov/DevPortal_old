package gitlab

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

func NewClient(baseURL, token string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// TestConnection tests the connection to GitLab
func (c *Client) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v4/user", c.baseURL), nil)
	if err != nil {
		return err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gitlab API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
	HTTPURLToRepo     string `json:"http_url_to_repo"`
}

// ProjectID exposes the numeric GitLab project id for member/permission lookups.
func (p Project) ProjectID() string { return fmt.Sprintf("%d", p.ID) }

// GitLabMember represents a project member with their access level.
type GitLabMember struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	State       string `json:"state"`
	AccessLevel int    `json:"access_level"`
}

// ListProjectMembers returns all members of a project, including members
// inherited from ancestor group shares (GitLab API: GET /projects/:id/members/all).
func (c *Client) ListProjectMembers(ctx context.Context, projectID string) ([]GitLabMember, error) {
	var members []GitLabMember
	for page := 1; ; page++ {
		url := fmt.Sprintf("%s/api/v4/projects/%s/members/all?per_page=100&page=%d", c.baseURL, projectID, page)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("PRIVATE-TOKEN", c.token)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		// Some self-managed instances may not expose /members/all — fall back to /members.
		if resp.StatusCode == http.StatusNotFound && page == 1 {
			resp.Body.Close()
			url = fmt.Sprintf("%s/api/v4/projects/%s/members?per_page=100", c.baseURL, projectID)
			req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("PRIVATE-TOKEN", c.token)
			resp, err = c.httpClient.Do(req)
			if err != nil {
				return nil, err
			}
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("gitlab members API error: %d - %s", resp.StatusCode, string(body))
		}

		var batch []GitLabMember
		decodeErr := json.NewDecoder(resp.Body).Decode(&batch)
		resp.Body.Close()
		if decodeErr != nil {
			return nil, decodeErr
		}
		members = append(members, batch...)
		if len(batch) < 100 {
			break
		}
	}
	return members, nil
}

// AccessLevelToRole maps a GitLab numeric access level to the portal role model.
func AccessLevelToRole(level int) string {
	switch {
	case level >= 50: // Owner
		return "admin"
	case level >= 40: // Maintainer
		return "maintainer"
	case level >= 30: // Developer
		return "developer"
	case level >= 20: // Reporter
		return "reporter"
	case level >= 10: // Guest
		return "guest"
	default:
		return ""
	}
}

type Branch struct {
	Name   string `json:"name"`
	Commit struct {
		ID string `json:"id"`
	} `json:"commit"`
}

type Tag struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Commit  struct {
		ID string `json:"id"`
	} `json:"commit"`
}

// ListProjects returns list of GitLab projects
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v4/projects", c.baseURL), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitlab API error: %d - %s", resp.StatusCode, string(body))
	}

	var projects []Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// ListBranches returns list of branches for a project
func (c *Client) ListBranches(ctx context.Context, projectID string) ([]Branch, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v4/projects/%s/repository/branches", c.baseURL, projectID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitlab API error: %d - %s", resp.StatusCode, string(body))
	}

	var branches []Branch
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}

	return branches, nil
}

// ListTags returns list of tags for a project
func (c *Client) ListTags(ctx context.Context, projectID string) ([]Tag, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v4/projects/%s/repository/tags", c.baseURL, projectID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitlab API error: %d - %s", resp.StatusCode, string(body))
	}

	var tags []Tag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	return tags, nil
}

// CreateTag creates a new tag in the repository
func (c *Client) CreateTag(ctx context.Context, projectID, tagName, ref string) error {
	payload := map[string]string{
		"tag_name": tagName,
		"ref":      ref,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v4/projects/%s/repository/tags", c.baseURL, projectID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gitlab API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// UpdateBranch updates a branch to point to a specific commit
func (c *Client) UpdateBranch(ctx context.Context, projectID, branchName, commitID string) error {
	payload := map[string]string{
		"branch":    branchName,
		"commit_id": commitID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/api/v4/projects/%s/repository/branches/%s", c.baseURL, projectID, branchName), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gitlab API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}
