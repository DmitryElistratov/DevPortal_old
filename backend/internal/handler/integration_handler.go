package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
	serviceRepo  *repository.ServiceRepository
	userRepo     *repository.UserRepository
	permRepo     *repository.PermissionRepository
	gitlabClient *gitlab.Client
}

func NewIntegrationHandler(repo *repository.IntegrationRepository, gitlabClient *gitlab.Client, opts ...func(*IntegrationHandler)) *IntegrationHandler {
	h := &IntegrationHandler{
		repo:         repo,
		gitlabClient: gitlabClient,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// WithServiceRepo attaches the services repository so Git sync can persist projects.
func WithServiceRepo(r *repository.ServiceRepository) func(*IntegrationHandler) {
	return func(h *IntegrationHandler) { h.serviceRepo = r }
}

// WithUserRepo attaches the users repository so Git sync can import GitLab users.
func WithUserRepo(r *repository.UserRepository) func(*IntegrationHandler) {
	return func(h *IntegrationHandler) { h.userRepo = r }
}

// WithPermRepo attaches the project permissions repository so Git sync can
// import per-project access levels (rights) from GitLab.
func WithPermRepo(r *repository.PermissionRepository) func(*IntegrationHandler) {
	return func(h *IntegrationHandler) { h.permRepo = r }
}

// syncProjectData is a provider-agnostic representation of a repository/project
// returned by Git integrations.
type syncProjectData struct {
	Name        string
	Description string
	URL         string // web URL of the project
	RepoURL     string // clone/repository URL stored in services.repository
}

// persistSyncedProjects upserts projects fetched from a Git provider into the
// services table so they become visible on the Services page after sync.
// Returns the number of successfully saved projects and their DB ids (index-aligned).
func (h *IntegrationHandler) persistSyncedProjects(projects []syncProjectData) (int, []string, error) {
	if h.serviceRepo == nil {
		return 0, nil, fmt.Errorf("service repository not initialized")
	}
	saved := 0
	ids := make([]string, len(projects))
	var firstErr error
	for i, p := range projects {
		if strings.TrimSpace(p.Name) == "" {
			continue
		}
		svc := &model.Service{
			ID:          generateIntegrationID(),
			Name:        p.Name,
			Type:        "service",
			Lifecycle:   "production",
			Description: p.Description,
			Repository:  p.RepoURL,
		}
		if svc.Repository == "" {
			svc.Repository = p.URL
		}
		if err := h.serviceRepo.Upsert(svc); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		ids[i] = svc.ID
		saved++
	}
	return saved, ids, firstErr
}

// importGitLabUsers creates or updates portal users for every unique GitLab
// member seen during sync, so the Users settings tab reflects Git accounts.
// Returns the number of users imported/updated.
func (h *IntegrationHandler) importGitLabUsers(members []gitlab.GitLabMember) int {
	if h.userRepo == nil {
		return 0
	}
	imported := 0
	seen := map[int]bool{}
	for _, m := range members {
		if m.ID == 0 || seen[m.ID] || strings.TrimSpace(m.Username) == "" {
			continue
		}
		seen[m.ID] = true
		externalID := fmt.Sprintf("%d", m.ID)
		fullName := m.Name
		if fullName == "" {
			fullName = m.Username
		}
		existing, err := h.userRepo.GetByExternalID("gitlab", externalID)
		if err == nil && existing != nil {
			// Refresh profile fields of an already-imported user.
			existing.FullName = fullName
			existing.Email = m.Email
			_ = h.userRepo.Update(existing)
			imported++
			continue
		}
		blocked := m.State != "active"
		role := "user"
		if blocked {
			// Blocked users are stored with the "blocked" role — it is rendered
			// in the Users tab and prevents them from gaining portal rights.
			role = "blocked"
		}
		user := &model.User{
			ID:           generateIntegrationID(),
			Username:     m.Username,
			Email:        m.Email,
			FullName:     fullName,
			Role:         role,
			AuthProvider: "gitlab",
			ExternalID:   externalID,
		}
		if err := h.userRepo.Create(user); err != nil {
			// Username collision — try to link the local account to GitLab instead.
			if local, lerr := h.userRepo.GetByUsername(m.Username); lerr == nil && local != nil {
				local.AuthProvider = "gitlab"
				local.ExternalID = externalID
				if local.Email == "" {
					local.Email = m.Email
				}
				if local.FullName == "" {
					local.FullName = fullName
				}
				_ = h.userRepo.Update(local)
				imported++
			}
			continue
		}
		imported++
	}
	return imported
}

// syncGitLabPermissions imports per-project access levels from GitLab into the
// project_permissions table. serviceIDs is index-aligned with projects; empty
// entries mean the project was not persisted and its permissions are skipped.
func (h *IntegrationHandler) syncGitLabPermissions(ctx context.Context, client *gitlab.Client, projects []gitlab.Project, serviceIDs []string, membersCache map[string][]gitlab.GitLabMember) (int, error) {
	if h.permRepo == nil || h.userRepo == nil {
		return 0, nil
	}
	total := 0
	var firstErr error
	for i, p := range projects {
		if i >= len(serviceIDs) || serviceIDs[i] == "" {
			continue
		}
		pid := p.ProjectID()
		members, ok := membersCache[pid]
		if !ok {
			var err error
			members, err = client.ListProjectMembers(ctx, pid)
			if err != nil {
				if firstErr == nil {
					firstErr = fmt.Errorf("project %s: %w", p.Name, err)
				}
				continue
			}
			membersCache[pid] = members
		}
		for _, m := range members {
			role := gitlab.AccessLevelToRole(m.AccessLevel)
			if role == "" {
				continue
			}
			user, err := h.userRepo.GetByExternalID("gitlab", fmt.Sprintf("%d", m.ID))
			if err != nil || user == nil {
				continue
			}
			perm := &model.ProjectPermission{
				ID:        generateIntegrationID(),
				UserID:    user.ID,
				ServiceID: serviceIDs[i],
				Role:      role,
				GrantedBy: "gitlab-sync",
			}
			if err := h.permRepo.Upsert(perm); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			total++
		}
	}
	return total, firstErr
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
	case "git", "gitea", "gerrit":
		// Frontend may send a generic "git"/other type; default to gitlab.
		return "gitlab"
	default:
		return integrationType
	}
}

// testWebhookConnection verifies that the configured notification webhook is
// reachable. Falls back to the integration URL when no separate webhook is set.
func testWebhookConnection(ctx context.Context, integrationURL, webhookURL string) error {
	target := strings.TrimSpace(webhookURL)
	if target == "" {
		target = strings.TrimSpace(integrationURL)
	}
	if target == "" {
		return fmt.Errorf("no webhook URL configured")
	}
	if _, err := url.ParseRequestURI(target); err != nil {
		return fmt.Errorf("invalid webhook URL %q", target)
	}
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("webhook endpoint returned HTTP %d", resp.StatusCode)
	}
	return nil
}

