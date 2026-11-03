package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Simulates a Vault dev-server with KV v1 mounted at "secret/" (the exact
// situation from the user's logs: metadata 404 {"errors":[]} + v1 path 404
// "Invalid path for a versioned K/V secrets engine").
func TestListSecretsKVv1Only(t *testing.T) {
	mux := http.NewServeMux()
	// KV v2 metadata endpoint does not exist -> real Vault style 404
	mux.HandleFunc("/v1/secret/metadata/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"errors":[]}`))
	})
	// KV v1 LIST: only valid when GET carries list=true or X-Vault-Method: LIST
	mux.HandleFunc("/v1/secret/", func(w http.ResponseWriter, r *http.Request) {
		isList := r.URL.Query().Get("list") == "true" || r.Header.Get("X-Vault-Method") == "LIST"
		if !isList {
			w.WriteHeader(404)
			w.Write([]byte(`{"errors":["Invalid path for a versioned K/V secrets engine"]}`))
			return
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"keys": []string{"app-db", "api-key"}},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	keys, err := c.ListSecrets(context.Background(), "secret")
	if err != nil {
		t.Fatalf("ListSecrets failed: %v", err)
	}
	if len(keys) != 2 || keys[0] != "app-db" {
		t.Fatalf("expected [app-db api-key], got %v", keys)
	}
}

// Fresh Vault without any KV mount: root LIST 404s everywhere; auto-mount must
// fire and retry must succeed.
func TestAutoMountOnEmptyVault(t *testing.T) {
	mounted := false
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/sys/mounts/secret", func(w http.ResponseWriter, r *http.Request) {
		mounted = true
		w.WriteHeader(204)
	})
	mux.HandleFunc("/v1/secret/metadata/", func(w http.ResponseWriter, r *http.Request) {
		if !mounted {
			w.WriteHeader(404)
			w.Write([]byte(`{"errors":[]}`))
			return
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"keys": []string{"created-secret"}},
		})
	})
	mux.HandleFunc("/v1/secret/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"errors":["Invalid path for a versioned K/V secrets engine"]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	keys, err := c.ListSecrets(context.Background(), "")
	if err != nil {
		t.Fatalf("ListSecrets failed: %v", err)
	}
	if !mounted {
		t.Fatal("auto-mount was not attempted")
	}
	if len(keys) != 1 || keys[0] != "created-secret" {
		t.Fatalf("expected [created-secret], got %v", keys)
	}
}
