package handler

import (
        "context"
        "crypto/rand"
        "encoding/hex"
        "fmt"
        "net/http"
        "time"

        "github.com/company/release-portal/internal/integration/argocd"
        "github.com/company/release-portal/internal/integration/bitbucket"
        "github.com/company/release-portal/internal/integration/github"
        "github.com/company/release-portal/internal/integration/gitlab"
        "github.com/company/release-portal/internal/integration/grafana"
        "github.com/company/release-portal/internal/integration/graylog"
        "github.com/company/release-portal/internal/integration/jira"
        "github.com/company/release-portal/internal/integration/kubernetes"
        "github.com/company/release-portal/internal/integration/vault"
        "github.com/company/release-portal/internal/model"
        "github.com/company/release-portal/internal/repository"
        "github.com/gin-gonic/gin"
)

type IntegrationHandler struct {
        repo         *repository.IntegrationRepository
        gitlabClient *gitlab.Client
}

func NewIntegrationHandler(repo *repository.IntegrationRepository, gitlabClient *gitlab.Client) *IntegrationHandler {
        return &IntegrationHandler{
                repo:         repo,
                gitlabClient: gitlabClient,
        }
}

func (h *IntegrationHandler) List(c *gin.Context) {
        integrations, err := h.repo.List()
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }
        c.JSON(http.StatusOK, integrations)
}

func (h *IntegrationHandler) Get(c *gin.Context) {
        id := c.Param("id")
        integration, err := h.repo.GetByID(id)
        if err != nil {
                c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
                return
        }
        c.JSON(http.StatusOK, integration)
}

func normalizeIntegrationType(integrationType string) string {
        switch integrationType {
        case "git":
                // Frontend may send a generic "git" type; default to gitlab.
                return "gitlab"
        default:
                return integrationType
        }
}

func generateIntegrationID() string {
        b := make([]byte, 16)
        if _, err := rand.Read(b); err != nil {
                return fmt.Sprintf("int-%d", time.Now().UnixNano())
        }
        return hex.EncodeToString(b)
}

func (h *IntegrationHandler) Create(c *gin.Context) {
        var integration model.Integration
        if err := c.ShouldBindJSON(&integration); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        integration.Type = normalizeIntegrationType(integration.Type)
        if integration.ID == "" {
                integration.ID = generateIntegrationID()
        }
        if integration.Config == nil {
                integration.Config = map[string]interface{}{}
        }

        if err := h.repo.Create(&integration); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusCreated, integration)
}

func (h *IntegrationHandler) Update(c *gin.Context) {
        id := c.Param("id")
        var integration model.Integration
        if err := c.ShouldBindJSON(&integration); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }
        integration.ID = id
        integration.Type = normalizeIntegrationType(integration.Type)
        if integration.Config == nil {
                integration.Config = map[string]interface{}{}
        }

        if err := h.repo.Update(&integration); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        c.JSON(http.StatusOK, integration)
}

func (h *IntegrationHandler) Delete(c *gin.Context) {
        id := c.Param("id")
        if err := h.repo.Delete(id); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }
        c.JSON(http.StatusNoContent, nil)
}

func (h *IntegrationHandler) GetByType(c *gin.Context) {
        integrationType := c.Param("type")
        integrations, err := h.repo.GetByType(integrationType)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }
        c.JSON(http.StatusOK, integrations)
}

