package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/company/release-portal/internal/integration/argocd"
	"github.com/company/release-portal/internal/integration/gitlab"
	"github.com/company/release-portal/internal/model"
	"github.com/company/release-portal/internal/repository"
)

type CanaryService struct {
	repo         *repository.CanaryRepository
	auditRepo    *repository.AuditRepository
	argocdClient *argocd.Client
	gitlabClient *gitlab.Client
}

func NewCanaryService(
	repo *repository.CanaryRepository,
	auditRepo *repository.AuditRepository,
	argocdClient *argocd.Client,
	gitlabClient *gitlab.Client,
) *CanaryService {
	return &CanaryService{
		repo:         repo,
		auditRepo:    auditRepo,
		argocdClient: argocdClient,
		gitlabClient: gitlabClient,
	}
}

func (s *CanaryService) Create(ctx context.Context, canary *model.CanaryDeployment, userID, userName string) error {
	canary.ID = fmt.Sprintf("canary-%d", time.Now().UnixNano())
	canary.StartedAt = time.Now()
	canary.UpdatedAt = time.Now()
	canary.Status = "running"

	// Сохраняем в БД
	if err := s.repo.Create(canary); err != nil {
		return fmt.Errorf("failed to create canary in database: %w", err)
	}

	// РЕАЛЬНЫЙ вызов ArgoCD API для создания canary deployment
	if s.argocdClient != nil {
		log.Printf("[Canary] Creating canary deployment in ArgoCD: app=%s, version=%s, percentage=%d%%",
			canary.ArgoApp, canary.CanaryVersion, canary.Percentage)

		if err := s.argocdClient.CreateCanary(ctx, canary.ArgoApp, canary.CanaryVersion, canary.Percentage); err != nil {
			// Откатываем запись в БД если ArgoCD вызов упал
			s.repo.Delete(canary.ID)
			return fmt.Errorf("failed to create canary in ArgoCD: %w", err)
		}

		log.Printf("[Canary] Successfully created canary deployment in ArgoCD")
	} else {
		log.Printf("[Canary] WARNING: ArgoCD client not configured, skipping real deployment")
	}

	// РЕАЛЬНЫЙ вызов GitLab API для создания тега
	if s.gitlabClient != nil && canary.GitTag != "" {
		log.Printf("[Canary] Creating Git tag: %s", canary.GitTag)

		if err := s.gitlabClient.CreateTag(ctx, canary.ServiceID, canary.GitTag, canary.CanaryRevision); err != nil {
			log.Printf("[Canary] WARNING: Failed to create Git tag: %v", err)
			// Не прерываем процесс, тег - это опционально
		} else {
			log.Printf("[Canary] Successfully created Git tag")
		}
	}

	// Аудит
	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "create_canary",
		Resource:   "canary_deployment",
		ResourceID: canary.ID,
		Details:    fmt.Sprintf("Created canary deployment for release %s with %d%% traffic", canary.CanaryVersion, canary.Percentage),
	})

	return nil
}

func (s *CanaryService) GetByID(id string) (*model.CanaryDeployment, error) {
	return s.repo.GetByID(id)
}

func (s *CanaryService) GetByReleaseID(releaseID string) (*model.CanaryDeployment, error) {
	return s.repo.GetByReleaseID(releaseID)
}

func (s *CanaryService) GetActiveByServiceID(serviceID string) (*model.CanaryDeployment, error) {
	return s.repo.GetActiveByServiceID(serviceID)
}

func (s *CanaryService) ListByServiceID(serviceID string) ([]model.CanaryDeployment, error) {
	return s.repo.ListByServiceID(serviceID)
}

func (s *CanaryService) UpdatePercentage(ctx context.Context, id string, percentage int, userID, userName string) error {
	// Получаем canary из БД
	canary, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get canary: %w", err)
	}

	// РЕАЛЬНЫЙ вызов ArgoCD API для обновления трафика
	if s.argocdClient != nil {
		log.Printf("[Canary] Updating traffic percentage in ArgoCD: app=%s, percentage=%d%%",
			canary.ArgoApp, percentage)

		if err := s.argocdClient.UpdateCanaryTraffic(ctx, canary.ArgoApp, percentage); err != nil {
			return fmt.Errorf("failed to update traffic in ArgoCD: %w", err)
		}

		log.Printf("[Canary] Successfully updated traffic percentage in ArgoCD")
	} else {
		log.Printf("[Canary] WARNING: ArgoCD client not configured, skipping real traffic update")
	}

	// Обновляем в БД
	if err := s.repo.UpdatePercentage(id, percentage); err != nil {
		return fmt.Errorf("failed to update percentage in database: %w", err)
	}

	// Аудит
	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "update_canary_percentage",
		Resource:   "canary_deployment",
		ResourceID: id,
		Details:    fmt.Sprintf("Updated canary traffic to %d%%", percentage),
	})

	return nil
}

func (s *CanaryService) Promote(ctx context.Context, id, userID, userName string) error {
	// Получаем canary из БД
	canary, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get canary: %w", err)
	}

	// РЕАЛЬНЫЙ вызов ArgoCD API для promote canary
	if s.argocdClient != nil {
		log.Printf("[Canary] Promoting canary in ArgoCD: app=%s", canary.ArgoApp)

		if err := s.argocdClient.PromoteCanary(ctx, canary.ArgoApp); err != nil {
			return fmt.Errorf("failed to promote canary in ArgoCD: %w", err)
		}

		log.Printf("[Canary] Successfully promoted canary in ArgoCD")
	} else {
		log.Printf("[Canary] WARNING: ArgoCD client not configured, skipping real promote")
	}

	// Обновляем статус в БД
	if err := s.repo.UpdateStatus(id, "promoted"); err != nil {
		return fmt.Errorf("failed to update status in database: %w", err)
	}

	// Аудит
	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "promote_canary",
		Resource:   "canary_deployment",
		ResourceID: id,
		Details:    "Promoted canary deployment to production",
	})

	return nil
}

func (s *CanaryService) Rollback(ctx context.Context, id, userID, userName string) error {
	// Получаем canary из БД
	canary, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get canary: %w", err)
	}

	// РЕАЛЬНЫЙ вызов ArgoCD API для rollback canary
	if s.argocdClient != nil {
		log.Printf("[Canary] Rolling back canary in ArgoCD: app=%s", canary.ArgoApp)

		if err := s.argocdClient.RollbackCanary(ctx, canary.ArgoApp); err != nil {
			return fmt.Errorf("failed to rollback canary in ArgoCD: %w", err)
		}

		log.Printf("[Canary] Successfully rolled back canary in ArgoCD")
	} else {
		log.Printf("[Canary] WARNING: ArgoCD client not configured, skipping real rollback")
	}

	// Обновляем статус в БД
	if err := s.repo.UpdateStatus(id, "rolled_back"); err != nil {
		return fmt.Errorf("failed to update status in database: %w", err)
	}

	// Аудит
	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "rollback_canary",
		Resource:   "canary_deployment",
		ResourceID: id,
		Details:    "Rolled back canary deployment",
	})

	return nil
}
