package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/company/release-portal/internal/integration/vault"
	"github.com/company/release-portal/internal/repository"
)

type VaultService struct {
	vaultClient *vault.Client
	integration *repository.IntegrationRepository
}

func NewVaultService(vaultClient *vault.Client) *VaultService {
	return &VaultService{vaultClient: vaultClient}
}

// WithIntegrationRepo enables lazy re-initialization of the Vault client from
// the DB. Without it, an integration created/updated after backend startup
// would leave vaultClient nil and every request failed with
// "vault client not configured" (HTTP 500).
func (s *VaultService) WithIntegrationRepo(repo *repository.IntegrationRepository) {
	s.integration = repo
}

// cachedClient remembers URL+token so we only rebuild the client when the
// integration config actually changed (prevents stale clients after edits).
type cachedClient struct {
	url   string
	token string
}

var lastConfig *cachedClient

// ensureClient returns a usable client, building one from the enabled "vault"
// integration in the database if needed. The client is re-created whenever the
// stored integration URL/token changes, so edits in Settings take effect
// without restarting the backend.
func (s *VaultService) ensureClient() (*vault.Client, error) {
	if s.integration == nil {
		if s.vaultClient != nil {
			return s.vaultClient, nil
		}
		return nil, fmt.Errorf("vault client not configured")
	}
	integrations, err := s.integration.GetEnabledByType("vault")
	if err != nil || len(integrations) == 0 {
		if s.vaultClient != nil && lastConfig == nil {
			// fall back to startup-constructed client only if DB read failed
			return s.vaultClient, nil
		}
		return nil, fmt.Errorf("no enabled Vault integration found; configure it in Settings")
	}
	it := integrations[0]
	token, _ := it.Config["token"].(string)
	if token == "" {
		token, _ = it.Config["password"].(string)
	}
	if token == "" {
		return nil, fmt.Errorf("Vault integration %q has no token configured", it.Name)
	}
	url := strings.TrimRight(it.URL, "/")
	if url == "" {
		url, _ = it.Config["url"].(string)
		url = strings.TrimRight(url, "/")
	}
	if url == "" {
		return nil, fmt.Errorf("Vault integration %q has no URL configured", it.Name)
	}
	// Rebuild only when config changed or client missing.
	if s.vaultClient == nil || lastConfig == nil || lastConfig.url != url || lastConfig.token != token {
		s.vaultClient = vault.NewClient(url, token)
		lastConfig = &cachedClient{url: url, token: token}
		log.Printf("[Vault] Client (re)initialized from integration %q (%s)", it.Name, url)
	}
	return s.vaultClient, nil
}

// GetSecret retrieves a secret from Vault
func (s *VaultService) GetSecret(ctx context.Context, path string) (*vault.Secret, error) {
	client, err := s.ensureClient()
	if err != nil {
		return nil, err
	}

	log.Printf("[Vault] Getting secret: %s", path)
	secret, err := client.GetSecret(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	log.Printf("[Vault] Successfully retrieved secret: %s", path)
	return secret, nil
}

// WriteSecret writes a secret to Vault
func (s *VaultService) WriteSecret(ctx context.Context, path string, data map[string]interface{}) error {
	client, err := s.ensureClient()
	if err != nil {
		return err
	}

	log.Printf("[Vault] Writing secret: %s", path)
	err = client.WriteSecret(ctx, path, data)
	if err != nil {
		return fmt.Errorf("failed to write secret: %w", err)
	}

	log.Printf("[Vault] Successfully wrote secret: %s", path)
	return nil
}

// DeleteSecret deletes a secret from Vault
func (s *VaultService) DeleteSecret(ctx context.Context, path string) error {
	client, err := s.ensureClient()
	if err != nil {
		return err
	}

	log.Printf("[Vault] Deleting secret: %s", path)
	err = client.DeleteSecret(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	log.Printf("[Vault] Successfully deleted secret: %s", path)
	return nil
}

// ListSecrets lists secrets at a path
func (s *VaultService) ListSecrets(ctx context.Context, path string) ([]string, error) {
	client, err := s.ensureClient()
	if err != nil {
		return nil, err
	}

	log.Printf("[Vault] Listing secrets at: %s", path)
	secrets, err := client.ListSecrets(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	log.Printf("[Vault] Found %d secrets", len(secrets))
	return secrets, nil
}

// ListSecretsExt returns directory entries with full details (data + metadata)
// for tree-view rendering in a single round trip.
func (s *VaultService) ListSecretsExt(ctx context.Context, path string) ([]vault.SecretEntryExt, error) {
	client, err := s.ensureClient()
	if err != nil {
		return nil, err
	}

	log.Printf("[Vault] Listing secrets (ext) at: %s", path)
	entries, err := client.ListSecretsExt(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	log.Printf("[Vault] Found %d entries (ext)", len(entries))
	return entries, nil
}

// Health checks Vault health
func (s *VaultService) Health(ctx context.Context) (bool, error) {
	client, err := s.ensureClient()
	if err != nil {
		return false, err
	}

	log.Printf("[Vault] Checking health")
	healthy, err := client.Health(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check health: %w", err)
	}

	log.Printf("[Vault] Health check: %v", healthy)
	return healthy, nil
}

// ListVersions returns version metadata for a secret path (KV v2)
func (s *VaultService) ListVersions(ctx context.Context, path string) ([]map[string]interface{}, error) {
	client, err := s.ensureClient()
	if err != nil {
		return nil, err
	}
	return client.ListVersions(ctx, path)
}

// RestoreVersion restores a deleted version of a secret (KV v2)
func (s *VaultService) RestoreVersion(ctx context.Context, path string, version int) error {
	client, err := s.ensureClient()
	if err != nil {
		return err
	}
	log.Printf("[Vault] Restoring version %d of secret: %s", version, path)
	return client.RestoreVersion(ctx, path, version)
}
