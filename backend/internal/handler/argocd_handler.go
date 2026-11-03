package handler

import (
	"net/http"

	"github.com/company/release-portal/internal/integration/argocd"
	"github.com/company/release-portal/internal/repository"
	"github.com/gin-gonic/gin"
)

// ArgoCDHandler exposes ArgoCD application management to the portal UI.
// It is the GitOps-native replacement for direct Kubernetes reads: all
// deployment state and sync operations go through ArgoCD, which owns the
// clusters as its destination.
type ArgoCDHandler struct {
	integrationRepo *repository.IntegrationRepository
}

func NewArgoCDHandler(integrationRepo *repository.IntegrationRepository) *ArgoCDHandler {
	return &ArgoCDHandler{integrationRepo: integrationRepo}
}

// resolveClient builds an ArgoCD client from the enabled argocd integration
// in DB (works without restart after adding/editing the integration).
func (h *ArgoCDHandler) resolveClient() (*argocd.Client, error) {
	if h.integrationRepo == nil {
		return nil, argocd.ErrNoIntegration
	}
	integrations, err := h.integrationRepo.GetEnabledByType("argocd")
	if err != nil {
		return nil, err
	}
	for _, integ := range integrations {
		token, _ := integ.Config["token"].(string)
		if token == "" {
			token, _ = integ.Config["apiToken"].(string)
		}
		if integ.URL != "" && token != "" {
			return argocd.NewClient(integ.URL, token), nil
		}
	}
	return nil, argocd.ErrNoIntegration
}

// ListApplications handles GET /api/argocd/applications
// Returns ArgoCD applications with sync/health status, source repo, target
// revision and destination namespace — the canonical deploy view.
func (h *ArgoCDHandler) ListApplications(c *gin.Context) {
	client, err := h.resolveClient()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	apps, err := client.ListApplications(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list ArgoCD applications: " + err.Error()})
		return
	}

	result := make([]gin.H, 0, len(apps))
	for _, app := range apps {
		item := gin.H{
			"name":           app.Metadata.Name,
			"namespace":      app.Metadata.Namespace,
			"sync_status":    app.Status.Sync.Status,
			"health_status":  app.Status.Health.Status,
			"revision":       app.Status.Sync.Revision,
			"repo_url":       app.Spec.Source.RepoURL,
			"path":           app.Spec.Source.Path,
			"targetRevision": app.Spec.Source.TargetRevision,
			"dest_server":    app.Spec.Destination.Server,
			"dest_namespace": app.Spec.Destination.Namespace,
		}
		result = append(result, item)
	}

	c.JSON(http.StatusOK, result)
}

// GetApplication handles GET /api/argocd/applications/:name
func (h *ArgoCDHandler) GetApplication(c *gin.Context) {
	client, err := h.resolveClient()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	app, err := client.GetApplication(c.Request.Context(), c.Param("name"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "application not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, app)
}

// SyncApplication handles POST /api/argocd/applications/:name/sync
// Triggers a GitOps sync — the only supported way to apply changes.
func (h *ArgoCDHandler) SyncApplication(c *gin.Context) {
	client, err := h.resolveClient()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	name := c.Param("name")
	if err := client.SyncApplication(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync application: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "sync triggered for " + name})
}