func (h *IntegrationHandler) Test(c *gin.Context) {
        id := c.Param("id")
        integration, err := h.repo.GetByID(id)
        if err != nil {
                c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
                return
        }

        ctx := context.Background()

        // Test connection based on integration type
        switch integration.Type {
        case "gitlab":
                if h.gitlabClient == nil {
                        c.JSON(http.StatusInternalServerError, gin.H{
                                "success": false,
                                "message": "GitLab client not initialized",
                        })
                        return
                }

                // Create a temporary client with integration credentials
                token, ok := integration.Config["token"].(string)
                if !ok || token == "" {
                        c.JSON(http.StatusBadRequest, gin.H{
                                "success": false,
                                "message": "GitLab token not configured",
                        })
                        return
                }

                tempClient := gitlab.NewClient(integration.URL, token)
                if err := tempClient.TestConnection(ctx); err != nil {
                        c.JSON(http.StatusOK, gin.H{
                                "success": false,
                                "message": "Connection test failed: " + err.Error(),
                                "type":    integration.Type,
                                "name":    integration.Name,
                        })
                        return
                }

        case "github":
                token, ok := integration.Config["token"].(string)
                if !ok || token == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "GitHub token not configured"})
                        return
                }
                client := github.NewClient(token)
                if err := client.TestConnection(ctx); err != nil {
                        c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error()})
                        return
                }

        case "bitbucket":
                username, _ := integration.Config["username"].(string)
                password, _ := integration.Config["password"].(string)
                if username == "" || password == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Bitbucket credentials not configured"})
                        return
                }
                client := bitbucket.NewClient(username, password)
                if err := client.TestConnection(ctx); err != nil {
                        c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error()})
                        return
                }

        case "jira":
                username, _ := integration.Config["username"].(string)
                token, _ := integration.Config["token"].(string)
                if username == "" || token == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Jira credentials not configured"})
                        return
                }
                client := jira.NewClient(integration.URL, username, token)
                if err := client.TestConnection(ctx); err != nil {
                        c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error()})
                        return
                }

        case "argocd":
                token, ok := integration.Config["token"].(string)
                if !ok || token == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ArgoCD token not configured"})
                        return
                }
                client := argocd.NewClient(integration.URL, token)
                if err := client.TestConnection(ctx); err != nil {
                        c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error()})
                        return
                }

        case "kubernetes":
                token, ok := integration.Config["token"].(string)
                if !ok || token == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Kubernetes token not configured"})
                        return
                }
                client := kubernetes.NewClient(integration.URL, token)
                if err := client.TestConnection(ctx); err != nil {
                        c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error()})
                        return
                }

        case "grafana":
                apiKey, ok := integration.Config["apiKey"].(string)
                if !ok || apiKey == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Grafana API key not configured"})
                        return
                }
                client := grafana.NewClient(integration.URL, apiKey)
                if err := client.TestConnection(ctx); err != nil {
                        c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error()})
                        return
                }

        case "graylog":
                username, _ := integration.Config["username"].(string)
                password, _ := integration.Config["password"].(string)
                if username == "" || password == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Graylog credentials not configured"})
                        return
                }
                client := graylog.NewClient(integration.URL, username, password)
                if err := client.TestConnection(ctx); err != nil {
                        c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error()})
                        return
                }

        case "vault":
                token, ok := integration.Config["token"].(string)
                if !ok || token == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Vault token not configured"})
                        return
                }
                client := vault.NewClient(integration.URL, token)
                if err := client.TestConnection(ctx); err != nil {
                        c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error()})
                        return
                }

        default:
                c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Unknown integration type: " + integration.Type})
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Connection test successful",
                "type":    integration.Type,
                "name":    integration.Name,
        })
}

