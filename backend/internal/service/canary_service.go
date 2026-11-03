package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/company/release-portal/internal/integration/argocd"
	"github.com/company/release-portal/internal/integration/gitlab"
	"github.com/company/release-portal/internal/model"
	"github.com/company/release-portal/internal/repository"
)

type CanaryService struct {
	repo              *repository.CanaryRepository
	auditRepo         *repository.AuditRepository
	releaseRepo       *repository.ReleaseRepository
	serviceRepo       *repository.ServiceRepository
	activeVersionRepo *repository.ActiveVersionRepository
	integrationRepo   *repository.IntegrationRepository
	argocdClient      *argocd.Client // global fallback from startup config
	gitlabClient      *gitlab.Client // global fallback from startup config
}

func NewCanaryService(
	repo *repository.CanaryRepository,
	auditRepo *repository.AuditRepository,
	releaseRepo *repository.ReleaseRepository,
	serviceRepo *repository.ServiceRepository,
	activeVersionRepo *repository.ActiveVersionRepository,
	integrationRepo *repository.IntegrationRepository,
	argocdClient *argocd.Client,
	gitlabClient *gitlab.Client,
) *CanaryService {
	return &CanaryService{
		repo:              repo,
		auditRepo:         auditRepo,
		releaseRepo:       releaseRepo,
		serviceRepo:       serviceRepo,
		activeVersionRepo: activeVersionRepo,
		integrationRepo:   integrationRepo,
		argocdClient:      argocdClient,
		gitlabClient:      gitlabClient,
	}
}

// resolveArgoClient returns an ArgoCD client. Priority: enabled argocd
// integration from DB (so newly added integrations work without restart) ->
// global client initialized at startup.
func (s *CanaryService) resolveArgoClient() (*argocd.Client, error) {
	if s.integrationRepo != nil {
		if integrations, err := s.integrationRepo.GetEnabledByType("argocd"); err == nil {
			for _, integ := range integrations {
				token, _ := integ.Config["token"].(string)
				if token == "" {
					token, _ = integ.Config["apiToken"].(string)
				}
				if integ.URL != "" && token != "" {
					log.Printf("[Canary] Using ArgoCD integration %q (%s)", integ.Name, integ.URL)
					return argocd.NewClient(integ.URL, token), nil
				}
			}
		}
	}
	if s.argocdClient != nil {
		return s.argocdClient, nil
	}
	return nil, errors.New("no enabled ArgoCD integration found — configure it in Settings and sync")
}

// resolveGitLabClient returns a GitLab client for tag creation if available.
func (s *CanaryService) resolveGitLabClient() *gitlab.Client {
	if s.integrationRepo != nil {
		if integrations, err := s.integrationRepo.GetEnabledByType("gitlab"); err == nil {
			for _, integ := range integrations {
				token, _ := integ.Config["token"].(string)
				if token == "" {
					token, _ = integ.Config["apiToken"].(string)
				}
				if integ.URL != "" && token != "" {
					return gitlab.NewClient(integ.URL, token)
				}
			}
		}
	}
	return s.gitlabClient
}

