package vault

import (
"context"
"encoding/json"
"net/http"
"net/http/httptest"
"strings"
"testing"
)

// fakeVault emulates a real Vault KV v2 engine: LIST is answered only when the
// request carries X-Vault-Method: LIST or X-Vault-List: true headers (file/raft
// storage behaviour). Literal "LIST" verb returns 405 like on nginx-fronted Vault.
func TestE2E_WriteThenListGet(t *testing.T) {
store := map[string]map[string]interface{}{}
mux := http.NewServeMux()
mux.HandleFunc("/v1/sys/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
mux.HandleFunc("/v1/secret/data/", func(w http.ResponseWriter, r *http.Request) {
p := strings.TrimPrefix(r.URL.Path, "/v1/secret/data/")
switch r.Method {
case http.MethodPost:
var body struct{ Data map[string]interface{} `json:"data"` }
json.NewDecoder(r.Body).Decode(&body)
store[p] = body.Data
w.WriteHeader(200)
w.Write([]byte(`{"data":{"created_version":1}}`))
case http.MethodGet:
d, ok := store[p]
if !ok {
w.WriteHeader(404)
w.Write([]byte(`{"errors":[]}`))
return
}
json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": d, "metadata": map[string]interface{}{"version": 1}}})
}
})
mux.HandleFunc("/v1/secret/metadata/", func(w http.ResponseWriter, r *http.Request) {
if r.Header.Get("X-Vault-Method") != "LIST" && r.Header.Get("X-Vault-List") != "true" {
w.WriteHeader(405)
w.Write([]byte(`{"errors":["method not allowed"]}`))
return
}
dir := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/secret/metadata/"), "/")
set := map[string]bool{}
for k := range store {
if dir == "" {
first := strings.SplitN(k, "/", 2)[0]
if strings.Contains(k[len(first)+1:], "/") {
set[first + "/"] = true
} else {
set[first] = true
}
continue
}
if strings.HasPrefix(k, dir+"/") {
rest := k[len(dir)+1:]
first := strings.SplitN(rest, "/", 2)[0]
if strings.Contains(rest[len(first):], "/") {
set[first+"/"] = true
} else {
set[first] = true
}
}
}
keys := []string{}
for s := range set {
keys = append(keys, s)
}
json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"keys": keys}})
})
srv := httptest.NewServer(mux)
defer srv.Close()

c := NewClient(srv.URL, "tok")
ctx := context.Background()
if err := c.WriteSecret(ctx, "secret/my/database", map[string]interface{}{"user": "admin"}); err != nil {
t.Fatalf("write: %v", err)
}
keys, err := c.ListSecrets(ctx, "secret/my")
if err != nil {
t.Fatalf("list: %v", err)
}
t.Logf("list(secret/my) = %v", keys)
found := false
for _, k := range keys {
if strings.TrimRight(k, "/") == "database" {
found = true
}
}
if !found {
t.Fatalf("expected database in %v", keys)
}
sec, err := c.GetSecret(ctx, "secret/my/database")
if err != nil {
t.Fatalf("get: %v", err)
}
if sec.Data["user"] != "admin" {
t.Fatalf("bad data %v", sec.Data)
}
}
