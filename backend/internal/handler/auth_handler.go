package handler

import (
	"net/http"

	"github.com/company/release-portal/internal/model"
	"github.com/company/release-portal/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("userID")
	user, err := h.authService.GetByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// ListUsers returns all portal users for the Users settings tab.
func (h *AuthHandler) ListUsers(c *gin.Context) {
	users, err := h.authService.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// OAuthLogin completes an external-provider login on the backend side.
// The frontend obtains a verified identity from the provider (OIDC id_token,
// GitHub/GitLab access token via their OAuth flow) and posts the resulting
// profile here; the backend maps it to a local user and issues a portal JWT.
func (h *AuthHandler) OAuthLogin(c *gin.Context) {
	var req struct {
		Provider   string `json:"provider"`
		ExternalID string `json:"externalId"`
		Username   string `json:"username"`
		Email      string `json:"email"`
		FullName   string `json:"fullName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	switch req.Provider {
	case "keycloak", "oidc", "github", "gitlab", "ldap":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported auth provider: " + req.Provider})
		return
	}
	if req.ExternalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "externalId is required"})
		return
	}

	resp, err := h.authService.LoginOAuth(req.Provider, req.ExternalID, req.Username, req.Email, req.FullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
