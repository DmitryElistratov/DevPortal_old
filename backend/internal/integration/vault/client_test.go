package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeVault emulates a real Vault KV v2 engine closely enough for the client:
// - LIST is GET + X-Vault-Method: LIST; unknown/leaf paths -> 404 {"errors":[]}
// - writes to /v1/secret/data/<path> store under data key
// - reads from /v1/secret/data/<path> return {data:{data:{...},metadata:{...}}}
type fakeVault struct {
	secrets map[string]map[string]interface{}
}

func (f *fakeVault) handler() http.Handler {
	mux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		isList := r.Header.Get("X-Vault-Method") == "LIST" || r.Method == "LIST"

		switch {
		case strings.HasPrefix(p, "/v1/secret/metadata/"):
			dir := strings.TrimPrefix(p, "/v1/secret/metadata/")
			dir = strings.TrimSuffix(dir, "/")
			var keys []string
			for k := range f.secrets {
				if dir == "" {
					base := strings.SplitN(k, "/", 2)
					if len(base) == 1 {
						keys = append(keys, base[0])
					} else {
						d := base[0] + "/"
						found := false
						for _, x := range keys {
							if x == d {
								found = true
								break
							}
						}
						if !found {
							keys = append(keys, d)
						}
					}
					continue
				}
				if strings.HasPrefix(k, dir+"/") {
					rest := strings.TrimPrefix(k, dir+"/")
					base := strings.SplitN(rest, "/", 2)
					name := base[0]
					if len(base) == 2 {
						name += "/"
					}
					found := false
					for _, x := range keys {
						if x == name {
							found = true
							break
						}
					}
					if !found {
						keys = append(keys, name)
					}
				}
			}
			if keys == nil {
				keys = []string{}
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"keys": keys}})
		case isList && strings.HasPrefix(p, "/v1/secret/"):
			// KV v1 style list on same path — emulate 404 like real Vault when engine is v2
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{"errors": []string{}})
		case strings.HasPrefix(p, "/v1/secret/data/"):
			path := strings.TrimPrefix(p, "/v1/secret/data/")
			switch r.Method {
			case http.MethodGet:
				data, ok := f.secrets[path]
				if !ok {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]interface{}{"errors": []string{}})
					return
				}
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"data":     data,
						"metadata": map[string]interface{}{"version": 1, "created_time": "2026-11-03T00:00:00Z"},
					},
				})
			case http.MethodPost, http.MethodPut:
				var req struct {
					Data map[string]interface{} `json:"data"`
				}
				json.NewDecoder(r.Body).Decode(&req)
				f.secrets[path] = req.Data
				json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"version": 1}})
			case http.MethodDelete:
				delete(f.secrets, path)
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{"errors": []string{"no handler"}})
		}
	})
	return mux
}

func TestWriteListGetRoundTrip(t *testing.T) {
	f := &fakeVault{secrets: map[string]map[string]interface{}{}}
	srv := httptest.NewServer(f.handler())
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	ctx := context.Background()

	// 1. Write nested secret like GUI does: path "secret/my/database"
	if err := c.WriteSecret(ctx, "secret/my/database", map[string]interface{}{"user": "admin", "pass": "s3cret"}); err != nil {
		t.Fatalf("WriteSecret failed: %v", err)
	}

	// 2. List at root "secret" must include "my/" directory
	keys, err := c.ListSecrets(ctx, "secret")
	if err != nil {
		t.Fatalf("ListSecrets(secret) failed: %v", err)
	}
	found := false
	for _, k := range keys {
		if k == "my/" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 'my/' in keys, got %v", keys)
	}

	// 3. List at "secret/my" must include leaf "database"
	keys, err = c.ListSecrets(ctx, "secret/my")
	if err != nil {
		t.Fatalf("ListSecrets(secret/my) failed: %v", err)
	}
	found = false
	for _, k := range keys {
		if k == "database" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 'database' in keys, got %v", keys)
	}

	// 4. Get the secret back with data
	sec, err := c.GetSecret(ctx, "secret/my/database")
	if err != nil {
		t.Fatalf("GetSecret failed: %v", err)
	}
	if sec.Data["user"] != "admin" || sec.Data["pass"] != "s3cret" {
		t.Fatalf("unexpected secret data: %+v", sec.Data)
	}

	// 5. Listing a LEAF path must not error out (real Vault returns 404 there)
	if _, err := c.ListSecrets(ctx, "secret/my/database/nonexistent"); err != nil {
		t.Fatalf("listing missing path should return empty list, got error: %v", err)
	}
}
