package graylog

import (
	"context"
	"encoding/base64"
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
	username   string
	password   string
	httpClient *http.Client
}

func NewClient(baseURL, username, password string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type Stream struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type SearchResponse struct {
	Messages []Message `json:"messages"`
	Total    int       `json:"total_results"`
}

type Message struct {
	Message map[string]interface{} `json:"message"`
	Index   string                 `json:"index"`
}

func (c *Client) getAuth() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(c.username+":"+c.password))
}

// TestConnection tests the connection to Graylog
func (c *Client) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/system/lbstatus", c.baseURL), nil)
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
		return fmt.Errorf("graylog API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListStreams returns list of streams
func (c *Client) ListStreams(ctx context.Context) ([]Stream, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/streams", c.baseURL), nil)
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
		return nil, fmt.Errorf("graylog API error: %d - %s", resp.StatusCode, string(body))
	}

	var result struct {
		Streams []Stream `json:"streams"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Streams, nil
}

// SearchLogs searches for log messages
func (c *Client) SearchLogs(ctx context.Context, query string, from, to time.Time, limit int) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("from", from.Format(time.RFC3339))
	params.Set("to", to.Format(time.RFC3339))
	params.Set("limit", fmt.Sprintf("%d", limit))

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/search/universal/absolute?%s", c.baseURL, params.Encode()), nil)
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
		return nil, fmt.Errorf("graylog API error: %d - %s", resp.StatusCode, string(body))
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetStream returns details of a specific stream
func (c *Client) GetStream(ctx context.Context, streamID string) (*Stream, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/streams/%s", c.baseURL, streamID), nil)
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
		return nil, fmt.Errorf("graylog API error: %d - %s", resp.StatusCode, string(body))
	}

	var stream Stream
	if err := json.NewDecoder(resp.Body).Decode(&stream); err != nil {
		return nil, err
	}

	return &stream, nil
}
