package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/company/release-portal/internal/integration/argocd"
	"github.com/company/release-portal/internal/integration/gitlab"
	"github.com/company/release-portal/internal/model"
	"github.com/company/release-portal/internal/repository"
)

// ErrNoGitOps is returned when a deploy cannot proceed through the GitOps
// chain because an integration or app mapping is missing.
var ErrNoGitOps = errors.New("gitops deployment not possible: configure ArgoCD and GitLab integrations in Settings")

// DeployService implements strict GitOps deployments: the portal NEVER talks
// to Kubernetes directly. A deploy is a two-step pipeline:
//  1. update the desired state (manifest / argocd Application) in the git
//     ops repository via the Git provider API (commit + push through GitLab);
//  2. let ArgoCD reconcile the cluster to that commit — the portal only
//     triggers a sync of the ArgoCD application and waits for health.
type DeployService struct {
	releaseRepo     *repository.ReleaseRepository
	serviceRepo     *repository.ServiceRepository
	activeRepo      *repository.ActiveVersionRepository
	integrationRepo *repository.IntegrationRepository
	auditRepo       *repository.AuditRepository

	argoFallback   *argocd.Client
	gitlabFallback *gitlab.Client
}

func NewDeployService(
	releaseRepo *repository.ReleaseRepository,
	serviceRepo *repository.ServiceRepository,
	activeRepo *repository.ActiveVersionRepository,
	integrationRepo *repository.IntegrationRepository,
	auditRepo *repository.AuditRepository,
	argoFallback *argocd.Client,
	gitlabFallback *gitlab.Client,
) *DeployService {
	return &DeployService{
		releaseRepo:     releaseRepo,
		serviceRepo:     serviceRepo,
		activeRepo:      activeRepo,
		integrationRepo: integrationRepo,
		auditRepo:       auditRepo,
		argoFallback:    argoFallback,
		gitlabFallback:  gitlabFallback,
	}
}

// resolveArgoClient prefers a fresh enabled argocd integration from DB so
// newly added integrations work without restart; falls back to startup client.
func (s *DeployService) resolveArgoClient() (*argocd.Client, error) {
	if s.integrationRepo != nil {
		if integrations, err := s.integrationRepo.GetEnabledByType("argocd"); err == nil {
			for _, integ := range integrations {
				token, _ := integ.Config["token"].(string)
				if token == "" {
					token, _ = integ.Config["apiToken"].(string)
				}
				if integ.URL != "" && token != "" {
					return argocd.NewClient(strings.TrimRight(integ.URL, "/"), token), nil
				}
			}
		}
	}
	if s.argoFallback != nil {
		return s.argoFallback, nil
	}
	return nil, fmt.Errorf("no enabled ArgoCD integration found — %w", ErrNoGitOps)
}

// resolveGitLabClient returns a GitLab client for committing the desired
// state into the git ops repo.
func (s *DeployService) resolveGitLabClient() (*gitlab.Client, error) {
	if s.integrationRepo != nil {
		if integrations, err := s.integrationRepo.GetEnabledByType("gitlab"); err == nil {
			for _, integ := range integrations {
				token, _ := integ.Config["token"].(string)
				if token == "" {
					token, _ = integ.Config["apiToken"].(string)
				}
				if integ.URL != "" && token != "" {
					return gitlab.NewClient(strings.TrimRight(integ.URL, "/"), token), nil
				}
			}
		}
	}
	if s.gitlabFallback != nil {
		return s.gitlabFallback, nil
	}
	return nil, fmt.Errorf("no enabled GitLab integration found — %w", ErrNoGitOps)
}