// configKeyAliases maps alternative spellings of credential keys that clients
// (or older frontend versions) may send, e.g. "api_key" instead of "apiKey".
var configKeyAliases = map[string][]string{
	"token":      {"apiToken", "api_token", "accessToken", "access_token", "apiKey", "api_key"},
	"apiKey":     {"api_key", "token", "apiToken", "access_token"},
	"password":   {"pass", "secret"},
	"username":   {"user", "login"},
	"projectKey": {"project_key", "project"},
	"webhookUrl": {"webhook_url", "webhook"},
}

// normalizeIntegrationConfig rewrites known credential keys to their canonical
// form so Test/Sync handlers can always find them under the expected name.
func normalizeIntegrationConfig(cfg map[string]interface{}) {
	if cfg == nil {
		return
	}
	for canonical, aliases := range configKeyAliases {
		if s, ok := cfg[canonical].(string); ok && s != "" {
			continue
		}
		delete(cfg, canonical)
		for _, alias := range aliases {
			if v, ok := cfg[alias]; ok {
				if s, isStr := v.(string); isStr && s != "" {
					cfg[canonical] = s
					break
				}
			}
		}
	}
}

// requireConfigString returns the first non-empty string value among the given keys.
func requireConfigString(cfg map[string]interface{}, keys ...string) (string, bool) {
	for _, k := range keys {
		if s, ok := cfg[k].(string); ok && s != "" {
			return s, true
		}
	}
	return "", false
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
	normalizeIntegrationConfig(integration.Config)

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

	existing, loadErr := h.repo.GetByID(id)
	if loadErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}
	if integration.URL == "" {
		integration.URL = existing.URL
	}
	if integration.Name == "" {
		integration.Name = existing.Name
	}
	if integration.Config == nil {
		integration.Config = existing.Config
	} else if existing.Config != nil {
		// Merge so partial updates don't wipe stored credentials.
		merged := existing.Config
		for k, v := range integration.Config {
			if s, isStr := v.(string); isStr && s == "" {
				continue
			}
			merged[k] = v
		}
		integration.Config = merged
	}
	if integration.Config == nil {
		integration.Config = map[string]interface{}{}
	}
	normalizeIntegrationConfig(integration.Config)

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
		token, ok := requireConfigString(integration.Config, "token", "apiToken", "apiKey", "access_token")
		if !ok {
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
		token, ok := requireConfigString(integration.Config, "token", "apiToken", "apiKey", "access_token")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "GitHub token not configured"})
			return
		}
		client := github.NewClient(token)
		if err := client.TestConnection(ctx); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error()})
			return
		}

	case "bitbucket":
		username, _ := requireConfigString(integration.Config, "username", "user", "login")
		password, _ := requireConfigString(integration.Config, "password", "pass", "secret", "token")
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
		username, _ := requireConfigString(integration.Config, "username", "user", "email", "login")
		token, _ := requireConfigString(integration.Config, "token", "apiToken", "api_token", "access_token")
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
		token, ok := requireConfigString(integration.Config, "token", "apiToken", "apiKey", "access_token")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ArgoCD token not configured"})
			return
		}
		client := argocd.NewClient(integration.URL, token)
		if err := client.TestConnection(ctx); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error()})
			return
		}

	case "kubernetes":
		token, ok := requireConfigString(integration.Config, "token", "apiToken", "apiKey", "access_token")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Kubernetes token not configured"})
			return
		}
		client := kubernetes.NewClient(integration.URL, token)
		if err := client.TestConnection(ctx); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error()})
			return
		}

	case "grafana":
		apiKey, ok := requireConfigString(integration.Config, "apiKey", "api_key", "token")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Grafana API key not configured"})
			return
		}
		client := grafana.NewClient(integration.URL, apiKey)
		if err := client.TestConnection(ctx); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error()})
			return
		}

	case "graylog":
		username, _ := requireConfigString(integration.Config, "username", "user", "login")
		password, _ := requireConfigString(integration.Config, "password", "pass", "secret", "token")
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
		token, ok := requireConfigString(integration.Config, "token", "apiToken", "apiKey", "access_token")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Vault token not configured"})
			return
		}
		client := vault.NewClient(integration.URL, token)
		if err := client.TestConnection(ctx); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error()})
			return
		}

	case "notification", "slack", "telegram":
		webhookURL, _ := requireConfigString(integration.Config, "webhookUrl", "webhook_url", "webhook")
		if webhookURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Notification webhook URL not configured"})
			return
		}
		if err := testWebhookConnection(ctx, integration.URL, webhookURL); err != nil {
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
		token, ok := requireConfigString(integration.Config, "token", "apiToken", "apiKey", "access_token")
		if !ok {
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

		// Persist synced projects into the services table so they show up on
		// the Services page after synchronization.
		synced := make([]syncProjectData, 0, len(projects))
		for _, p := range projects {
			synced = append(synced, syncProjectData{
				Name:        p.Name,
				Description: p.Description,
				URL:         p.WebURL,
				RepoURL:     p.HTTPURLToRepo,
			})
		}
		saved, serviceIDs, saveErr := h.persistSyncedProjects(synced)
		if saveErr != nil && saved == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to save synced projects: " + saveErr.Error(),
			})
			return
		}

		// Import GitLab users and per-project access levels (rights).
		membersCache := make(map[string][]gitlab.GitLabMember, len(projects))
		usersImported := 0
		permsSynced := 0
		var permWarn string
		for _, p := range projects {
			pid := p.ProjectID()
			members, err := tempClient.ListProjectMembers(ctx, pid)
			if err != nil {
				if permWarn == "" {
					permWarn = "members fetch failed for some projects: " + err.Error()
				}
				continue
			}
			membersCache[pid] = members
		}
		allMembers := make([]gitlab.GitLabMember, 0)
		for _, ms := range membersCache {
			allMembers = append(allMembers, ms...)
		}
		usersImported = h.importGitLabUsers(allMembers)
		permsSynced, permErr := h.syncGitLabPermissions(ctx, tempClient, projects, serviceIDs, membersCache)
		if permErr != nil && permWarn == "" {
			permWarn = "permissions sync warning: " + permErr.Error()
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

		resp := gin.H{
			"success":            true,
			"message":            "Sync completed successfully",
			"type":               integration.Type,
			"name":               integration.Name,
			"synced_at":          integration.UpdatedAt,
			"projects_count":     len(projects),
			"saved_count":        saved,
			"users_imported":     usersImported,
			"permissions_synced": permsSynced,
		}
		if permWarn != "" {
			resp["warning"] = permWarn
		}
		c.JSON(http.StatusOK, resp)
		return

	case "github":
		token, ok := requireConfigString(integration.Config, "token", "apiToken", "apiKey", "access_token")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "GitHub token not configured"})
			return
		}
		client := github.NewClient(token)
		repos, err := client.ListRepositories(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to sync repositories: " + err.Error()})
			return
		}
		synced := make([]syncProjectData, 0, len(repos))
		for _, r := range repos {
			synced = append(synced, syncProjectData{
				Name:    r.Name,
				URL:     r.HTMLURL,
				RepoURL: r.HTMLURL,
			})
		}
		saved, _, saveErr := h.persistSyncedProjects(synced)
		if saveErr != nil && saved == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to save synced repositories: " + saveErr.Error()})
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
			"saved_count":    saved,
		})
		return

	case "bitbucket":
		username, _ := requireConfigString(integration.Config, "username", "user", "login")
		password, _ := requireConfigString(integration.Config, "password", "pass", "secret", "token")
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
		bbSynced := make([]syncProjectData, 0, len(repos))
		for _, r := range repos {
			bbSynced = append(bbSynced, syncProjectData{
				Name:    r.Name,
				URL:     r.Links.HTML.Href,
				RepoURL: r.Links.HTML.Href,
			})
		}
		bbSaved, _, bbSaveErr := h.persistSyncedProjects(bbSynced)
		if bbSaveErr != nil && bbSaved == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to save synced repositories: " + bbSaveErr.Error()})
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
			"saved_count":    bbSaved,
		})
		return

	case "jira":
		username, _ := requireConfigString(integration.Config, "username", "user", "email", "login")
		token, _ := requireConfigString(integration.Config, "token", "apiToken", "api_token", "access_token")
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
		jiraSynced := make([]syncProjectData, 0, len(projects))
		for _, pr := range projects {
			name := strings.TrimSpace(pr.Name)
			if name == "" {
				name = pr.Key
			}
			jiraSynced = append(jiraSynced, syncProjectData{Name: name})
		}
		jiraSaved, _, jiraSaveErr := h.persistSyncedProjects(jiraSynced)
		if jiraSaveErr != nil && jiraSaved == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to save synced projects: " + jiraSaveErr.Error()})
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
			"saved_count":    jiraSaved,
		})
		return

	case "argocd":
		token, ok := requireConfigString(integration.Config, "token", "apiToken", "apiKey", "access_token")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ArgoCD token not configured"})
			return
		}
		client := argocd.NewClient(integration.URL, token)
		apps, err := client.ListApplications(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to sync applications: " + err.Error()})
			return
		}
		argoSynced := make([]syncProjectData, 0, len(apps))
		for _, app := range apps {
			argoSynced = append(argoSynced, syncProjectData{
				Name:    app.Metadata.Name,
				RepoURL: app.Spec.Source.RepoURL,
			})
		}
		argoSaved, _, argoSaveErr := h.persistSyncedProjects(argoSynced)
		if argoSaveErr != nil && argoSaved == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to save synced applications: " + argoSaveErr.Error()})
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
			"saved_count":    argoSaved,
		})
		return

	case "kubernetes":
		token, ok := requireConfigString(integration.Config, "token", "apiToken", "apiKey", "access_token")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Kubernetes token not configured"})
			return
		}
		client := kubernetes.NewClient(integration.URL, token)
		namespace, _ := requireConfigString(integration.Config, "namespace", "ns")
		if namespace == "" {
			namespace = "default"
		}
		deployments, err := client.ListDeployments(ctx, namespace)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to sync deployments: " + err.Error()})
			return
		}
		k8sSynced := make([]syncProjectData, 0, len(deployments))
		for _, d := range deployments {
			k8sSynced = append(k8sSynced, syncProjectData{Name: d.Metadata.Name})
		}
		k8sSaved, _, k8sSaveErr := h.persistSyncedProjects(k8sSynced)
		if k8sSaveErr != nil && k8sSaved == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to save synced deployments: " + k8sSaveErr.Error()})
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
			"projects_count": len(deployments),
			"saved_count":    k8sSaved,
		})
		return

	case "grafana":
		apiKey, ok := requireConfigString(integration.Config, "apiKey", "api_key", "token")
		if !ok {
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
		username, _ := requireConfigString(integration.Config, "username", "user", "login")
		password, _ := requireConfigString(integration.Config, "password", "pass", "secret", "token")
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
		token, ok := requireConfigString(integration.Config, "token", "apiToken", "apiKey", "access_token")
		if !ok {
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

	case "notification", "slack", "telegram":
		webhookURL, _ := requireConfigString(integration.Config, "webhookUrl", "webhook_url", "webhook")
		if webhookURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Notification webhook URL not configured"})
			return
		}
		if err := testWebhookConnection(ctx, integration.URL, webhookURL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to verify webhook: " + err.Error()})
			return
		}
		integration.UpdatedAt = time.Now()
		if err := h.repo.Update(integration); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"message":   "Sync completed successfully",
			"type":      integration.Type,
			"name":      integration.Name,
			"synced_at": integration.UpdatedAt,
		})
		return

	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Unknown integration type: " + integration.Type})
		return
	}
}
