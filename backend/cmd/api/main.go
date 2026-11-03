package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/company/release-portal/internal/config"
	"github.com/company/release-portal/internal/handler"
	"github.com/company/release-portal/internal/integration/argocd"
	"github.com/company/release-portal/internal/integration/gitlab"
	"github.com/company/release-portal/internal/integration/grafana"
	"github.com/company/release-portal/internal/integration/graylog"
	"github.com/company/release-portal/internal/integration/jira"
	"github.com/company/release-portal/internal/integration/kubernetes"
	"github.com/company/release-portal/internal/integration/vault"
	"github.com/company/release-portal/internal/middleware"
	"github.com/company/release-portal/internal/repository"
	"github.com/company/release-portal/internal/scheduler"
	"github.com/company/release-portal/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Database
	db, err := repository.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
	releaseRepo := repository.NewReleaseRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	integrationRepo := repository.NewIntegrationRepository(db)
	activeVersionRepo := repository.NewActiveVersionRepository(db)
	canaryRepo := repository.NewCanaryRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)

	// Initialize integration clients from database
	var argocdClient *argocd.Client
	var gitlabClient *gitlab.Client
	var jiraClient *jira.Client
	var kubernetesClient *kubernetes.Client
	var grafanaClient *grafana.Client
	var graylogClient *graylog.Client
	var vaultClient *vault.Client

	// Load ArgoCD integration
	if argoIntegrations, err := integrationRepo.GetEnabledByType("argocd"); err == nil && len(argoIntegrations) > 0 {
		integration := argoIntegrations[0]
		if token, ok := integration.Config["token"].(string); ok {
			argocdClient = argocd.NewClient(integration.URL, token)
			log.Printf("[Integration] ArgoCD client initialized: %s", integration.URL)
		}
	}
	log.Printf("[Integration] ArgoCD client status: %v", argocdClient != nil)

	// Load GitLab integration
	if gitlabIntegrations, err := integrationRepo.GetEnabledByType("gitlab"); err == nil && len(gitlabIntegrations) > 0 {
		integration := gitlabIntegrations[0]
		if token, ok := integration.Config["token"].(string); ok {
			gitlabClient = gitlab.NewClient(integration.URL, token)
			log.Printf("[Integration] GitLab client initialized: %s", integration.URL)
		}
	}
	log.Printf("[Integration] GitLab client status: %v", gitlabClient != nil)

	// Load Jira integration
	if jiraIntegrations, err := integrationRepo.GetEnabledByType("jira"); err == nil && len(jiraIntegrations) > 0 {
		integration := jiraIntegrations[0]
		username, _ := integration.Config["username"].(string)
		token, _ := integration.Config["token"].(string)
		if username != "" && token != "" {
			jiraClient = jira.NewClient(integration.URL, username, token)
			log.Printf("[Integration] Jira client initialized: %s", integration.URL)
		}
	}
	log.Printf("[Integration] Jira client status: %v", jiraClient != nil)

	// Load Kubernetes integration
	if k8sIntegrations, err := integrationRepo.GetEnabledByType("kubernetes"); err == nil && len(k8sIntegrations) > 0 {
		integration := k8sIntegrations[0]
		if token, ok := integration.Config["token"].(string); ok {
			kubernetesClient = kubernetes.NewClient(integration.URL, token)
			log.Printf("[Integration] Kubernetes client initialized: %s", integration.URL)
		}
	}
	log.Printf("[Integration] Kubernetes client status: %v", kubernetesClient != nil)

	// Load Grafana integration
	if grafanaIntegrations, err := integrationRepo.GetEnabledByType("grafana"); err == nil && len(grafanaIntegrations) > 0 {
		integration := grafanaIntegrations[0]
		if apiKey, ok := integration.Config["apiKey"].(string); ok {
			grafanaClient = grafana.NewClient(integration.URL, apiKey)
			log.Printf("[Integration] Grafana client initialized: %s", integration.URL)
		}
	}
	log.Printf("[Integration] Grafana client status: %v", grafanaClient != nil)

	// Load Graylog integration
	if graylogIntegrations, err := integrationRepo.GetEnabledByType("graylog"); err == nil && len(graylogIntegrations) > 0 {
		integration := graylogIntegrations[0]
		username, _ := integration.Config["username"].(string)
		password, _ := integration.Config["password"].(string)
		if username != "" && password != "" {
			graylogClient = graylog.NewClient(integration.URL, username, password)
			log.Printf("[Integration] Graylog client initialized: %s", integration.URL)
		}
	}
	log.Printf("[Integration] Graylog client status: %v", graylogClient != nil)

	// Load Vault integration
	if vaultIntegrations, err := integrationRepo.GetEnabledByType("vault"); err == nil && len(vaultIntegrations) > 0 {
		integration := vaultIntegrations[0]
		if token, ok := integration.Config["token"].(string); ok {
			vaultClient = vault.NewClient(integration.URL, token)
			log.Printf("[Integration] Vault client initialized: %s", integration.URL)
		}
	}
	log.Printf("[Integration] Vault client status: %v", vaultClient != nil)

	// Services
	authService := service.NewAuthService(userRepo, permissionRepo, cfg.JWTSecret)
	serviceMgmtService := service.NewServiceManagementService(serviceRepo, auditRepo)
	releaseService := service.NewReleaseService(releaseRepo, auditRepo)
	activeVersionService := service.NewActiveVersionService(activeVersionRepo, auditRepo, argocdClient, gitlabClient)
	canaryService := service.NewCanaryService(canaryRepo, auditRepo, releaseRepo, serviceRepo, activeVersionRepo, integrationRepo, argocdClient, gitlabClient)
	jiraService := service.NewJiraService(releaseRepo, auditRepo, jiraClient)
	kubernetesService := service.NewKubernetesService(kubernetesClient)
	grafanaService := service.NewGrafanaService(grafanaClient)
	graylogService := service.NewGraylogService(graylogClient)
	vaultService := service.NewVaultService(vaultClient)
	vaultService.WithIntegrationRepo(integrationRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	serviceHandler := handler.NewServiceHandler(serviceMgmtService)
	releaseHandler := handler.NewReleaseHandler(releaseService)
	auditHandler := handler.NewAuditHandler(auditRepo)
	integrationHandler := handler.NewIntegrationHandler(
		integrationRepo, gitlabClient,
		handler.WithServiceRepo(serviceRepo),
		handler.WithUserRepo(userRepo),
		handler.WithPermRepo(permissionRepo),
	)
	activeVersionHandler := handler.NewActiveVersionHandler(activeVersionService)
	canaryHandler := handler.NewCanaryHandler(canaryService)
	jiraHandler := handler.NewJiraHandler(jiraService)
	kubernetesHandler := handler.NewKubernetesHandler(kubernetesService)
	grafanaHandler := handler.NewGrafanaHandler(grafanaService)
	graylogHandler := handler.NewGraylogHandler(graylogService)
	vaultHandler := handler.NewVaultHandler(vaultService)

	// Router
	router := gin.Default()
	router.Use(gin.Recovery())

	// CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Public routes
	auth := router.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/oauth", authHandler.OAuthLogin)
	}

	// Public list of enabled auth providers so the login page (unauthenticated)
	// can render "Sign in with GitLab/GitHub/Keycloak" buttons. Only safe
	// fields are exposed — no client secrets or tokens.
	router.GET("/api/auth-providers", func(c *gin.Context) {
		integrations, err := integrationRepo.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		type publicProvider struct {
			Name     string `json:"name"`
			Provider string `json:"provider"`
			URL      string `json:"url"`
			ClientID string `json:"clientId"`
			Realm    string `json:"realm"`
		}
		out := make([]publicProvider, 0)
		for _, it := range integrations {
			if it.Type != "auth" || !it.Enabled {
				continue
			}
			provider, _ := it.Config["provider"].(string)
			if provider == "" || provider == "ldap" {
				continue
			}
			clientID, _ := it.Config["clientId"].(string)
			if clientID == "" {
				continue
			}
			realm, _ := it.Config["realm"].(string)
			out = append(out, publicProvider{
				Name:     it.Name,
				Provider: provider,
				URL:      strings.TrimRight(it.URL, "/"),
				ClientID: clientID,
				Realm:    realm,
			})
		}
		c.JSON(http.StatusOK, out)
	})

	// Protected routes
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Services
		services := api.Group("/services")
		{
			services.GET("", serviceHandler.List)
			services.GET("/:id", serviceHandler.Get)
			services.POST("", middleware.RequireRole("admin"), serviceHandler.Create)
			services.PUT("/:id", middleware.RequireRole("admin"), serviceHandler.Update)
			services.DELETE("/:id", middleware.RequireRole("admin"), serviceHandler.Delete)
		}

		// Releases
		releases := api.Group("/releases")
		{
			releases.GET("", releaseHandler.List)
			releases.GET("/:id", releaseHandler.Get)
			releases.POST("", releaseHandler.Create)
			releases.PUT("/:id", releaseHandler.Update)
			releases.POST("/:id/deploy", releaseHandler.Deploy)
			releases.POST("/:id/rollback", releaseHandler.Rollback)
			releases.POST("/:id/approve", releaseHandler.Approve)
		}

		// Audit
		audit := api.Group("/audit")
		{
			audit.GET("", auditHandler.List)
		}

		// User
		api.GET("/me", authHandler.Me)
		api.GET("/users", middleware.RequireRole("admin"), authHandler.ListUsers)

		// Integrations
		integrations := api.Group("/integrations")
		{
			integrations.GET("", integrationHandler.List)
			integrations.POST("", middleware.RequireRole("admin"), integrationHandler.Create)
			integrations.GET("/type/:type", integrationHandler.GetByType)
			integrations.GET("/:id", integrationHandler.Get)
			integrations.PUT("/:id", middleware.RequireRole("admin"), integrationHandler.Update)
			integrations.DELETE("/:id", middleware.RequireRole("admin"), integrationHandler.Delete)
			integrations.POST("/:id/test", integrationHandler.Test)
			integrations.POST("/:id/sync", integrationHandler.Sync)
		}

		// Active Versions
		activeVersions := api.Group("/active-versions")
		{
			activeVersions.POST("", activeVersionHandler.Deploy)
			activeVersions.GET("/service/:serviceId", activeVersionHandler.GetByService)
			activeVersions.PUT("/:versionId/traffic", activeVersionHandler.UpdateTraffic)
			activeVersions.POST("/:versionId/rollback", activeVersionHandler.Rollback)
			activeVersions.POST("/:versionId/promote", activeVersionHandler.Promote)
		}

		// Canary Deployments
		canary := api.Group("/canary")
		{
			canary.POST("", canaryHandler.Create)
			canary.GET("", canaryHandler.ListAll)
			canary.GET("/release/:releaseId/active", canaryHandler.GetActiveByRelease)
			canary.GET("/release/:releaseId", canaryHandler.GetByRelease)
			canary.GET("/service/:serviceId/active", canaryHandler.GetActive)
			canary.GET("/service/:serviceId/list", canaryHandler.List)
			canary.GET("/:id", canaryHandler.Get)
			canary.GET("/:id/analyze", canaryHandler.Analyze)
			canary.PUT("/:id/percentage", canaryHandler.UpdatePercentage)
			canary.POST("/:id/promote", canaryHandler.Promote)
			canary.POST("/:id/rollback", canaryHandler.Rollback)
		}

		// Jira Integration
		jira := api.Group("/jira")
		{
			jira.GET("/issues/:key", jiraHandler.GetIssue)
			jira.POST("/issues/search", jiraHandler.SearchIssues)
			jira.POST("/issues", jiraHandler.CreateIssue)
			jira.POST("/issues/:key/transitions", jiraHandler.TransitionIssue)
			jira.POST("/issues/:key/link", jiraHandler.LinkIssueToRelease)
		}

		// Kubernetes Integration
		k8s := api.Group("/kubernetes")
		{
			k8s.GET("/pods", kubernetesHandler.ListPods)
			k8s.GET("/deployments", kubernetesHandler.ListDeployments)
			k8s.GET("/services", kubernetesHandler.ListServices)
			k8s.GET("/pods/:name/logs", kubernetesHandler.GetPodLogs)
		}

		// Grafana Integration
		grafana := api.Group("/grafana")
		{
			grafana.GET("/dashboards", grafanaHandler.ListDashboards)
			grafana.GET("/dashboards/:uid", grafanaHandler.GetDashboard)
			grafana.GET("/datasources", grafanaHandler.ListDatasources)
			grafana.POST("/query", grafanaHandler.QueryMetrics)
		}

		// Graylog Integration
		graylog := api.Group("/graylog")
		{
			graylog.POST("/search", graylogHandler.SearchLogs)
			graylog.GET("/streams", graylogHandler.ListStreams)
			graylog.GET("/streams/:id", graylogHandler.GetStream)
		}

		// Vault Integration
		vault := api.Group("/vault")
		{
			// NOTE: use a wildcard (*path) instead of ":path" — Gin treats ":" params as
			// single-segment, so nested secret paths like /vault/secrets/secret/my/database
			// returned 404. Handlers must strip the leading "/" from c.Param("path").
			vault.GET("/secrets", vaultHandler.ListSecrets)
			vault.POST("/secrets/*path", func(c *gin.Context) {
				p := strings.TrimPrefix(c.Param("path"), "/")
				if strings.HasSuffix(p, "/restore") {
					vaultHandler.RestoreVersion(c, strings.TrimSuffix(p, "/restore"))
					return
				}
				if strings.HasSuffix(p, "/versions") {
					vaultHandler.ListVersions(c, strings.TrimSuffix(p, "/versions"))
					return
				}
				vaultHandler.WriteSecret(c)
			})
			vault.PUT("/secrets/*path", vaultHandler.UpdateSecret)
			vault.DELETE("/secrets/*path", vaultHandler.DeleteSecret)
			vault.GET("/secrets/*path", func(c *gin.Context) {
				p := strings.TrimPrefix(c.Param("path"), "/")
				if strings.HasSuffix(p, "/versions") {
					vaultHandler.ListVersions(c, strings.TrimSuffix(p, "/versions"))
					return
				}
				vaultHandler.GetSecret(c)
			})
			vault.GET("/health", vaultHandler.Health)
		}
	}

	// Start sync scheduler
	syncScheduler := scheduler.NewSyncScheduler(integrationRepo)
	syncCtx, syncCancel := context.WithCancel(context.Background())
	defer syncCancel()
	go syncScheduler.Start(syncCtx)
	log.Println("Sync scheduler started (interval: 1 minute)")

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
