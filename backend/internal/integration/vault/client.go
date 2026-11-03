package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client

	// KV engine autodetection (lazy, cached, periodically re-validated)
	engineMu      sync.Mutex
	engine        kvEngine
	engineReady   bool
	engineChecked time.Time
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

// detectKVEngine probes the Vault server once (lazily, cached per client) to
// find where a KV secrets engine is actually mounted and which version it is
// (v1 vs v2). This fixes the persistent-404 class of bugs: dev-mode Vault in
// Docker does NOT persist /vault/file data across container restarts, so the
// hardcoded assumption "secret/ is mounted as KV v2" silently breaks after any
// `docker compose up` of a fresh vault container. Detection order:
//  1. sys/mounts (needs list/read on sys/) — authoritative;
//  2. write-probe at secret/ + kv_versions endpoint — works even when sys/
//     is not readable by the token.
//
// The chosen engine (mount+version) is then used to build every request path.
type kvEngine struct {
	Mount   string // e.g. "secret", "kv-v2", "cubbyhole"
	Version int    // 1 or 2
}

func (e kvEngine) readPath(rel string) string {
	if e.Version == 2 {
		return fmt.Sprintf("/v1/%s/data/%s", e.Mount, rel)
	}
	return fmt.Sprintf("/v1/%s/%s", e.Mount, rel)
}

func (e kvEngine) listPath(rel string) string {
	if e.Version == 2 {
		return fmt.Sprintf("/v1/%s/metadata/%s", e.Mount, rel)
	}
	return fmt.Sprintf("/v1/%s/%s", e.Mount, rel)
}

func (e kvEngine) deletePath(rel string) string {
	if e.Version == 2 {
		return fmt.Sprintf("/v1/%s/data/%s", e.Mount, rel)
	}
	return fmt.Sprintf("/v1/%s/%s", e.Mount, rel)
}

// resolveEngine returns the detected KV engine, running detection once and
// caching the result. On failure it falls back to the historical default
// (secret/, KV v2) so behaviour never regresses. The cached engine is
// re-validated with a cheap write probe: dev-mode Vault containers lose their
// /vault/file storage between restarts, which can leave sys/mounts reporting
// a KV mount that no longer accepts any writes — without re-validation every
// operation would keep failing with 404 forever.
func (c *Client) resolveEngine(ctx context.Context) kvEngine {
	c.engineMu.Lock()
	defer c.engineMu.Unlock()
	if !c.engineReady {
		c.engine = c.detectEngine(ctx)
		c.engineReady = true
		log.Printf("[Vault] using KV engine: mount=%q version=v%d", c.engine.Mount, c.engine.Version)
		return c.engine
	}
	// Re-validate cached engine periodically (not on every request).
	if time.Since(c.engineChecked) < 30*time.Second {
		return c.engine
	}
	if c.probeWriteable(ctx, c.engine) {
		c.engineChecked = time.Now()
		return c.engine
	}
	log.Printf("[Vault] cached KV engine %q/v%d no longer writable — remounting and re-detecting", c.engine.Mount, c.engine.Version)
	c.ensureKVSecretMount(ctx)
	c.engine = c.detectEngine(ctx)
	c.engineChecked = time.Now()
	log.Printf("[Vault] re-detected KV engine: mount=%q version=v%d", c.engine.Mount, c.engine.Version)
	return c.engine
}