// GitOpsConfig reads optional gitops repo settings stored on the argocd
// integration config:
//   - gitopsProject: project id/path holding manifests (defaults to the
//     application's own source repo)
//   - gitopsPath:    path inside that repo (default "manifests")
//   - gitopsBranch:  branch to commit to (default "main")
func (s *DeployService) gitopsSettings() (project, path, branch string) {
	project, path, branch = "", "manifests", "main"
	if s.integrationRepo == nil {
		return
	}
	integrations, err := s.integrationRepo.GetEnabledByType("argocd")
	if err != nil {
		return
	}
	for _, integ := range integrations {
		if p, ok := integ.Config["gitopsProject"].(string); ok && strings.TrimSpace(p) != "" {
			project = strings.TrimSpace(p)
		}
		if p, ok := integ.Config["gitopsPath"].(string); ok && strings.TrimSpace(p) != "" {
			path = strings.TrimSpace(p)
		}
		if b, ok := integ.Config["gitopsBranch"].(string); ok && strings.TrimSpace(b) != "" {
			branch = strings.TrimSpace(b)
		}
		if project != "" || path != "" {
			break
		}
	}
	return
}

// DeployRequest describes a GitOps deployment initiated from the portal.
type DeployRequest struct {
	ReleaseID   string `json:"releaseId"`
	Environment string `json:"environment"`
	// AppName optionally overrides the ArgoCD application name.
	AppName string `json:"appName,omitempty"`
}

// DeployResult reports what happened in each GitOps stage.
type DeployResult struct {
	Stage         string `json:"stage"` // git_commit | argocd_sync
	GitOpsProject string `json:"gitops_project,omitempty"`
	GitOpsCommit  string `json:"gitops_commit,omitempty"`
	ArgoApp       string `json:"argo_app"`
	TargetRev     string `json:"target_revision"`
	SyncStatus    string `json:"sync_status,omitempty"`
	HealthStatus  string `json:"health_status,omitempty"`
	Message       string `json:"message"`
}

