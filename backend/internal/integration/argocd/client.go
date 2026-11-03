package argocd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ErrNoIntegration is returned when no enabled ArgoCD integration with a
// valid URL and token is configured in the portal.
var ErrNoIntegration = errors.New("no enabled ArgoCD integration found — configure it in Settings (URL + token), press Test Connection and Sync")

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

type Application struct {
	Metadata Metadata `json:"metadata"`
	Spec     Spec     `json:"spec"`
	Status   Status   `json:"status"`
}

type Metadata struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Annotations map[string]string `json:"annotations"`
}

type Spec struct {
	Source      Source      `json:"source"`
	Destination Destination `json:"destination"`
}

type Source struct {
	RepoURL        string `json:"repoURL"`
	Path           string `json:"path"`
	TargetRevision string `json:"targetRevision"`
}

type Destination struct {
	Server    string `json:"server"`
	Namespace string `json:"namespace"`
}

type Status struct {
	Sync   SyncStatus   `json:"sync"`
	Health HealthStatus `json:"health"`
}

type SyncStatus struct {
	Status   string `json:"status"`
	Revision string `json:"revision"`
}

type HealthStatus struct {
	Status string `json:"status"`
}

// TestConnection tests the connection to ArgoCD
func (c *Client) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/session/userinfo", c.baseURL), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("argocd API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListApplications returns list of ArgoCD applications
func (c *Client) ListApplications(ctx context.Context) ([]Application, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/applications", c.baseURL), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("argocd API error: %d - %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []Application `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Items, nil
}

// GetApplication returns details of a specific application
func (c *Client) GetApplication(ctx context.Context, appName string) (*Application, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/applications/%s", c.baseURL, appName), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("argocd API error: %d - %s", resp.StatusCode, string(body))
	}

	var app Application
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, err
	}

	return &app, nil
}

// SyncApplication triggers sync for an application
func (c *Client) SyncApplication(ctx context.Context, appName string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/applications/%s/sync", c.baseURL, appName), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("argocd API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// patchApplication sends a JSON-patch request to update an ArgoCD application.
func (c *Client) patchApplication(ctx context.Context, appName string, patches []map[string]interface{}) error {
	body, err := json.Marshal(patches)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/applications/%s", c.baseURL, appName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// ArgoCD REST API accepts PATCH semantics via PUT with a JSON patch array;
	// some versions return 200, tolerate that plus 204.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("argocd API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// SetTargetRevision points the application source at a specific revision
// (commit SHA / tag / branch) without touching traffic annotations.
func (c *Client) SetTargetRevision(ctx context.Context, appName, revision string) error {
	patches := []map[string]interface{}{
		{
			"op":    "replace",
			"path":  "/spec/source/targetRevision",
			"value": revision,
		},
	}
	return c.patchApplication(ctx, appName, patches)
}

// CreateCanary creates a canary deployment with traffic splitting:
// sets targetRevision to the canary ref, applies canary annotations and
// triggers a sync so ArgoCD actually rolls out the new revision.
func (c *Client) CreateCanary(ctx context.Context, appName, version string, percentage int) error {
	patches := []map[string]interface{}{
		{
			"op":    "replace",
			"path":  "/spec/source/targetRevision",
			"value": version,
		},
		{
			"op":    "add",
			"path":  "/metadata/annotations/canary.argoproj.io~1traffic-split",
			"value": fmt.Sprintf("%d", percentage),
		},
		{
			"op":    "add",
			"path":  "/metadata/annotations/canary.argoproj.io~1mode",
			"value": "canary",
		},
	}

	if err := c.patchApplication(ctx, appName, patches); err != nil {
		return err
	}

	// Trigger sync so the new revision is actually deployed
	return c.SyncApplication(ctx, appName)
}

// UpdateCanaryTraffic updates traffic percentage for canary deployment
func (c *Client) UpdateCanaryTraffic(ctx context.Context, appName string, percentage int) error {
	patches := []map[string]interface{}{
		{
			"op":    "add",
			"path":  "/metadata/annotations/canary.argoproj.io~1traffic-split",
			"value": fmt.Sprintf("%d", percentage),
		},
	}
	return c.patchApplication(ctx, appName, patches)
}

// PromoteCanary promotes canary to 100% traffic and clears canary mode.
func (c *Client) PromoteCanary(ctx context.Context, appName string) error {
	patches := []map[string]interface{}{
		{
			"op":    "add",
			"path":  "/metadata/annotations/canary.argoproj.io~1traffic-split",
			"value": "100",
		},
		{
			"op":    "add",
			"path":  "/metadata/annotations/canary.argoproj.io~1mode",
			"value": "stable",
		},
	}
	if err := c.patchApplication(ctx, appName, patches); err != nil {
		return err
	}
	return c.SyncApplication(ctx, appName)
}

// RollbackCanary rolls back canary to 0% traffic and returns app to stable mode.
func (c *Client) RollbackCanary(ctx context.Context, appName string) error {
	patches := []map[string]interface{}{
		{
			"op":    "add",
			"path":  "/metadata/annotations/canary.argoproj.io~1traffic-split",
			"value": "0",
		},
		{
			"op":    "add",
			"path":  "/metadata/annotations/canary.argoproj.io~1mode",
			"value": "stable",
		},
	}
	if err := c.patchApplication(ctx, appName, patches); err != nil {
		return err
	}
	return c.SyncApplication(ctx, appName)
}

// DeployVersion deploys a specific version
func (c *Client) DeployVersion(ctx context.Context, appName, version string, percentage int) error {
	return c.CreateCanary(ctx, appName, version, percentage)
}

// UpdateTrafficPercentage updates traffic percentage
func (c *Client) UpdateTrafficPercentage(ctx context.Context, appName string, percentage int) error {
	return c.UpdateCanaryTraffic(ctx, appName, percentage)
}

// PromoteVersion promotes a version to 100% traffic
func (c *Client) PromoteVersion(ctx context.Context, appName, version string) error {
	return c.PromoteCanary(ctx, appName)
}

// RollbackVersion rolls back to stable version
func (c *Client) RollbackVersion(ctx context.Context, appName string) error {
	return c.RollbackCanary(ctx, appName)
}
