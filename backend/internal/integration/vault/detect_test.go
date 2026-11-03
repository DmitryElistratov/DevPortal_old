package vault

import (
"context"
"encoding/json"
"net/http"
"net/http/httptest"
"strings"
"testing"
)

// fakeVault simulates a real Vault KV v2 mount at "secret/" honoring only the
// literal LIST verb and the data/metadata path layout.
func fakeVaultKV2(t *testing.T, mounted bool) *httptest.Server {
store := map[string]map[string]interface{}{}
return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
path := strings.TrimPrefix(r.URL.Path, "/v1/")
if path == "sys/health" {
json.NewEncoder(w).Encode(map[string]interface{}{"sealed": false})
return
}
if path == "sys/mounts" {
if !mounted {
json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"mounts": map[string]interface{}{}}})
return
}
json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"mounts": map[string]interface{}{
"secret/": map[string]interface{}{"type": "kv", "options": map[string]string{"version": "2"}},
}}})
return
}
if r.Method == "POST" && path == "sys/mounts/secret" {
mounted = true
w.WriteHeader(204)
return
}
if !mounted {
w.WriteHeader(404)
json.NewEncoder(w).Encode(map[string]interface{}{"errors": []string{}})
return
}
switch {
case strings.HasPrefix(path, "secret/data/") && (r.Method == "GET"):
key := strings.TrimPrefix(path, "secret/data/")
d, ok := store[key]
if !ok {
w.WriteHeader(404)
json.NewEncoder(w).Encode(map[string]interface{}{"errors": []string{}})
return
}
json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": d, "metadata": map[string]interface{}{"version": 1}}})
case strings.HasPrefix(path, "secret/data/") && (r.Method == "POST" || r.Method == "PUT"):
var body struct {
Data map[string]interface{} `json:"data"`
}
json.NewDecoder(r.Body).Decode(&body)
store[strings.TrimPrefix(path, "secret/data/")] = body.Data
w.WriteHeader(204)
case r.Method == "LIST" && strings.HasPrefix(path, "secret/metadata/"):
prefix := strings.TrimSuffix(strings.TrimPrefix(path, "secret/metadata/"), "/")
set := map[string]bool{}
for k := range store {
if prefix == "" {
first := strings.SplitN(k, "/", 2)[0]
if strings.Contains(k, "/") {
first += "/"
}
set[first] = true
} else if strings.HasPrefix(k, prefix+"/") {
rest := strings.TrimPrefix(k, prefix+"/")
first := strings.SplitN(rest, "/", 2)[0]
if strings.Contains(rest, "/") {
first += "/"
}
set[first] = true
}
}
keys := []string{}
for k := range set {
keys = append(keys, k)
}
if len(keys) == 0 {
w.WriteHeader(404)
json.NewEncoder(w).Encode(map[string]interface{}{"errors": []string{}})
return
}
json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"keys": keys}})
default:
w.WriteHeader(404)
json.NewEncoder(w).Encode(map[string]interface{}{"errors": []string{}})
}
}))
}

func TestDetectEngineViaSysMounts(t *testing.T) {
srv := fakeVaultKV2(t, true)
defer srv.Close()
c := NewClient(srv.URL, "tok")
e := c.detectEngine(context.Background())
if e.Mount != "secret" || e.Version != 2 {
t.Fatalf("expected secret/v2, got %+v", e)
}
}

func TestWriteListReadRoundTripFreshVault(t *testing.T) {
// Fresh dev vault: nothing mounted yet -> auto-mount + re-detect must heal it.
srv := fakeVaultKV2(t, false)
defer srv.Close()
c := NewClient(srv.URL, "tok")
ctx := context.Background()

if err := c.WriteSecret(ctx, "secret/my/database", map[string]interface{}{"user": "admin", "pass": "s3cret"}); err != nil {
t.Fatalf("write failed: %v", err)
}
keys, err := c.ListSecrets(ctx, "secret/my")
if err != nil {
t.Fatalf("list failed: %v", err)
}
found := false
for _, k := range keys {
if strings.TrimSuffix(k, "/") == "database" {
found = true
}
}
if !found {
t.Fatalf("listed keys %v do not contain database", keys)
}
sec, err := c.GetSecret(ctx, "secret/my/database")
if err != nil {
t.Fatalf("get failed: %v", err)
}
if sec.Data["pass"] != "s3cret" {
t.Fatalf("bad data: %+v", sec.Data)
}
}
