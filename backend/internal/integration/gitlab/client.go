package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// ---------- CI/CD pipelines & Merge Requests (TestOps) ----------

// Pipeline — сжатое представление GitLab CI pipeline.
type Pipeline struct {
	ID     int    `json:"id"`
	SHA    string `json:"sha"`
	Ref    string `json:"ref"`
	Status string `json:"status"`
	WebURL string `json:"web_url"`
}

// Job — stage/status пайплайна для отображения в UI.
type Job struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Stage    string  `json:"stage"`
	Status   string  `json:"status"`
	Coverage float64 `json:"coverage"`
}

// SecurityReport — агрегированный отчёт безопасности GitLab (SAST/DAST/Secret Detection).
type SecurityReport struct {
	Total     int      `json:"total"`
	Critical  int      `json:"critical"`
	High      int      `json:"high"`
	Medium    int      `json:"medium"`
	Low       int      `json:"low"`
	Unknown   int      `json:"unknown"`
	Available bool     `json:"available"` // найдены ли security-данные
	Sources   []string `json:"sources,omitempty"`
}

// GetSecurityReport собирает security report для пайплайнов на ref.
// Сначала пробуется /api/v4/projects/:id/security-report (GitLab 13.3+),
// при недоступности — fallback по наличию security-job'ов в последнем пайплайне.
func (c *Client) GetSecurityReport(ctx context.Context, projectID, ref string) (*SecurityReport, error) {
	rep := &SecurityReport{}
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/security-report?per_page=100", c.baseURL, url.QueryEscape(projectID))
	if ref != "" {
		apiURL += "&ref=" + url.QueryEscape(ref)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err == nil {
		req.Header.Set("PRIVATE-TOKEN", c.token)
		resp, err := c.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			var data struct {
				Findings struct {
					Count int `json:"count"`
					Nodes []struct {
						Severity string `json:"severity"`
					} `json:"nodes"`
				} `json:"findings"`
			}
			if json.NewDecoder(resp.Body).Decode(&data) == nil && (data.Findings.Count > 0 || len(data.Findings.Nodes) >= 0) {
				rep.Available = true
				rep.Total = data.Findings.Count
				for _, n := range data.Findings.Nodes {
					switch strings.ToLower(n.Severity) {
					case "critical":
						rep.Critical++
					case "high":
						rep.High++
					case "medium":
						rep.Medium++
					case "low":
						rep.Low++
					default:
						rep.Unknown++
					}
				}
				resp.Body.Close()
				return rep, nil
			}
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	// fallback: ищем security-job'ы в последнем пайплайне на ref
	pipelines, err := c.ListPipelines(ctx, projectID, ref)
	if err != nil || len(pipelines) == 0 {
		return rep, nil
	}
	jobs, err := c.GetPipelineJobs(ctx, projectID, pipelines[0].ID)
	if err != nil {
		return rep, nil
	}
	for _, j := range jobs {
		lname := strings.ToLower(j.Name)
		if strings.Contains(lname, "sast") || strings.Contains(lname, "secret") ||
			strings.Contains(lname, "dast") || strings.Contains(lname, "security") {
			rep.Available = true
			rep.Sources = append(rep.Sources, j.Name)
		}
	}
	return rep, nil
}

// TestSummary — результаты тестов последнего пайплайна на ref.
type TestSummary struct {
	Status      string  `json:"status"` // success | failed | running | pending | none
	Passed      int     `json:"passed"`
	Failed      int     `json:"failed"`
	Skipped     int     `json:"skipped"`
	Total       int     `json:"total"`
	PassRate    float64 `json:"passRate"`
	Coverage    float64 `json:"coverage"`
	PipelineID  int     `json:"pipelineId,omitempty"`
	PipelineURL string  `json:"pipelineUrl,omitempty"`
	Ref         string  `json:"ref,omitempty"`
}

// GetTestSummary собирает статусы тестовых job'ов последнего пайплайна ref.
func (c *Client) GetTestSummary(ctx context.Context, projectID, ref string) (*TestSummary, error) {
	ts := &TestSummary{Status: "none"}
	pipelines, err := c.ListPipelines(ctx, projectID, ref)
	if err != nil || len(pipelines) == 0 {
		return ts, nil
	}
	pl := pipelines[0]
	ts.PipelineID = pl.ID
	ts.PipelineURL = pl.WebURL
	ts.Ref = pl.Ref
	ts.Status = pl.Status
	jobs, err := c.GetPipelineJobs(ctx, projectID, pl.ID)
	if err == nil {
		var covSum float64
		var covCnt int
		for _, j := range jobs {
			lname := strings.ToLower(j.Name)
			isTest := strings.Contains(lname, "test") || strings.Contains(lname, "spec") ||
				strings.Contains(lname, "e2e") || strings.Contains(lname, "unit") ||
				strings.Contains(lname, "integration")
			if isTest {
				switch strings.ToLower(j.Status) {
				case "success", "passed":
					ts.Passed++
				case "failed":
					ts.Failed++
				case "skipped", "canceled", "manual":
					ts.Skipped++
				}
			}
			if j.Coverage > 0 {
				covSum += j.Coverage
				covCnt++
			}
		}
		ts.Total = ts.Passed + ts.Failed + ts.Skipped
		if ts.Total > 0 {
			ts.PassRate = float64(ts.Passed) / float64(ts.Total) * 100
		}
		if covCnt > 0 {
			ts.Coverage = covSum / float64(covCnt)
		}
	}
	return ts, nil
}

// ListPipelines returns recent pipelines for a project, optionally filtered by ref (branch).
func (c *Client) ListPipelines(ctx context.Context, projectID, ref string) ([]Pipeline, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/pipelines?per_page=20", c.baseURL, projectID)
	if ref != "" {
		apiURL += "&ref=" + url.QueryEscape(ref)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
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
	var pipelines []Pipeline
	if err := json.NewDecoder(resp.Body).Decode(&pipelines); err != nil {
		return nil, err
	}
	return pipelines, nil
}

// GetPipelineJobs returns jobs of a pipeline.
func (c *Client) GetPipelineJobs(ctx context.Context, projectID string, pipelineID int) ([]Job, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/pipelines/%d/jobs", c.baseURL, projectID, pipelineID)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
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
	var jobs []Job
	if err := json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

// TriggerPipeline starts a new pipeline on the given ref with optional variables.
func (c *Client) TriggerPipeline(ctx context.Context, projectID, ref string, vars map[string]string) (*Pipeline, error) {
	form := url.Values{}
	form.Set("ref", ref)
	for k, v := range vars {
		form.Add("variables[]", k+"="+v)
	}

	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/pipeline", c.baseURL, projectID)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitlab API error: %d - %s", resp.StatusCode, string(body))
	}
	var p Pipeline
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// MergeRequest — сжатое представление MR.
type MergeRequest struct {
	IID          int    `json:"iid"`
	Title        string `json:"title"`
	State        string `json:"state"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Author       struct {
		Username string `json:"username"`
	} `json:"author"`
	WebURL string `json:"web_url"`
}

// ListMergeRequests returns MRs for a project; state may be opened|merged|closed or empty for all.
func (c *Client) ListMergeRequests(ctx context.Context, projectID, state string) ([]MergeRequest, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests?per_page=30&order_by=updated_at", c.baseURL, projectID)
	if state != "" {
		apiURL += "&state=" + url.QueryEscape(state)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
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
	var mrs []MergeRequest
	if err := json.NewDecoder(resp.Body).Decode(&mrs); err != nil {
		return nil, err
	}
	return mrs, nil
}
