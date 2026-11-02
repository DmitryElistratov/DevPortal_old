package service

import (
	"context"
	"fmt"
	"log"

	"github.com/company/release-portal/internal/integration/vault"
)

type VaultService struct {
	vaultClient *vault.Client
}

func NewVaultService(vaultClient *vault.Client) *VaultService {
	return &VaultService{vaultClient: vaultClient}
}

// GetSecret retrieves a secret from Vault
func (s *VaultService) GetSecret(ctx context.Context, path string) (*vault.Secret, error) {
	if s.vaultClient == nil {
		return nil, fmt.Errorf("vault client not configured")
	}

	log.Printf("[Vault] Getting secret: %s", path)
	secret, err := s.vaultClient.GetSecret(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	log.Printf("[Vault] Successfully retrieved secret: %s", path)
	return secret, nil
}

// WriteSecret writes a secret to Vault
func (s *VaultService) WriteSecret(ctx context.Context, path string, data map[string]interface{}) error {
	if s.vaultClient == nil {
		return fmt.Errorf("vault client not configured")
	}

	log.Printf("[Vault] Writing secret: %s", path)
	err := s.vaultClient.WriteSecret(ctx, path, data)
	if err != nil {
		return fmt.Errorf("failed to write secret: %w", err)
	}

	log.Printf("[Vault] Successfully wrote secret: %s", path)
	return nil
}

// DeleteSecret deletes a secret from Vault
func (s *VaultService) DeleteSecret(ctx context.Context, path string) error {
	if s.vaultClient == nil {
		return fmt.Errorf("vault client not configured")
	}

	log.Printf("[Vault] Deleting secret: %s", path)
	err := s.vaultClient.DeleteSecret(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	log.Printf("[Vault] Successfully deleted secret: %s", path)
	return nil
}

// ListSecrets lists secrets at a path
func (s *VaultService) ListSecrets(ctx context.Context, path string) ([]string, error) {
	if s.vaultClient == nil {
		return nil, fmt.Errorf("vault client not configured")
	}

	log.Printf("[Vault] Listing secrets at: %s", path)
	secrets, err := s.vaultClient.ListSecrets(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	log.Printf("[Vault] Found %d secrets", len(secrets))
	return secrets, nil
}

// Health checks Vault health
func (s *VaultService) Health(ctx context.Context) (bool, error) {
	if s.vaultClient == nil {
		return false, fmt.Errorf("vault client not configured")
	}

	log.Printf("[Vault] Checking health")
	healthy, err := s.vaultClient.Health(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check health: %w", err)
	}

	log.Printf("[Vault] Health check: %v", healthy)
	return healthy, nil
}