// Create starts a canary deployment for a release:
// 1. validates release/service, resolves stable version and Argo app name;
// 2. creates a real canary in ArgoCD (target revision + traffic annotation + sync);
// 3. persists the canary record and updates active_versions traffic split;
// 4. optionally tags the commit in GitLab;
// 5. writes an audit entry.
func (s *CanaryService) Create(ctx context.Context, canary *model.CanaryDeployment, userID, userName string) error {
	if canary.ReleaseID == "" {
		return errors.New("releaseId is required")
	}

	release, err := s.releaseRepo.GetByID(canary.ReleaseID)
	if err != nil {
		return fmt.Errorf("release not found: %w", err)
	}
	canary.ServiceID = release.ServiceID
	canary.CanaryVersion = release.Version
	canary.CanaryRevision = release.GitCommit

	if existing, err := s.repo.GetActiveByReleaseID(canary.ReleaseID); err == nil && existing != nil {
		return fmt.Errorf("canary deployment already in progress for this release (id=%s, status=%s)", existing.ID, existing.Status)
	}

	// Resolve stable version from active production versions of the service
	if canary.StableVersion == "" {
		if actives, err := s.activeVersionRepo.GetByServiceID(release.ServiceID, "production"); err == nil {
			for _, av := range actives {
				if av.Status == "active" && av.ReleaseID != release.ID {
					canary.StableVersion = av.Version
					canary.StableRevision = av.GitCommit
					break
				}
			}
		}
	}
	if canary.StableVersion == "" {
		canary.StableVersion = "stable"
	}

	// Resolve Argo application name: explicit > service name
	if canary.ArgoApp == "" {
		if svc, err := s.serviceRepo.GetByID(release.ServiceID); err == nil && svc != nil {
			canary.ArgoApp = svc.Name
		}
	}
	if canary.ArgoApp == "" {
		return errors.New("unable to determine ArgoCD application name (service not found and argoApp not provided)")
	}

	if canary.Percentage <= 0 {
		canary.Percentage = 5
	}
	if canary.Percentage > 50 {
		return errors.New("initial canary traffic percentage must be between 1 and 50")
	}

	if canary.GitTag == "" {
		canary.GitTag = fmt.Sprintf("canary-%s-%d", release.Version, time.Now().Unix())
	}

	argoClient, err := s.resolveArgoClient()
	if err != nil {
		return err
	}

	log.Printf("[Canary] Creating canary in ArgoCD: app=%s, version=%s, revision=%s, traffic=%d%%",
		canary.ArgoApp, canary.CanaryVersion, canary.CanaryRevision, canary.Percentage)

	ref := canary.CanaryRevision
	if ref == "" {
		ref = canary.CanaryVersion
	}
	if err := argoClient.CreateCanary(ctx, canary.ArgoApp, ref, canary.Percentage); err != nil {
		return fmt.Errorf("failed to create canary in ArgoCD: %w", err)
	}

	canary.ID = fmt.Sprintf("canary-%d", time.Now().UnixNano())
	canary.StartedAt = time.Now()
	canary.UpdatedAt = time.Now()
	canary.Status = "running"

	if err := s.repo.Create(canary); err != nil {
		// try to undo ArgoCD side effects
		if rbErr := argoClient.RollbackCanary(ctx, canary.ArgoApp); rbErr != nil {
			log.Printf("[Canary] WARNING: DB insert failed and ArgoCD rollback failed too: db=%v argo=%v", err, rbErr)
		}
		return fmt.Errorf("failed to create canary in database: %w", err)
	}

	s.applyTrafficSplit(release.ServiceID, release.ID, canary.Percentage)

	// Optional GitLab tag (non-fatal)
	if glClient := s.resolveGitLabClient(); glClient != nil && canary.CanaryRevision != "" {
		if err := glClient.CreateTag(ctx, release.ServiceID, canary.GitTag, canary.CanaryRevision); err != nil {
			log.Printf("[Canary] WARNING: Failed to create Git tag %s: %v", canary.GitTag, err)
		} else {
			log.Printf("[Canary] Created Git tag %s", canary.GitTag)
		}
	}

	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "create_canary",
		Resource:   "canary_deployment",
		ResourceID: canary.ID,
		Details:    fmt.Sprintf("Created canary deployment for release %s (%s) with %d%% traffic via ArgoCD app %s", canary.CanaryVersion, canary.ReleaseID, canary.Percentage, canary.ArgoApp),
	})

	return nil
}

// applyTrafficSplit records the traffic split in active_versions so the
// catalog reflects how much traffic goes to the canary release.
func (s *CanaryService) applyTrafficSplit(serviceID, releaseID string, canaryPct int) {
	if s.activeVersionRepo == nil {
		return
	}
	actives, err := s.activeVersionRepo.GetByServiceID(serviceID, "production")
	if err != nil {
		return
	}
	for _, av := range actives {
		if av.ReleaseID == releaseID {
			s.activeVersionRepo.UpdateTrafficPercentage(av.ID, canaryPct)
		} else if av.Status == "active" {
			stable := 100 - canaryPct
			if stable < 0 {
				stable = 0
			}
			s.activeVersionRepo.UpdateTrafficPercentage(av.ID, stable)
		}
	}
}