func (c *Client) detectEngine(ctx context.Context) kvEngine {
	fallback := kvEngine{Mount: "secret", Version: 2}

	// 1) Authoritative: list mounts via sys/mounts.
	statusCode, body, err := c.doJSON(ctx, "GET", fmt.Sprintf("%s/v1/sys/mounts", c.baseURL), nil)
	if err == nil && statusCode == http.StatusOK {
		var result struct {
			Data struct {
				Mounts map[string]struct {
					Type    string `json:"type"`
					Options struct {
						Version string `json:"version"`
					} `json:"options"`
				} `json:"mounts"`
			} `json:"data"`
		}
		if json.Unmarshal(body, &result) == nil {
			var best *kvEngine
			for path, m := range result.Data.Mounts {
				if !strings.Contains(m.Type, "kv") {
					continue
				}
				ver := 2
				if m.Options.Version == "1" {
					ver = 1
				}
				e := kvEngine{Mount: strings.Trim(path, "/"), Version: ver}
				// prefer the conventional "secret" mount, otherwise first found
				if e.Mount == "secret" || best == nil {
					best = &e
				}
			}
			if best != nil {
				return *best
			}
		}
	}

	// 2) Write-probe fallback (no sys/ permissions needed): try KV v2 first.
	probeKey := "__portal_probe__"
	v2Payload := bytes.NewReader([]byte(`{"data":{"` + probeKey + `":"1"}}`))
	if st, _, err := c.doJSON(ctx, "POST", fmt.Sprintf("%s/v1/secret/data/%s", c.baseURL, probeKey), v2Payload); err == nil && (st == http.StatusOK || st == http.StatusNoContent || st == http.StatusCreated) {
		// confirm it really behaves as v2 (versions endpoint exists)
		if stv, _, _ := c.doJSON(ctx, "GET", fmt.Sprintf("%s/v1/secret/kv_versions/%s?list=true", c.baseURL, probeKey), nil); stv == http.StatusOK {
			return kvEngine{Mount: "secret", Version: 2}
		}
		// write under data/ succeeded but no versions -> some proxy semantics; still treat as v2
		return kvEngine{Mount: "secret", Version: 2}
	}
	// try KV v1-style direct write
	v1Payload := bytes.NewReader([]byte(`{"` + probeKey + `":"1"}`))
	if st, _, err := c.doJSON(ctx, "POST", fmt.Sprintf("%s/v1/secret/%s", c.baseURL, probeKey), v1Payload); err == nil && (st == http.StatusOK || st == http.StatusNoContent || st == http.StatusCreated) {
		return kvEngine{Mount: "secret", Version: 1}
	}

	// nothing reachable — caller may attempt auto-mount, keep defaults
	return fallback
}

