package handler

import (
	"net/http"

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
	path := c.Param("path")

	secret, err := h.vaultService.GetSecret(c.Request.Context(), path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, secret)
}

// WriteSecret handles POST /api/vault/secrets/:path
func (h *VaultHandler) WriteSecret(c *gin.Context) {
	path := c.Param("path")

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
	path := c.Param("path")

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

// ListVersions handles GET /api/vault/secrets/:path/versions
func (h *VaultHandler) ListVersions(c *gin.Context) {
	path := c.Param("path")

	versions, err := h.vaultService.ListVersions(c.Request.Context(), path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions})
}

// RestoreVersion handles POST /api/vault/secrets/:path/restore
func (h *VaultHandler) RestoreVersion(c *gin.Context) {
	path := c.Param("path")

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
