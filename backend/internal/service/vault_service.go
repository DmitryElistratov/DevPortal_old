package service

import (
	"context"
	"fmt"
	"log"

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

// ensureClient returns a usable client, building one from the enabled "vault"
// integration in the database if needed.
func (s *VaultService) ensureClient() (*vault.Client, error) {
	if s.vaultClient != nil {
		return s.vaultClient, nil
	}
	if s.integration == nil {
		return nil, fmt.Errorf("vault client not configured")
	}
	integrations, err := s.integration.GetEnabledByType("vault")
	if err != nil || len(integrations) == 0 {
		return nil, fmt.Errorf("no enabled Vault integration found; configure it in Settings")
	}
	it := integrations[0]
	token, _ := it.Config["token"].(string)
	if token == "" {
		return nil, fmt.Errorf("Vault integration %q has no token configured", it.Name)
	}
	s.vaultClient = vault.NewClient(it.URL, token)
	log.Printf("[Vault] Client lazily initialized from integration %q (%s)", it.Name, it.URL)
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
