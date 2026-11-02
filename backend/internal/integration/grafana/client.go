package grafana

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type Dashboard struct {
	ID    int    `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Datasource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// TestConnection tests the connection to Grafana
func (c *Client) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/health", c.baseURL), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("grafana API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListDashboards returns list of dashboards
func (c *Client) ListDashboards(ctx context.Context) ([]Dashboard, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/search", c.baseURL), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("grafana API error: %d - %s", resp.StatusCode, string(body))
	}

	var dashboards []Dashboard
	if err := json.NewDecoder(resp.Body).Decode(&dashboards); err != nil {
		return nil, err
	}

	return dashboards, nil
}

// GetDashboard returns details of a specific dashboard
func (c *Client) GetDashboard(ctx context.Context, uid string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/dashboards/uid/%s", c.baseURL, uid), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("grafana API error: %d - %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// ListDatasources returns list of datasources
func (c *Client) ListDatasources(ctx context.Context) ([]Datasource, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/datasources", c.baseURL), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("grafana API error: %d - %s", resp.StatusCode, string(body))
	}

	var datasources []Datasource
	if err := json.NewDecoder(resp.Body).Decode(&datasources); err != nil {
		return nil, err
	}

	return datasources, nil
}

type MetricQuery struct {
	Expr         string `json:"expr"`
	LegendFormat string `json:"legendFormat,omitempty"`
}

type MetricResponse struct {
	Status string          `json:"status"`
	Data   MetricQueryData `json:"data"`
}

type MetricQueryData struct {
	ResultType string        `json:"resultType"`
	Result     []MetricValue `json:"result"`
}

type MetricValue struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}

// QueryMetrics executes a PromQL query
func (c *Client) QueryMetrics(ctx context.Context, query string) (*MetricResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/datasources/proxy/1/api/v1/query?query=%s", c.baseURL, query), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("grafana API error: %d - %s", resp.StatusCode, string(body))
	}

	var result MetricResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
