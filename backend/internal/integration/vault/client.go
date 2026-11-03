package vault

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

type Secret struct {
	Data     map[string]interface{} `json:"data"`
	Metadata SecretMetadata         `json:"metadata"`
}

type SecretMetadata struct {
	CreatedTime  string `json:"created_time"`
	UpdatedTime  string `json:"updated_time"`
	Version      int    `json:"version"`
	DeletionTime string `json:"deletion_time"`
	Destroyed    bool   `json:"destroyed"`
}

// TestConnection tests the connection to Vault
func (c *Client) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/sys/health", c.baseURL), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// normalizePath strips the KV mount prefix so callers can pass either
// "secret/data/foo" or "foo" and we build correct paths for both KV v1 and v2.
func normalizePath(path string) string {
	p := strings.TrimPrefix(path, "/")
	for _, prefix := range []string{"data/", "metadata/", "delete/"} {
		p = strings.TrimPrefix(p, prefix)
	}
	// strip mount prefix if the whole path starts with the default mount
	if p == "secret" || strings.HasPrefix(p, "secret/") {
		p = strings.TrimPrefix(strings.TrimPrefix(p, "secret"), "/")
	}
	return p
}

func (c *Client) doJSON(ctx context.Context, method, url string, body io.Reader) (int, []byte, error) {
	// Vault's HTTP API implements LIST as GET with the X-Vault-Method header;
	// a literal "LIST" verb is not valid HTTP and fails against real Vault.
	vaultList := false
	if method == "LIST" {
		vaultList = true
		method = http.MethodGet
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("X-Vault-Token", c.token)
	if vaultList {
		req.Header.Set("X-Vault-Method", "LIST")
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data, nil
}

// GetSecret retrieves a secret from Vault, supporting both KV v1 and v2 engines.
func (c *Client) GetSecret(ctx context.Context, path string) (*Secret, error) {
	rel := normalizePath(path)

	// Try KV v2 first: /v1/<mount>/data/<path>
	statusCode, body, err := c.doJSON(ctx, "GET", fmt.Sprintf("%s/v1/secret/data/%s", c.baseURL, rel), nil)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusOK {
		var result struct {
			Data Secret `json:"data"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
		// KV v2 response nests the payload under data.data; if that inner map is
		// missing (or empty) fall through to the KV v1 shape below.
		if len(result.Data.Data) > 0 {
			return &result.Data, nil
		}
	}

	// Fallback to KV v1: /v1/<mount>/<path>
	statusCode, body, err = c.doJSON(ctx, "GET", fmt.Sprintf("%s/v1/secret/%s", c.baseURL, rel), nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("vault API error: %d - %s", statusCode, string(body))
	}

	// KV v1 returns the key-value map directly under "data"
	var flat map[string]interface{}
	if err := json.Unmarshal(body, &struct {
		Data *map[string]interface{} `json:"data"`
	}{Data: &flat}); err != nil {
		return nil, err
	}
	if flat == nil {
		return nil, fmt.Errorf("vault returned empty secret data")
	}
	return &Secret{Data: flat}, nil
}

// WriteSecret writes a secret to Vault (KV v2 semantics; falls back to v1 on 404).
func (c *Client) WriteSecret(ctx context.Context, path string, data map[string]interface{}) error {
	rel := normalizePath(path)

	v2Payload, err := json.Marshal(map[string]interface{}{"data": data})
	if err != nil {
		return err
	}
	statusCode, body, err := c.doJSON(ctx, "POST", fmt.Sprintf("%s/v1/secret/data/%s", c.baseURL, rel), bytes.NewReader(v2Payload))
	if err != nil {
		return err
	}
	if statusCode == http.StatusOK || statusCode == http.StatusNoContent || statusCode == http.StatusCreated {
		return nil
	}
	if statusCode == http.StatusNotFound {
		// Fallback: KV v1 engine stores payload directly under the path
		v1Payload, err := json.Marshal(data)
		if err != nil {
			return err
		}
		statusCode, body, err = c.doJSON(ctx, "POST", fmt.Sprintf("%s/v1/secret/%s", c.baseURL, rel), bytes.NewReader(v1Payload))
		if err != nil {
			return err
		}
		if statusCode == http.StatusOK || statusCode == http.StatusNoContent || statusCode == http.StatusCreated {
			return nil
		}
	}
	return fmt.Errorf("vault API error: %d - %s", statusCode, string(body))
}

// DeleteSecret deletes a secret from Vault (soft-delete metadata in v2, hard delete in v1).
func (c *Client) DeleteSecret(ctx context.Context, path string) error {
	rel := normalizePath(path)

	statusCode, body, err := c.doJSON(ctx, "DELETE", fmt.Sprintf("%s/v1/secret/data/%s", c.baseURL, rel), nil)
	if err != nil {
		return err
	}
	if statusCode == http.StatusOK || statusCode == http.StatusNoContent {
		return nil
	}
	if statusCode == http.StatusNotFound {
		statusCode, body, err = c.doJSON(ctx, "DELETE", fmt.Sprintf("%s/v1/secret/%s", c.baseURL, rel), nil)
		if err != nil {
			return err
		}
		if statusCode == http.StatusOK || statusCode == http.StatusNoContent {
			return nil
		}
	}
	return fmt.Errorf("vault API error: %d - %s", statusCode, string(body))
}

// ListVersions returns KV v2 version metadata for a secret path.
func (c *Client) ListVersions(ctx context.Context, path string) ([]map[string]interface{}, error) {
	rel := normalizePath(path)
	statusCode, body, err := c.doJSON(ctx, "GET", fmt.Sprintf("%s/v1/secret/metadata/%s?list=true", c.baseURL, rel), nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("vault API error: %d - %s", statusCode, string(body))
	}
	var result struct {
		Data struct {
			Versions []map[string]interface{} `json:"versions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Data.Versions, nil
}

// RestoreVersion restores a previously deleted KV v2 version of a secret.
func (c *Client) RestoreVersion(ctx context.Context, path string, version int) error {
	rel := normalizePath(path)
	payload, err := json.Marshal(map[string]interface{}{"versions": []int{version}})
	if err != nil {
		return err
	}
	statusCode, body, err := c.doJSON(ctx, "POST", fmt.Sprintf("%s/v1/secret/delete/%s", c.baseURL, rel), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	if statusCode == http.StatusOK || statusCode == http.StatusNoContent {
		return nil
	}
	return fmt.Errorf("vault API error: %d - %s", statusCode, string(body))
}

// ListSecrets lists secret keys at a path, supporting KV v2 (metadata endpoint)
// and KV v1 (LIST verb on the path itself). A 404 on an empty/nonexistent
// directory is treated as an empty list rather than an error.
func (c *Client) ListSecrets(ctx context.Context, path string) ([]string, error) {
	rel := normalizePath(path)
	if rel == "" {
		rel = ""
	} else if !strings.HasSuffix(rel, "/") {
		rel += "/"
	}

	extractKeys := func(body []byte) ([]string, bool) {
		var result struct {
			Data struct {
				Keys []string `json:"keys"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, false
		}
		return result.Data.Keys, true
	}

	// KV v2: LIST /v1/secret/metadata/<path> (doJSON translates LIST into
	// GET + X-Vault-Method: LIST, which is how Vault's HTTP API implements it)
	statusCode, body, err := c.doJSON(ctx, "LIST", fmt.Sprintf("%s/v1/secret/metadata/%s", c.baseURL, rel), nil)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusOK {
		if keys, ok := extractKeys(body); ok && keys != nil {
			return keys, nil
		}
	}

	// KV v1 fallback: LIST /v1/secret/<path>
	statusCode, body, err = c.doJSON(ctx, "LIST", fmt.Sprintf("%s/v1/secret/%s", c.baseURL, rel), nil)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusOK {
		if keys, ok := extractKeys(body); ok && keys != nil {
			return keys, nil
		}
	}

	// Empty/nonexistent directory -> empty list, not an error.
	// Real Vault answers 404 {"errors":[]} for listing a missing path and also
	// for a path that holds a leaf secret rather than a directory.
	if statusCode == http.StatusNotFound {
		return []string{}, nil
	}
	return nil, fmt.Errorf("vault API error: %d - %s", statusCode, string(body))
}

// Health checks Vault health
func (c *Client) Health(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/sys/health", c.baseURL), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
