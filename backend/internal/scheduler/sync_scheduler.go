package scheduler

import (
	"context"
	"log"
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
)

type SyncScheduler struct {
	integrationRepo *repository.IntegrationRepository

	// Clients for each integration type
	gitlabClients    map[string]*gitlab.Client
	githubClients    map[string]*github.Client
	bitbucketClients map[string]*bitbucket.Client
	jiraClients      map[string]*jira.Client
	argocdClients    map[string]*argocd.Client
	k8sClients       map[string]*kubernetes.Client
	grafanaClients   map[string]*grafana.Client
	graylogClients   map[string]*graylog.Client
	vaultClients     map[string]*vault.Client

	syncInterval time.Duration
}

func NewSyncScheduler(integrationRepo *repository.IntegrationRepository) *SyncScheduler {
	return &SyncScheduler{
		integrationRepo:  integrationRepo,
		gitlabClients:    make(map[string]*gitlab.Client),
		githubClients:    make(map[string]*github.Client),
		bitbucketClients: make(map[string]*bitbucket.Client),
		jiraClients:      make(map[string]*jira.Client),
		argocdClients:    make(map[string]*argocd.Client),
		k8sClients:       make(map[string]*kubernetes.Client),
		grafanaClients:   make(map[string]*grafana.Client),
		graylogClients:   make(map[string]*graylog.Client),
		vaultClients:     make(map[string]*vault.Client),
		syncInterval:     1 * time.Minute, // Sync every minute
	}
}

func (s *SyncScheduler) Start(ctx context.Context) {
	log.Println("Starting sync scheduler...")

	// Initial sync
	s.syncAllIntegrations(ctx)

	// Start periodic sync
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Sync scheduler stopped")
			return
		case <-ticker.C:
			s.syncAllIntegrations(ctx)
		}
	}
}

func (s *SyncScheduler) syncAllIntegrations(ctx context.Context) {
	log.Println("Starting sync for all integrations...")

	// Get all enabled integrations
	integrations, err := s.integrationRepo.List()
	if err != nil {
		log.Printf("Failed to list integrations: %v", err)
		return
	}

	for _, integration := range integrations {
		if !integration.Enabled {
			continue
		}

		switch integration.Type {
		case "gitlab":
			s.syncGitLab(ctx, integration)
		case "github":
			s.syncGitHub(ctx, integration)
		case "bitbucket":
			s.syncBitbucket(ctx, integration)
		case "jira":
			s.syncJira(ctx, integration)
		case "argocd":
			s.syncArgoCD(ctx, integration)
		case "kubernetes":
			s.syncKubernetes(ctx, integration)
		case "grafana":
			s.syncGrafana(ctx, integration)
		case "graylog":
			s.syncGraylog(ctx, integration)
		case "vault":
			s.syncVault(ctx, integration)
		}
	}

	log.Println("Sync completed for all integrations")
}

func (s *SyncScheduler) syncGitLab(ctx context.Context, integration model.Integration) {
	token, ok := integration.Config["token"].(string)
	if !ok || token == "" {
		log.Printf("[GitLab] No token configured for %s", integration.Name)
		return
	}

	client := gitlab.NewClient(integration.URL, token)

	// Test connection
	if err := client.TestConnection(ctx); err != nil {
		log.Printf("[GitLab] Connection test failed for %s: %v", integration.Name, err)
		return
	}

	// Sync projects
	projects, err := client.ListProjects(ctx)
	if err != nil {
		log.Printf("[GitLab] Failed to sync projects for %s: %v", integration.Name, err)
		return
	}

	// Update last sync time
	integration.UpdatedAt = time.Now()
	if err := s.integrationRepo.Update(&integration); err != nil {
		log.Printf("[GitLab] Failed to update integration %s: %v", integration.Name, err)
		return
	}

	log.Printf("[GitLab] Synced %d projects for %s", len(projects), integration.Name)
}