// Deploy executes the full GitOps loop for a release.
func (s *DeployService) Deploy(ctx context.Context, req DeployRequest, userID, userName string) (*DeployResult, error) {
	if req.ReleaseID == "" {
		return nil, errors.New("releaseId is required")
	}
	env := req.Environment
	if env == "" {
		env = "staging"
	}
	if env != "staging" && env != "production" {
		return nil, fmt.Errorf("invalid environment: %s", env)
	}

	release, err := s.releaseRepo.GetByID(req.ReleaseID)
	if err != nil {
		return nil, fmt.Errorf("release not found: %w", err)
	}
	svc, err := s.serviceRepo.GetByID(release.ServiceID)
	if err != nil || svc == nil {
		return nil, fmt.Errorf("service for release not found")
	}

	argoClient, err := s.resolveArgoClient()
	if err != nil {
		return nil, err
	}

	// --- Stage 0: locate the ArgoCD application (GitOps source of truth) ----
	appName := strings.TrimSpace(req.AppName)
	var app *argocd.Application
	if appName != "" {
		app, _ = argoClient.GetApplication(ctx, appName)
	}
	if app == nil {
		// search by service name, then by repo URL match
		apps, lerr := argoClient.ListApplications(ctx)
		if lerr != nil {
			return nil, fmt.Errorf("failed to list ArgoCD applications: %w", lerr)
		}
		for i := range apps {
			if apps[i].Metadata.Name == svc.Name {
				app = &apps[i]
				break
			}
		}
		if app == nil && svc.Repository != "" {
			for i := range apps {
				if apps[i].Spec.Source.RepoURL != "" && apps[i].Spec.Source.RepoURL == svc.Repository {
					app = &apps[i]
					break
				}
			}
		}
		if app == nil {
			names := make([]string, 0, len(apps))
			for _, a := range apps {
				names = append(names, a.Metadata.Name)
			}
			return nil, fmt.Errorf("no ArgoCD application matches service %q (available: %s). Set the app name explicitly or sync the ArgoCD integration", svc.Name, strings.Join(names, ", "))
		}
	}
	appName = app.Metadata.Name

	targetRev := release.GitCommit
	if targetRev == "" {
		targetRev = release.Branch
	}
	if targetRev == "" {
		targetRev = release.Version
	}

	result := &DeployResult{
		Stage:     "argocd_sync",
		ArgoApp:   appName,
		TargetRev: targetRev,
	}

	// --- Stage 1: commit the desired state into the git ops repo ------------
	// In pure GitOps the cluster state changes ONLY via a git commit. We use
	// the GitLab API as the transport for that commit (UpdateBranch ->
	// files/{path}/kustomization.yaml PUT with branch=main). If no gitops
	// repo is configured we fall back to pointing the ArgoCD application at
	// the release revision directly — still declarative, applied by ArgoCD.
	glClient, glErr := s.resolveGitLabClient()
	gitopsProject, gitopsPath, gitopsBranch := s.gitopsSettings()
	if gitopsProject == "" && app.Spec.Source.RepoURL != "" {
		gitopsProject = gitopsProjectFromRepoURL(app.Spec.Source.RepoURL)
	}
	if glErr == nil && gitopsProject != "" {
		manifest := renderManifest(svc.Name, appName, release, env, targetRev, app)
		filePath := fmt.Sprintf("%s/%s/%s/deployment.yaml", strings.Trim(gitopsPath, "/"), svc.Name, env)
		if err := glClient.UpdateFile(ctx, gitopsProject, filePath, gitopsBranch, manifest,
			fmt.Sprintf("Deploy %s %s (%s) to %s via Release Portal", svc.Name, release.Version, targetRev, env)); err != nil {
			log.Printf("[Deploy] WARNING: gitops commit failed (%v); falling back to direct ArgoCD targetRevision update", err)
		} else {
			result.Stage = "git_commit"
			result.GitOpsProject = gitopsProject
			log.Printf("[Deploy] Committed desired state to %s:%s (%s)", gitopsProject, filePath, gitopsBranch)
		}
	}

	// --- Stage 2: ArgoCD reconciles the cluster to the git state ------------
	if result.Stage != "git_commit" {
		// Declarative fallback: pin the app source to the release revision.
		if err := argoClient.SetTargetRevision(ctx, appName, targetRev); err != nil {
			return nil, fmt.Errorf("failed to set target revision in ArgoCD: %w", err)
		}
	}
	if err := argoClient.SyncApplication(ctx, appName); err != nil {
		return nil, fmt.Errorf("failed to trigger ArgoCD sync: %w", err)
	}

	// Wait briefly for ArgoCD to report sync/health (best effort).
	syncStatus, healthStatus := "Syncing", "Progressing"
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second)
		if cur, err := argoClient.GetApplication(ctx, appName); err == nil && cur != nil {
			syncStatus = cur.Status.Sync.Status
			healthStatus = cur.Status.Health.Status
			if syncStatus == "Synced" && (healthStatus == "Healthy" || healthStatus == "Degraded") {
				break
			}
		}
	}
	result.SyncStatus = syncStatus
	result.HealthStatus = healthStatus
	result.Message = fmt.Sprintf("GitOps deploy of %s %s to %s via app %s: sync=%s health=%s",
		svc.Name, release.Version, env, appName, syncStatus, healthStatus)

	// --- Bookkeeping (portal DB mirrors reality, it does not drive it) ------
	newStatus := "staging"
	if env == "production" {
		newStatus = "production"
	}
	if err := s.releaseRepo.UpdateStatus(release.ID, newStatus); err != nil {
		log.Printf("[Deploy] WARNING: failed to update release status: %v", err)
	}
	if env == "production" && s.activeRepo != nil {
		av := model.ActiveVersion{
			ID:                fmt.Sprintf("av-%d", time.Now().UnixNano()),
			ServiceID:         svc.ID,
			ReleaseID:         release.ID,
			Version:           release.Version,
			Branch:            release.Branch,
			GitCommit:         release.GitCommit,
			Environment:       "production",
			TrafficPercentage: 100,
			Status:            "active",
			DeployedAt:        time.Now(),
			DeployedBy:        userName,
		}
		if err := s.activeRepo.Create(&av); err != nil {
			log.Printf("[Deploy] WARNING: failed to record active version: %v", err)
		}
	}

	if s.auditRepo != nil {
		s.auditRepo.Create(&model.AuditEntry{
			ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
			UserID:     userID,
			UserName:   userName,
			Action:     "deploy_gitops",
			Resource:   "release",
			ResourceID: release.ID,
			Details:    result.Message,
		})
	}

	return result, nil
}

