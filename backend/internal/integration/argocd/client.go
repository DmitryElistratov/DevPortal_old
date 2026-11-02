package argocd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
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

// CreateCanary creates a canary deployment with traffic splitting
func (c *Client) CreateCanary(ctx context.Context, appName, version string, percentage int) error {
	payload := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{
				"canary.argoproj.io/traffic-split": fmt.Sprintf("%d", percentage),
				"canary.argoproj.io/mode":          "canary",
			},
		},
		"spec": map[string]interface{}{
			"source": map[string]interface{}{
				"targetRevision": version,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/api/v1/applications/%s", c.baseURL, appName), bytes.NewReader(body))
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

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("argocd API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// UpdateCanaryTraffic updates traffic percentage for canary deployment
func (c *Client) UpdateCanaryTraffic(ctx context.Context, appName string, percentage int) error {
	payload := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{
				"canary.argoproj.io/traffic-split": fmt.Sprintf("%d", percentage),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/api/v1/applications/%s", c.baseURL, appName), bytes.NewReader(body))
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

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("argocd API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// PromoteCanary promotes canary to 100% traffic
func (c *Client) PromoteCanary(ctx context.Context, appName string) error {
	return c.UpdateCanaryTraffic(ctx, appName, 100)
}

// RollbackCanary rolls back canary to 0% traffic
func (c *Client) RollbackCanary(ctx context.Context, appName string) error {
	payload := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{
				"canary.argoproj.io/traffic-split": "0",
				"canary.argoproj.io/mode":          "stable",
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/api/v1/applications/%s", c.baseURL, appName), bytes.NewReader(body))
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

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("argocd API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
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