func (s *SyncScheduler) syncGitHub(ctx context.Context, integration model.Integration) {
	token, ok := integration.Config["token"].(string)
	if !ok || token == "" {
		log.Printf("[GitHub] No token configured for %s", integration.Name)
		return
	}

	client := github.NewClient(token)

	// Test connection
	if err := client.TestConnection(ctx); err != nil {
		log.Printf("[GitHub] Connection test failed for %s: %v", integration.Name, err)
		return
	}

	// Sync repositories
	repos, err := client.ListRepositories(ctx)
	if err != nil {
		log.Printf("[GitHub] Failed to sync repositories for %s: %v", integration.Name, err)
		return
	}

	// Update last sync time
	integration.UpdatedAt = time.Now()
	if err := s.integrationRepo.Update(&integration); err != nil {
		log.Printf("[GitHub] Failed to update integration %s: %v", integration.Name, err)
		return
	}

	log.Printf("[GitHub] Synced %d repositories for %s", len(repos), integration.Name)
}

func (s *SyncScheduler) syncBitbucket(ctx context.Context, integration model.Integration) {
	username, _ := integration.Config["username"].(string)
	password, _ := integration.Config["password"].(string)

	if username == "" || password == "" {
		log.Printf("[Bitbucket] No credentials configured for %s", integration.Name)
		return
	}

	client := bitbucket.NewClient(username, password)

	// Test connection
	if err := client.TestConnection(ctx); err != nil {
		log.Printf("[Bitbucket] Connection test failed for %s: %v", integration.Name, err)
		return
	}

	// Sync repositories
	repos, err := client.ListRepositories(ctx)
	if err != nil {
		log.Printf("[Bitbucket] Failed to sync repositories for %s: %v", integration.Name, err)
		return
	}

	// Update last sync time
	integration.UpdatedAt = time.Now()
	if err := s.integrationRepo.Update(&integration); err != nil {
		log.Printf("[Bitbucket] Failed to update integration %s: %v", integration.Name, err)
		return
	}

	log.Printf("[Bitbucket] Synced %d repositories for %s", len(repos), integration.Name)
}

func (s *SyncScheduler) syncJira(ctx context.Context, integration model.Integration) {
	username, _ := integration.Config["username"].(string)
	token, _ := integration.Config["token"].(string)

	if username == "" || token == "" {
		log.Printf("[Jira] No credentials configured for %s", integration.Name)
		return
	}

	client := jira.NewClient(integration.URL, username, token)

	// Test connection
	if err := client.TestConnection(ctx); err != nil {
		log.Printf("[Jira] Connection test failed for %s: %v", integration.Name, err)
		return
	}

	// Sync projects
	projects, err := client.ListProjects(ctx)
	if err != nil {
		log.Printf("[Jira] Failed to sync projects for %s: %v", integration.Name, err)
		return
	}

	// Update last sync time
	integration.UpdatedAt = time.Now()
	if err := s.integrationRepo.Update(&integration); err != nil {
		log.Printf("[Jira] Failed to update integration %s: %v", integration.Name, err)
		return
	}

	log.Printf("[Jira] Synced %d projects for %s", len(projects), integration.Name)
}

func (s *SyncScheduler) syncArgoCD(ctx context.Context, integration model.Integration) {
	token, ok := integration.Config["token"].(string)
	if !ok || token == "" {
		log.Printf("[ArgoCD] No token configured for %s", integration.Name)
		return
	}

	client := argocd.NewClient(integration.URL, token)

	// Test connection
	if err := client.TestConnection(ctx); err != nil {
		log.Printf("[ArgoCD] Connection test failed for %s: %v", integration.Name, err)
		return
	}

	// Sync applications
	apps, err := client.ListApplications(ctx)
	if err != nil {
		log.Printf("[ArgoCD] Failed to sync applications for %s: %v", integration.Name, err)
		return
	}

	// Update last sync time
	integration.UpdatedAt = time.Now()
	if err := s.integrationRepo.Update(&integration); err != nil {
		log.Printf("[ArgoCD] Failed to update integration %s: %v", integration.Name, err)
		return
	}

	log.Printf("[ArgoCD] Synced %d applications for %s", len(apps), integration.Name)
}

