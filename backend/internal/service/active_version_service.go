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

type ActiveVersionService struct {
	repo         *repository.ActiveVersionRepository
	auditRepo    *repository.AuditRepository
	argocdClient *argocd.Client
	gitlabClient *gitlab.Client
}

func NewActiveVersionService(
	repo *repository.ActiveVersionRepository,
	auditRepo *repository.AuditRepository,
	argocdClient *argocd.Client,
	gitlabClient *gitlab.Client,
) *ActiveVersionService {
	return &ActiveVersionService{
		repo:         repo,
		auditRepo:    auditRepo,
		argocdClient: argocdClient,
		gitlabClient: gitlabClient,
	}
}

func (s *ActiveVersionService) Deploy(ctx context.Context, version *model.ActiveVersion, userID, userName string) error {
	version.ID = fmt.Sprintf("ver-%d", time.Now().UnixNano())
	version.DeployedAt = time.Now()
	version.DeployedBy = userName
	version.Status = "active"

	// Сохраняем в БД
	if err := s.repo.Create(version); err != nil {
		return fmt.Errorf("failed to create version in database: %w", err)
	}

	// РЕАЛЬНЫЙ вызов ArgoCD API для деплоя версии
	if s.argocdClient != nil {
		log.Printf("[ActiveVersion] Deploying version in ArgoCD: app=%s, version=%s, traffic=%d%%",
			version.ServiceID, version.Version, version.TrafficPercentage)

		if err := s.argocdClient.DeployVersion(ctx, version.ServiceID, version.Version, version.TrafficPercentage); err != nil {
			// Откатываем запись в БД если ArgoCD вызов упал
			s.repo.Delete(version.ID)
			return fmt.Errorf("failed to deploy version in ArgoCD: %w", err)
		}

		log.Printf("[ActiveVersion] Successfully deployed version in ArgoCD")
	} else {
		log.Printf("[ActiveVersion] WARNING: ArgoCD client not configured, skipping real deployment")
	}

	// РЕАЛЬНЫЙ вызов GitLab API для обновления ветки
	if s.gitlabClient != nil && version.Branch != "" {
		log.Printf("[ActiveVersion] Updating Git branch: %s", version.Branch)

		if err := s.gitlabClient.UpdateBranch(ctx, version.ServiceID, version.Branch, version.GitCommit); err != nil {
			log.Printf("[ActiveVersion] WARNING: Failed to update Git branch: %v", err)
			// Не прерываем процесс, обновление ветки - это опционально
		} else {
			log.Printf("[ActiveVersion] Successfully updated Git branch")
		}
	}

	// Аудит
	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "deploy_version",
		Resource:   "active_version",
		ResourceID: version.ID,
		Details:    fmt.Sprintf("Deployed version %s to %s", version.Version, version.Environment),
	})

	return nil
}

func (s *ActiveVersionService) GetByServiceID(serviceID, environment string) ([]model.ActiveVersion, error) {
	return s.repo.GetByServiceID(serviceID, environment)
}

func (s *ActiveVersionService) UpdateTrafficPercentage(ctx context.Context, id string, percentage int, userID, userName string) error {
	// Получаем version из БД
	version, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	// РЕАЛЬНЫЙ вызов ArgoCD API для обновления трафика
	if s.argocdClient != nil {
		log.Printf("[ActiveVersion] Updating traffic percentage in ArgoCD: app=%s, percentage=%d%%",
			version.ServiceID, percentage)

		if err := s.argocdClient.UpdateTrafficPercentage(ctx, version.ServiceID, percentage); err != nil {
			return fmt.Errorf("failed to update traffic in ArgoCD: %w", err)
		}

		log.Printf("[ActiveVersion] Successfully updated traffic percentage in ArgoCD")
	} else {
		log.Printf("[ActiveVersion] WARNING: ArgoCD client not configured, skipping real traffic update")
	}

	// Обновляем в БД
	if err := s.repo.UpdateTrafficPercentage(id, percentage); err != nil {
		return fmt.Errorf("failed to update percentage in database: %w", err)
	}

	// Аудит
	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "update_traffic",
		Resource:   "active_version",
		ResourceID: id,
		Details:    fmt.Sprintf("Updated traffic percentage to %d%%", percentage),
	})

	return nil
}

func (s *ActiveVersionService) Rollback(ctx context.Context, id, userID, userName string) error {
	// Получаем version из БД
	version, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	// РЕАЛЬНЫЙ вызов ArgoCD API для rollback версии
	if s.argocdClient != nil {
		log.Printf("[ActiveVersion] Rolling back version in ArgoCD: app=%s", version.ServiceID)

		if err := s.argocdClient.RollbackVersion(ctx, version.ServiceID); err != nil {
			return fmt.Errorf("failed to rollback version in ArgoCD: %w", err)
		}

		log.Printf("[ActiveVersion] Successfully rolled back version in ArgoCD")
	} else {
		log.Printf("[ActiveVersion] WARNING: ArgoCD client not configured, skipping real rollback")
	}

	// Обновляем статус в БД
	if err := s.repo.UpdateStatus(id, "rolling_back"); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	if err := s.repo.UpdateStatus(id, "inactive"); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Аудит
	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "rollback_version",
		Resource:   "active_version",
		ResourceID: id,
		Details:    "Rolled back version",
	})

	return nil
}

func (s *ActiveVersionService) Promote(ctx context.Context, id, userID, userName string) error {
	// Get version to find service and environment
	version, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	// РЕАЛЬНЫЙ вызов ArgoCD API для promote версии
	if s.argocdClient != nil {
		log.Printf("[ActiveVersion] Promoting version in ArgoCD: app=%s, version=%s",
			version.ServiceID, version.Version)

		if err := s.argocdClient.PromoteVersion(ctx, version.ServiceID, version.Version); err != nil {
			return fmt.Errorf("failed to promote version in ArgoCD: %w", err)
		}

		log.Printf("[ActiveVersion] Successfully promoted version in ArgoCD")
	} else {
		log.Printf("[ActiveVersion] WARNING: ArgoCD client not configured, skipping real promote")
	}

	// Deactivate all other versions
	if err := s.repo.DeactivateAllExcept(version.ServiceID, version.Environment, id); err != nil {
		return fmt.Errorf("failed to deactivate other versions: %w", err)
	}

	// Set this version to 100% traffic
	if err := s.repo.UpdateTrafficPercentage(id, 100); err != nil {
		return fmt.Errorf("failed to update traffic percentage: %w", err)
	}

	// Аудит
	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		UserName:   userName,
		Action:     "promote_version",
		Resource:   "active_version",
		ResourceID: id,
		Details:    "Promoted version to 100% traffic",
	})

	return nil
}