// Rollback reverts an ArgoCD application to its previous synced revision —
// again entirely through GitOps tooling (Argo rollback), never kubectl.
func (s *DeployService) Rollback(ctx context.Context, releaseID, userID, userName string) (*DeployResult, error) {
	release, err := s.releaseRepo.GetByID(releaseID)
	if err != nil {
		return nil, fmt.Errorf("release not found: %w", err)
	}
	svc, err := s.serviceRepo.GetByID(release.ServiceID)
	if err != nil || svc == nil {
		return nil, errors.New("service for release not found")
	}
	argoClient, err := s.resolveArgoClient()
	if err != nil {
		return nil, err
	}

	appName := svc.Name
	if app, err := argoClient.GetApplication(ctx, appName); err != nil || app == nil {
		if apps, lerr := argoClient.ListApplications(ctx); lerr == nil {
			found := false
			for i := range apps {
				if apps[i].Metadata.Name == svc.Name {
					appName = apps[i].Metadata.Name
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("no ArgoCD application found for service %q", svc.Name)
			}
		}
	}

	if err := argoClient.RollbackCanary(ctx, appName); err != nil {
		return nil, fmt.Errorf("failed to rollback via ArgoCD: %w", err)
	}
	if err := s.releaseRepo.UpdateStatus(releaseID, "rolled_back"); err != nil {
		log.Printf("[Deploy] WARNING: failed to update release status: %v", err)
	}
	if s.auditRepo != nil {
		s.auditRepo.Create(&model.AuditEntry{
			ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
			UserID:     userID,
			UserName:   userName,
			Action:     "rollback_gitops",
			Resource:   "release",
			ResourceID: releaseID,
			Details:    fmt.Sprintf("Rolled back %s via ArgoCD app %s", release.Version, appName),
		})
	}
	return &DeployResult{
		Stage:   "argocd_sync",
		ArgoApp: appName,
		Message: fmt.Sprintf("Rollback triggered through ArgoCD app %s", appName),
	}, nil
}

// gitopsProjectFromRepoURL converts https://host/group/sub/repo(.git) into
// url-encoded group/sub/repo for the GitLab API.
func gitopsProjectFromRepoURL(repoURL string) string {
	path := repoURL
	if u, err := url.Parse(repoURL); err == nil && u.Host != "" {
		path = strings.Trim(u.Path, "/")
	} else if i := strings.Index(path, ":"); i >= 0 && !strings.Contains(path[:i], "/") {
		path = path[i+1:] // git@host:group/repo.git
	}
	path = strings.TrimSuffix(path, ".git")
	if path == "" {
		return ""
	}
	return url.PathEscape(path)
}

// renderManifest produces the desired-state document committed to the gitops
// repo. It pins the app source to the release revision and records the
// environment/version in annotations so ArgoCD diffs are meaningful.
func renderManifest(serviceName, appName string, release *model.Release, env, targetRev string, app *argocd.Application) string {
	repoURL, srcPath := "", ""
	if app != nil {
		repoURL = app.Spec.Source.RepoURL
		srcPath = app.Spec.Source.Path
	}
	return fmt.Sprintf(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: %[2]s
  annotations:
    release-portal.company/service: %[1]s
    release-portal.company/environment: %[4]s
    release-portal.company/version: %[3]s
spec:
  source:
    repoURL: %[5]s
    path: %[6]s
    targetRevision: %[7]s
  destination:
    server: %[8]s
    namespace: %[9]s
`, serviceName, appName, release.Version, env, repoURL, srcPath,
		targetRev, destOr(app, "https://kubernetes.default.svc"), destNamespaceOr(app, env))
}

func destOr(a *argocd.Application, def string) string {
	if a != nil && a.Spec.Destination.Server != "" {
		return a.Spec.Destination.Server
	}
	return def
}

func destNamespaceOr(a *argocd.Application, def string) string {
	if a != nil && a.Spec.Destination.Namespace != "" {
		return a.Spec.Destination.Namespace
	}
	return def
}