func (h *IntegrationHandler) Sync(c *gin.Context) {
        id := c.Param("id")

        integration, err := h.repo.GetByID(id)
        if err != nil {
                c.JSON(http.StatusNotFound, gin.H{
                        "success": false,
                        "message": "Integration not found",
                        "id":      id,
                })
                return
        }

        ctx := c.Request.Context()

        // Sync based on integration type
        switch integration.Type {
        case "gitlab":
                token, ok := integration.Config["token"].(string)
                if !ok || token == "" {
                        c.JSON(http.StatusBadRequest, gin.H{
                                "success": false,
                                "message": "GitLab token not configured",
                        })
                        return
                }

                tempClient := gitlab.NewClient(integration.URL, token)

                // Get projects from GitLab
                projects, err := tempClient.ListProjects(ctx)
                if err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{
                                "success": false,
                                "message": "Failed to sync projects: " + err.Error(),
                                "type":    integration.Type,
                                "name":    integration.Name,
                        })
                        return
                }

                // Update lastSync timestamp
                integration.UpdatedAt = time.Now()
                if err := h.repo.Update(integration); err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{
                                "success": false,
                                "message": "Failed to update integration: " + err.Error(),
                        })
                        return
                }

                c.JSON(http.StatusOK, gin.H{
                        "success":        true,
                        "message":        "Sync completed successfully",
                        "type":           integration.Type,
                        "name":           integration.Name,
                        "synced_at":      integration.UpdatedAt,
                        "projects_count": len(projects),
                })
                return

        case "github":
                token, ok := integration.Config["token"].(string)
                if !ok || token == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "GitHub token not configured"})
                        return
                }
                client := github.NewClient(token)
                repos, err := client.ListRepositories(ctx)
                if err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to sync repositories: " + err.Error()})
                        return
                }
                integration.UpdatedAt = time.Now()
                if err := h.repo.Update(integration); err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                        return
                }
                c.JSON(http.StatusOK, gin.H{
                        "success":        true,
                        "message":        "Sync completed successfully",
                        "type":           integration.Type,
                        "name":           integration.Name,
                        "synced_at":      integration.UpdatedAt,
                        "projects_count": len(repos),
                })
                return

        case "bitbucket":
                username, _ := integration.Config["username"].(string)
                password, _ := integration.Config["password"].(string)
                if username == "" || password == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Bitbucket credentials not configured"})
                        return
                }
                client := bitbucket.NewClient(username, password)
                repos, err := client.ListRepositories(ctx)
                if err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to sync repositories: " + err.Error()})
                        return
                }
                integration.UpdatedAt = time.Now()
                if err := h.repo.Update(integration); err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                        return
                }
                c.JSON(http.StatusOK, gin.H{
                        "success":        true,
                        "message":        "Sync completed successfully",
                        "type":           integration.Type,
                        "name":           integration.Name,
                        "synced_at":      integration.UpdatedAt,
                        "projects_count": len(repos),
                })
                return

        case "jira":
                username, _ := integration.Config["username"].(string)
                token, _ := integration.Config["token"].(string)
                if username == "" || token == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Jira credentials not configured"})
                        return
                }
                client := jira.NewClient(integration.URL, username, token)
                projects, err := client.ListProjects(ctx)
                if err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to sync projects: " + err.Error()})
                        return
                }
                integration.UpdatedAt = time.Now()
                if err := h.repo.Update(integration); err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                        return
                }
                c.JSON(http.StatusOK, gin.H{
                        "success":        true,
                        "message":        "Sync completed successfully",
                        "type":           integration.Type,
                        "name":           integration.Name,
                        "synced_at":      integration.UpdatedAt,
                        "projects_count": len(projects),
                })
                return

        case "argocd":
                token, ok := integration.Config["token"].(string)
                if !ok || token == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ArgoCD token not configured"})
                        return
                }
                client := argocd.NewClient(integration.URL, token)
                apps, err := client.ListApplications(ctx)
                if err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to sync applications: " + err.Error()})
                        return
                }
                integration.UpdatedAt = time.Now()
                if err := h.repo.Update(integration); err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                        return
                }
                c.JSON(http.StatusOK, gin.H{
                        "success":        true,
                        "message":        "Sync completed successfully",
                        "type":           integration.Type,
                        "name":           integration.Name,
                        "synced_at":      integration.UpdatedAt,
                        "projects_count": len(apps),
                })
                return

        case "kubernetes":
                token, ok := integration.Config["token"].(string)
                if !ok || token == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Kubernetes token not configured"})
                        return
                }
                client := kubernetes.NewClient(integration.URL, token)
                pods, err := client.ListPods(ctx, "default")
                if err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to sync pods: " + err.Error()})
                        return
                }
                integration.UpdatedAt = time.Now()
                if err := h.repo.Update(integration); err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                        return
                }
                c.JSON(http.StatusOK, gin.H{
                        "success":        true,
                        "message":        "Sync completed successfully",
                        "type":           integration.Type,
                        "name":           integration.Name,
                        "synced_at":      integration.UpdatedAt,
                        "projects_count": len(pods),
                })
                return

        case "grafana":
                apiKey, ok := integration.Config["apiKey"].(string)
                if !ok || apiKey == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Grafana API key not configured"})
                        return
                }
                client := grafana.NewClient(integration.URL, apiKey)
                dashboards, err := client.ListDashboards(ctx)
                if err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to sync dashboards: " + err.Error()})
                        return
                }
                integration.UpdatedAt = time.Now()
                if err := h.repo.Update(integration); err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                        return
                }
                c.JSON(http.StatusOK, gin.H{
                        "success":        true,
                        "message":        "Sync completed successfully",
                        "type":           integration.Type,
                        "name":           integration.Name,
                        "synced_at":      integration.UpdatedAt,
                        "projects_count": len(dashboards),
                })
                return

        case "graylog":
                username, _ := integration.Config["username"].(string)
                password, _ := integration.Config["password"].(string)
                if username == "" || password == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Graylog credentials not configured"})
                        return
                }
                client := graylog.NewClient(integration.URL, username, password)
                streams, err := client.ListStreams(ctx)
                if err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to sync streams: " + err.Error()})
                        return
                }
                integration.UpdatedAt = time.Now()
                if err := h.repo.Update(integration); err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                        return
                }
                c.JSON(http.StatusOK, gin.H{
                        "success":        true,
                        "message":        "Sync completed successfully",
                        "type":           integration.Type,
                        "name":           integration.Name,
                        "synced_at":      integration.UpdatedAt,
                        "projects_count": len(streams),
                })
                return

        case "vault":
                token, ok := integration.Config["token"].(string)
                if !ok || token == "" {
                        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Vault token not configured"})
                        return
                }
                client := vault.NewClient(integration.URL, token)
                secrets, err := client.ListSecrets(ctx, "")
                if err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to sync secrets: " + err.Error()})
                        return
                }
                integration.UpdatedAt = time.Now()
                if err := h.repo.Update(integration); err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                        return
                }
                c.JSON(http.StatusOK, gin.H{
                        "success":        true,
                        "message":        "Sync completed successfully",
                        "type":           integration.Type,
                        "name":           integration.Name,
                        "synced_at":      integration.UpdatedAt,
                        "projects_count": len(secrets),
                })
                return

        default:
                c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Unknown integration type: " + integration.Type})
                return
        }
}