func (s *CanaryService) GetByID(id string) (*model.CanaryDeployment, error) {
	return s.repo.GetByID(id)
}

func (s *CanaryService) GetByReleaseID(releaseID string) (*model.CanaryDeployment, error) {
	return s.repo.GetByReleaseID(releaseID)
}

func (s *CanaryService) GetActiveByReleaseID(releaseID string) (*model.CanaryDeployment, error) {
	return s.repo.GetActiveByReleaseID(releaseID)
}

func (s *CanaryService) GetActiveByServiceID(serviceID string) (*model.CanaryDeployment, error) {
	return s.repo.GetActiveByServiceID(serviceID)
}

func (s *CanaryService) ListByServiceID(serviceID string) ([]model.CanaryDeployment, error) {
	return s.repo.ListByServiceID(serviceID)
}

func (s *CanaryService) ListAll() ([]model.CanaryDeployment, error) {
	return s.repo.ListAll()
}

// UpdatePercentage changes live traffic split in ArgoCD and DB.
func (s *CanaryService) UpdatePercentage(ctx context.Context, id string, percentage int, userID, userName string) error {
	if percentage < 1 || percentage > 99 {
		return errors.New("percentage must be between 1 and 99 (use promote/rollback for 100/0)")
	}

	canary, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get canary: %w", err)
	}
	if canary.Status != "running" && canary.Status != "pending" {
		return fmt.Errorf("canary is not active (status=%s)", canary.Status)
	}

	argoClient, err := s.resolveArgoClient()
	if err != nil {
		return err
	}

	log.Printf("[Canary] Updating traffic percentage in ArgoCD: app=%s, percentage=%d%%", canary.ArgoApp, percentage)
	if err := argoClient.UpdateCanaryTraffic(ctx, canary.ArgoApp, percentage); err != nil {
		return fmt.Errorf("failed to update traffic in ArgoCD: %w", err)
	}

	if err := s.repo.UpdatePercentage(id, percentage); err != nil {
		return fmt.Errorf("failed to update percentage in database: %w", err)
	}

	s.applyTrafficSplit(canary.ServiceID, canary.ReleaseID, percentage)

	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "update_canary_percentage",
		Resource:   "canary_deployment",
		ResourceID: id,
		Details:    fmt.Sprintf("Updated canary traffic to %d%% (app %s)", percentage, canary.ArgoApp),
	})

	return nil
}

// Promote shifts all traffic to the canary version and finalizes it as the
// new active production version.
func (s *CanaryService) Promote(ctx context.Context, id, userID, userName string) error {
	canary, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get canary: %w", err)
	}
	if canary.Status == "promoted" {
		return errors.New("canary already promoted")
	}
	if canary.Status == "rolled_back" {
		return errors.New("canary was rolled back and cannot be promoted")
	}

	argoClient, err := s.resolveArgoClient()
	if err != nil {
		return err
	}

	log.Printf("[Canary] Promoting canary in ArgoCD: app=%s", canary.ArgoApp)
	if err := argoClient.PromoteCanary(ctx, canary.ArgoApp); err != nil {
		return fmt.Errorf("failed to promote canary in ArgoCD: %w", err)
	}

	if err := s.repo.UpdateStatus(id, "promoted"); err != nil {
		return fmt.Errorf("failed to update status in database: %w", err)
	}
	if err := s.repo.UpdatePercentage(id, 100); err != nil {
		log.Printf("[Canary] WARNING: failed to set 100%% in DB: %v", err)
	}

	// Record the promoted version as the new active production version
	if release, err := s.releaseRepo.GetByID(canary.ReleaseID); err == nil {
		newAV := model.ActiveVersion{
			ID:                fmt.Sprintf("av-%d", time.Now().UnixNano()),
			ServiceID:         canary.ServiceID,
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
		if err := s.activeVersionRepo.Create(&newAV); err != nil {
			log.Printf("[Canary] WARNING: failed to record promoted active version: %v", err)
		}
		if err := s.releaseRepo.UpdateStatus(release.ID, "production"); err != nil {
			log.Printf("[Canary] WARNING: failed to mark release as production: %v", err)
		}
	}

	s.applyTrafficSplit(canary.ServiceID, canary.ReleaseID, 100)

	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "promote_canary",
		Resource:   "canary_deployment",
		ResourceID: id,
		Details:    fmt.Sprintf("Promoted canary %s to 100%% traffic (app %s)", canary.CanaryVersion, canary.ArgoApp),
	})

	return nil
}