func (s *SyncScheduler) syncKubernetes(ctx context.Context, integration model.Integration) {
	token, ok := integration.Config["token"].(string)
	if !ok || token == "" {
		log.Printf("[Kubernetes] No token configured for %s", integration.Name)
		return
	}

	client := kubernetes.NewClient(integration.URL, token)

	// Test connection
	if err := client.TestConnection(ctx); err != nil {
		log.Printf("[Kubernetes] Connection test failed for %s: %v", integration.Name, err)
		return
	}

	// Sync pods from default namespace
	pods, err := client.ListPods(ctx, "default")
	if err != nil {
		log.Printf("[Kubernetes] Failed to sync pods for %s: %v", integration.Name, err)
		return
	}

	// Update last sync time
	integration.UpdatedAt = time.Now()
	if err := s.integrationRepo.Update(&integration); err != nil {
		log.Printf("[Kubernetes] Failed to update integration %s: %v", integration.Name, err)
		return
	}

	log.Printf("[Kubernetes] Synced %d pods for %s", len(pods), integration.Name)
}

func (s *SyncScheduler) syncGrafana(ctx context.Context, integration model.Integration) {
	apiKey, ok := integration.Config["apiKey"].(string)
	if !ok || apiKey == "" {
		log.Printf("[Grafana] No API key configured for %s", integration.Name)
		return
	}

	client := grafana.NewClient(integration.URL, apiKey)

	// Test connection
	if err := client.TestConnection(ctx); err != nil {
		log.Printf("[Grafana] Connection test failed for %s: %v", integration.Name, err)
		return
	}

	// Sync dashboards
	dashboards, err := client.ListDashboards(ctx)
	if err != nil {
		log.Printf("[Grafana] Failed to sync dashboards for %s: %v", integration.Name, err)
		return
	}

	// Update last sync time
	integration.UpdatedAt = time.Now()
	if err := s.integrationRepo.Update(&integration); err != nil {
		log.Printf("[Grafana] Failed to update integration %s: %v", integration.Name, err)
		return
	}

	log.Printf("[Grafana] Synced %d dashboards for %s", len(dashboards), integration.Name)
}

func (s *SyncScheduler) syncGraylog(ctx context.Context, integration model.Integration) {
	username, _ := integration.Config["username"].(string)
	password, _ := integration.Config["password"].(string)

	if username == "" || password == "" {
		log.Printf("[Graylog] No credentials configured for %s", integration.Name)
		return
	}

	client := graylog.NewClient(integration.URL, username, password)

	// Test connection
	if err := client.TestConnection(ctx); err != nil {
		log.Printf("[Graylog] Connection test failed for %s: %v", integration.Name, err)
		return
	}

	// Sync streams
	streams, err := client.ListStreams(ctx)
	if err != nil {
		log.Printf("[Graylog] Failed to sync streams for %s: %v", integration.Name, err)
		return
	}

	// Update last sync time
	integration.UpdatedAt = time.Now()
	if err := s.integrationRepo.Update(&integration); err != nil {
		log.Printf("[Graylog] Failed to update integration %s: %v", integration.Name, err)
		return
	}

	log.Printf("[Graylog] Synced %d streams for %s", len(streams), integration.Name)
}

func (s *SyncScheduler) syncVault(ctx context.Context, integration model.Integration) {
	token, ok := integration.Config["token"].(string)
	if !ok || token == "" {
		log.Printf("[Vault] No token configured for %s", integration.Name)
		return
	}

	client := vault.NewClient(integration.URL, token)

	// Test connection
	if err := client.TestConnection(ctx); err != nil {
		log.Printf("[Vault] Connection test failed for %s: %v", integration.Name, err)
		return
	}

	// Sync secrets
	secrets, err := client.ListSecrets(ctx, "")
	if err != nil {
		log.Printf("[Vault] Failed to sync secrets for %s: %v", integration.Name, err)
		return
	}

	// Update last sync time
	integration.UpdatedAt = time.Now()
	if err := s.integrationRepo.Update(&integration); err != nil {
		log.Printf("[Vault] Failed to update integration %s: %v", integration.Name, err)
		return
	}

	log.Printf("[Vault] Synced %d secrets for %s", len(secrets), integration.Name)
}
