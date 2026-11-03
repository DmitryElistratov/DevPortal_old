package handler

import (
	"net/http"
	"strings"

	"github.com/company/release-portal/internal/service"
	"github.com/gin-gonic/gin"
)

type VaultHandler struct {
	vaultService *service.VaultService
}

func NewVaultHandler(vaultService *service.VaultService) *VaultHandler {
	return &VaultHandler{vaultService: vaultService}
}

// GetSecret handles GET /api/vault/secrets/:path
func (h *VaultHandler) GetSecret(c *gin.Context) {
	path := strings.TrimPrefix(c.Param("path"), "/")

	secret, err := h.vaultService.GetSecret(c.Request.Context(), path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// A KV v2 soft-deleted or destroyed version returns 200 with empty data and
	// non-empty deletion markers — surface that as 404 so the GUI doesn't render
	// a phantom secret entry.
	if len(secret.Data) == 0 && (secret.Metadata.DeletionTime != "" || secret.Metadata.Destroyed) {
		c.JSON(http.StatusNotFound, gin.H{"error": "secret not found or deleted"})
		return
	}

	c.JSON(http.StatusOK, secret)
}

// WriteSecret handles POST /api/vault/secrets/:path
func (h *VaultHandler) WriteSecret(c *gin.Context) {
	path := strings.TrimPrefix(c.Param("path"), "/")

	var req struct {
		Data map[string]interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.vaultService.WriteSecret(c.Request.Context(), path, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "secret written successfully"})
}

// DeleteSecret handles DELETE /api/vault/secrets/:path
func (h *VaultHandler) DeleteSecret(c *gin.Context) {
	path := strings.TrimPrefix(c.Param("path"), "/")

	err := h.vaultService.DeleteSecret(c.Request.Context(), path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "secret deleted successfully"})
}

// ListSecrets handles GET /api/vault/secrets
func (h *VaultHandler) ListSecrets(c *gin.Context) {
	path := c.DefaultQuery("path", "")

	secrets, err := h.vaultService.ListSecrets(c.Request.Context(), path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, secrets)
}

// UpdateSecret handles PUT /api/vault/secrets/:path (alias of WriteSecret)
func (h *VaultHandler) UpdateSecret(c *gin.Context) {
	h.WriteSecret(c)
}

// ListVersions returns version metadata. The path is either passed explicitly by
// the router wrapper (nested wildcard paths like /secrets/a/b/versions) or taken
// from the "path" param with the "/versions" suffix stripped.
func (h *VaultHandler) ListVersions(c *gin.Context, explicitPath ...string) {
	path := strings.TrimPrefix(c.Param("path"), "/")
	path = strings.TrimSuffix(path, "/versions")
	if len(explicitPath) > 0 && explicitPath[0] != "" {
		path = explicitPath[0]
	}

	versions, err := h.vaultService.ListVersions(c.Request.Context(), path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions})
}

// RestoreVersion restores a KV secret version. The path is either passed
// explicitly by the router wrapper or taken from the wildcard "path" param with
// the "/restore" suffix stripped.
func (h *VaultHandler) RestoreVersion(c *gin.Context, explicitPath ...string) {
	path := strings.TrimPrefix(c.Param("path"), "/")
	path = strings.TrimSuffix(path, "/restore")
	if len(explicitPath) > 0 && explicitPath[0] != "" {
		path = explicitPath[0]
	}

	var req struct {
		Version int `json:"version" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.vaultService.RestoreVersion(c.Request.Context(), path, req.Version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "version restored successfully"})
}

// Health handles GET /api/vault/health
func (h *VaultHandler) Health(c *gin.Context) {
	healthy, err := h.vaultService.Health(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"healthy": healthy})
}