// Rollback returns all traffic to the stable version and restores the previous
// target revision in ArgoCD.
func (s *CanaryService) Rollback(ctx context.Context, id, userID, userName string) error {
	canary, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get canary: %w", err)
	}
	if canary.Status == "rolled_back" {
		return errors.New("canary already rolled back")
	}

	argoClient, err := s.resolveArgoClient()
	if err != nil {
		return err
	}

	log.Printf("[Canary] Rolling back canary in ArgoCD: app=%s", canary.ArgoApp)

	// Restore previous stable revision first, then cut traffic to 0
	ref := canary.StableRevision
	if ref == "" {
		ref = canary.StableVersion
	}
	if ref != "" && ref != "stable" {
		if err := argoClient.SetTargetRevision(ctx, canary.ArgoApp, ref); err != nil {
			log.Printf("[Canary] WARNING: failed to restore stable revision %s: %v", ref, err)
		}
	}
	if err := argoClient.RollbackCanary(ctx, canary.ArgoApp); err != nil {
		return fmt.Errorf("failed to rollback canary in ArgoCD: %w", err)
	}

	if err := s.repo.UpdateStatus(id, "rolled_back"); err != nil {
		return fmt.Errorf("failed to update status in database: %w", err)
	}
	if err := s.repo.UpdatePercentage(id, 0); err != nil {
		log.Printf("[Canary] WARNING: failed to set 0%% in DB: %v", err)
	}

	s.applyTrafficSplit(canary.ServiceID, canary.ReleaseID, 0)

	if release, err := s.releaseRepo.GetByID(canary.ReleaseID); err == nil {
		if err := s.releaseRepo.UpdateStatus(release.ID, "rolled_back"); err != nil {
			log.Printf("[Canary] WARNING: failed to mark release as rolled back: %v", err)
		}
	}

	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "rollback_canary",
		Resource:   "canary_deployment",
		ResourceID: id,
		Details:    fmt.Sprintf("Rolled back canary %s to stable %s (app %s)", canary.CanaryVersion, canary.StableVersion, canary.ArgoApp),
	})

	return nil
}

// Analyze fetches live sync/health status of the canary application from
// ArgoCD so the UI can show real state instead of just the DB record.
func (s *CanaryService) Analyze(ctx context.Context, id string) (map[string]interface{}, error) {
	canary, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get canary: %w", err)
	}

	result := map[string]interface{}{
		"canary":          canary,
		"argoConnected":   false,
		"syncStatus":      "Unknown",
		"healthStatus":    "Unknown",
		"currentRevision": "",
		"checkedAt":       time.Now(),
	}

	argoClient, err := s.resolveArgoClient()
	if err != nil {
		result["error"] = err.Error()
		return result, nil
	}

	app, err := argoClient.GetApplication(ctx, canary.ArgoApp)
	if err != nil {
		result["error"] = fmt.Sprintf("failed to get app from ArgoCD: %v", err)
		return result, nil
	}

	result["argoConnected"] = true
	result["syncStatus"] = app.Status.Sync.Status
	result["healthStatus"] = app.Status.Health.Status
	result["currentRevision"] = app.Status.Sync.Revision
	result["targetRevision"] = app.Spec.Source.TargetRevision
	if v, ok := app.Metadata.Annotations["canary.argoproj.io/traffic-split"]; ok {
		result["liveTraffic"] = v
	}

	return result, nil
}
