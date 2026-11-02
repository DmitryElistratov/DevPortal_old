package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	apiServer  string
	token      string
	httpClient *http.Client
}

func NewClient(apiServer, token string) *Client {
	return &Client{
		apiServer:  apiServer,
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type Pod struct {
	Metadata Metadata `json:"metadata"`
	Status   PodStatus `json:"status"`
}

type Metadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type PodStatus struct {
	Phase string `json:"phase"`
}

type Deployment struct {
	Metadata Metadata         `json:"metadata"`
	Spec     DeploymentSpec   `json:"spec"`
	Status   DeploymentStatus `json:"status"`
}

type DeploymentSpec struct {
	Replicas int `json:"replicas"`
}

type DeploymentStatus struct {
	ReadyReplicas int `json:"readyReplicas"`
}

type Service struct {
	Metadata Metadata    `json:"metadata"`
	Spec     ServiceSpec `json:"spec"`
}

type ServiceSpec struct {
	Type      string `json:"type"`
	ClusterIP string `json:"clusterIP"`
}

// TestConnection tests the connection to Kubernetes
func (c *Client) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api", c.apiServer), nil)
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
		return fmt.Errorf("kubernetes API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListPods returns list of pods in a namespace
func (c *Client) ListPods(ctx context.Context, namespace string) ([]Pod, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/namespaces/%s/pods", c.apiServer, namespace), nil)
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
		return nil, fmt.Errorf("kubernetes API error: %d - %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []Pod `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Items, nil
}

// ListDeployments returns list of deployments in a namespace
func (c *Client) ListDeployments(ctx context.Context, namespace string) ([]Deployment, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/apis/apps/v1/namespaces/%s/deployments", c.apiServer, namespace), nil)
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
		return nil, fmt.Errorf("kubernetes API error: %d - %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []Deployment `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Items, nil
}

// ListServices returns list of services in a namespace
func (c *Client) ListServices(ctx context.Context, namespace string) ([]Service, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/namespaces/%s/services", c.apiServer, namespace), nil)
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
		return nil, fmt.Errorf("kubernetes API error: %d - %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []Service `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Items, nil
}

// GetPodLogs returns logs from a pod
func (c *Client) GetPodLogs(ctx context.Context, namespace, podName string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/log", c.apiServer, namespace, podName), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("kubernetes API error: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