func (c *Client) doJSON(ctx context.Context, method, rawURL string, body io.Reader) (int, []byte, error) {
	// Vault's HTTP API implements LIST via the literal "LIST" verb (the only
	// officially documented way; hashicorp/vault handles it in its own router).
	// Go's net/http supports arbitrary methods, so we send LIST directly.
	// NOTE: GET+?list=true and X-Vault-Method headers are NOT honored by real
	// Vault — earlier attempts using them produced persistent 404s.
	if method == "LIST" && !strings.HasSuffix(rawURL, "/") {
		rawURL += "/" // Vault requires a trailing slash for directory listings
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("X-Vault-Token", c.token)
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
	engine := c.resolveEngine(ctx)

	// Try the detected engine layout first, then the other version as fallback.
	primary := engine
	secondary := kvEngine{Mount: engine.Mount, Version: otherVersion(engine.Version)}

	statusCode, body, err := c.doJSON(ctx, "GET", c.baseURL+primary.readPath(rel), nil)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusOK && primary.Version == 2 {
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
	if statusCode == http.StatusOK && primary.Version == 1 {
		var flat map[string]interface{}
		if err := json.Unmarshal(body, &struct {
			Data *map[string]interface{} `json:"data"`
		}{Data: &flat}); err == nil && flat != nil && len(flat) > 0 {
			return &Secret{Data: flat}, nil
		}
	}

	// Fallback: the other KV version (covers stale/failed autodetection).
	statusCode, body, err = c.doJSON(ctx, "GET", c.baseURL+secondary.readPath(rel), nil)
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

// otherVersion flips KV engine version 1 <-> 2 (used for fallback probes).
func otherVersion(v int) int {
	if v == 1 {
		return 2
	}
	return 1
}

// WriteSecret writes a secret to Vault using the detected KV engine, falling
// back to the other version's layout on 404.
func (c *Client) WriteSecret(ctx context.Context, path string, data map[string]interface{}) error {
	rel := normalizePath(path)
	engine := c.resolveEngine(ctx)

	tryWrite := func(e kvEngine) (int, []byte, error) {
		var payloadBytes []byte
		if e.Version == 2 {
			payloadBytes, _ = json.Marshal(map[string]interface{}{"data": data})
		} else {
			payloadBytes, _ = json.Marshal(data)
		}
		return c.doJSON(ctx, "POST", c.baseURL+e.readPath(rel), bytes.NewReader(payloadBytes))
	}

	statusCode, body, err := tryWrite(engine)
	if err != nil {
		return err
	}
	if statusCode == http.StatusOK || statusCode == http.StatusNoContent || statusCode == http.StatusCreated {
		return nil
	}
	if statusCode == http.StatusNotFound {
		// Detection may be stale/wrong — try the other KV layout once.
		statusCode, body, err = tryWrite(kvEngine{Mount: engine.Mount, Version: otherVersion(engine.Version)})
		if err != nil {
			return err
		}
		if statusCode == http.StatusOK || statusCode == http.StatusNoContent || statusCode == http.StatusCreated {
			return nil
		}
		// Both layouts 404: the mount reported by sys/mounts is likely dead
		// (fresh dev Vault restart wiped /vault/file). Remount and retry once.
		log.Printf("[Vault] write 404 on both KV layouts for %q — attempting auto-mount of secret/", rel)
		c.engineMu.Lock()
		c.engineReady = false
		c.engineMu.Unlock()
		if c.ensureKVSecretMount(ctx) {
			newEngine := c.resolveEngine(ctx)
			statusCode, body, err = tryWrite(newEngine)
			if err != nil {
				return err
			}
			if statusCode == http.StatusOK || statusCode == http.StatusNoContent || statusCode == http.StatusCreated {
				return nil
			}
			if altSt, altBody, altErr := tryWrite(kvEngine{Mount: newEngine.Mount, Version: otherVersion(newEngine.Version)}); altErr == nil &&
				(altSt == http.StatusOK || altSt == http.StatusNoContent || altSt == http.StatusCreated) {
				return nil
			} else if altErr != nil {
				return altErr
			} else {
				statusCode, body = altSt, altBody
			}
		}
	}
	return fmt.Errorf("vault API error: %d - %s", statusCode, string(body))
}

// DeleteSecret deletes a secret from Vault (soft-delete metadata in v2, hard delete in v1).
func (c *Client) DeleteSecret(ctx context.Context, path string) error {
	rel := normalizePath(path)
	engine := c.resolveEngine(ctx)

	statusCode, body, err := c.doJSON(ctx, "DELETE", c.baseURL+engine.deletePath(rel), nil)
	if err != nil {
		return err
	}
	if statusCode == http.StatusOK || statusCode == http.StatusNoContent {
		return nil
	}
	if statusCode == http.StatusNotFound {
		statusCode, body, err = c.doJSON(ctx, "DELETE", c.baseURL+kvEngine{Mount: engine.Mount, Version: otherVersion(engine.Version)}.deletePath(rel), nil)
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
	engine := c.resolveEngine(ctx)
	// KV v2 version history lives in the "ca_metadata" (current) and
	// "metadata" (all versions incl. deleted) sections of /secret/metadata/<path>.
	statusCode, body, err := c.doJSON(ctx, "GET", c.baseURL+fmt.Sprintf("/v1/%s/metadata/%s", engine.Mount, rel), nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("vault API error: %d - %s", statusCode, string(body))
	}
	var result struct {
		Data struct {
			Metadata struct {
				Versions []map[string]interface{} `json:"versions"`
			} `json:"metadata"`
			CAMetadata struct {
				Versions []map[string]interface{} `json:"versions"`
			} `json:"cam_metadata"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	versions := result.Data.Metadata.Versions
	if len(versions) == 0 {
		versions = result.Data.CAMetadata.Versions
	}
	return versions, nil
}

// RestoreVersion restores a previously deleted KV v2 version of a secret.
func (c *Client) RestoreVersion(ctx context.Context, path string, version int) error {
	rel := normalizePath(path)
	engine := c.resolveEngine(ctx)
	payload, err := json.Marshal(map[string]interface{}{"versions": []int{version}})
	if err != nil {
		return err
	}
	statusCode, body, err := c.doJSON(ctx, "POST", c.baseURL+fmt.Sprintf("/v1/%s/delete/%s", engine.Mount, rel), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	if statusCode == http.StatusOK || statusCode == http.StatusNoContent {
		return nil
	}
	return fmt.Errorf("vault API error: %d - %s", statusCode, string(body))
}

// diagnoseMounts fetches /v1/sys/mounts and logs all mounted secrets engines.
// Used when a LIST fails with 404 so the backend log shows exactly what is
// mounted in Vault (mount path + type/options) versus the hardcoded "secret/"
// prefix our client uses — this pinpoints misconfigured integration URLs or
// non-default mounts instead of a bare 404.
func (c *Client) diagnoseMounts(ctx context.Context) {
	statusCode, body, err := c.doJSON(ctx, "GET", fmt.Sprintf("%s/v1/sys/mounts", c.baseURL), nil)
	if err != nil {
		log.Printf("[Vault] diagnostics: sys/mounts request failed: %v", err)
		return
	}
	if statusCode != http.StatusOK {
		log.Printf("[Vault] diagnostics: sys/mounts -> %d %.300s (token may lack read access to sys/)", statusCode, string(body))
		return
	}
	var result struct {
		Data struct {
			Mounts map[string]struct {
				Type    string `json:"type"`
				Options struct {
					Version string `json:"version"`
				} `json:"options"`
			} `json:"mounts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[Vault] diagnostics: cannot parse sys/mounts: %v", err)
		return
	}
	for path, m := range result.Data.Mounts {
		if strings.Contains(m.Type, "kv") {
			log.Printf("[Vault] diagnostics: KV engine mounted at %q (type=%s, version=%s)", path, m.Type, m.Options.Version)
		}
	}
}

// ensureKVSecretMount attempts to enable a KV v2 secrets engine at the
// default "secret/" mount. This makes the portal self-healing: if Vault was
// started without a mounted KV store (or with an empty file/raft backend),
// the first LIST that returns 404 triggers one-time setup and retry.
func (c *Client) ensureKVSecretMount(ctx context.Context) bool {
	payload := bytes.NewReader([]byte(`{"type":"kv","options":{"version":"2"}}`))
	statusCode, body, err := c.doJSON(ctx, "POST", fmt.Sprintf("%s/v1/sys/mounts/secret", c.baseURL), payload)
	if err != nil {
		log.Printf("[Vault] auto-mount request failed: %v", err)
		return false
	}
	if statusCode == http.StatusOK || statusCode == http.StatusNoContent || statusCode == http.StatusCreated {
		log.Printf("[Vault] KV v2 secrets engine auto-mounted at secret/")
		return true
	}
	// path already in use -> mount exists, nothing to do
	if statusCode == http.StatusBadRequest && strings.Contains(string(body), "path is already in use") {
		return false
	}
	log.Printf("[Vault] auto-mount KV v2 returned %d - %.200s", statusCode, string(body))
	return false
}

// probeWriteable verifies that the detected KV engine actually accepts writes
// at <mount>/data/probe (KV v2). Fresh dev-mode Vault containers lose their
// /vault/file storage on restart, so sys/mounts can report a stale KV v2
// entry while every data write answers 404. When this happens we re-run the
// auto-mount (which Vault rejects as "already in use" only if the mount truly
// works) and force re-detection. Returns true when writes are OK.
func (c *Client) probeWriteable(ctx context.Context, e kvEngine) bool {
	probeKey := "__portal_write_probe__"
	payload := bytes.NewReader([]byte(`{"data":{"` + probeKey + `":"1"}}`))
	statusCode, _, err := c.doJSON(ctx, "POST", c.baseURL+e.readPath(probeKey), payload)
	if err != nil {
		return false
	}
	return statusCode == http.StatusOK || statusCode == http.StatusNoContent || statusCode == http.StatusCreated
}

// ListSecrets lists secret keys at a path, supporting KV v2 (metadata endpoint)
// and KV v1 (LIST verb on the path itself). A 404 on an empty/nonexistent
// directory is treated as an empty list rather than an error. If both LIST
// attempts return 404 for the root listing, we try to auto-enable a KV v2
// engine once and retry — this fixes the common "fresh dev Vault" case.
func (c *Client) ListSecrets(ctx context.Context, path string) ([]string, error) {
	rel := normalizePath(path)
	if rel != "" && !strings.HasSuffix(rel, "/") {
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

	listOnce := func() (keys []string, found bool, notFound bool) {
		engine := c.resolveEngine(ctx)
		// KV v2: LIST /v1/<mount>/metadata/<path>; KV v1: LIST /v1/<mount>/<path>.
		listURL := fmt.Sprintf("%s/v1/%s/metadata/%s", c.baseURL, engine.Mount, rel)
		if engine.Version == 1 {
			listURL = fmt.Sprintf("%s/v1/%s/%s", c.baseURL, engine.Mount, strings.TrimSuffix(rel, "/"))
		}
		status, body, err := c.doJSON(ctx, "LIST", listURL, nil)
		if err != nil {
			log.Printf("[Vault] LIST %q transport error: %v", listURL, err)
			return nil, false, false
		}
		log.Printf("[Vault] LIST %q -> %d %.200s", listURL, status, string(body))
		if status == http.StatusOK {
			if k, ok := extractKeys(body); ok && k != nil {
				return k, true, false
			}
		}
		// Fallback: the other KV layout on the same mount (covers stale detection).
		altURL := fmt.Sprintf("%s/v1/%s/%s", c.baseURL, engine.Mount, strings.TrimSuffix(rel, "/"))
		if engine.Version == 1 {
			altURL = fmt.Sprintf("%s/v1/%s/metadata/%s", c.baseURL, engine.Mount, rel)
		}
		altStatus, altBody, err := c.doJSON(ctx, "LIST", altURL, nil)
		if err != nil {
			log.Printf("[Vault] LIST fallback %q transport error: %v", altURL, err)
			return nil, false, false
		}
		log.Printf("[Vault] LIST fallback %q -> %d %.200s", altURL, altStatus, string(altBody))
		if altStatus == http.StatusOK {
			if k, ok := extractKeys(altBody); ok && k != nil {
				return k, true, false
			}
		}
		nf := status == http.StatusNotFound && altStatus == http.StatusNotFound
		return nil, false, nf
	}

	keys, found, notFound := listOnce()
	if found {
		return keys, nil
	}

	// Root listing that 404s on both engines usually means no KV mount exists.
	if notFound && rel == "" && c.ensureKVSecretMount(ctx) {
		// re-detect with the fresh mount
		c.engineMu.Lock()
		c.engineReady = false
		c.engineMu.Unlock()
		if keys, found, _ = listOnce(); found {
			return keys, nil
		}
	}

	// Still nothing — log what is actually mounted in Vault so the operator can
	// see whether the engine lives under a non-default mount path.
	if !found {
		c.diagnoseMounts(ctx)
	}

	// Empty/nonexistent directory -> empty list, not an error.
	// Real Vault answers 404 {"errors":[]} for listing a missing path and also
	// for a path that holds a leaf secret rather than a directory.
	if notFound {
		return []string{}, nil
	}
	return nil, fmt.Errorf("vault LIST failed for path %q", rel)
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
