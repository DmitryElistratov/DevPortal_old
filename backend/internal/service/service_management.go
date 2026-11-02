package service

import (
	"fmt"
	"time"

	"github.com/company/release-portal/internal/model"
	"github.com/company/release-portal/internal/repository"
)

type ServiceManagementService struct {
	repo      *repository.ServiceRepository
	auditRepo *repository.AuditRepository
}

func NewServiceManagementService(repo *repository.ServiceRepository, auditRepo *repository.AuditRepository) *ServiceManagementService {
	return &ServiceManagementService{repo: repo, auditRepo: auditRepo}
}

func (s *ServiceManagementService) Create(svc *model.Service, userID, userName string) error {
	svc.ID = fmt.Sprintf("svc-%d", generateTimestamp())
	if err := s.repo.Create(svc); err != nil {
		return err
	}

	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", generateTimestamp()),
		UserID:     userID,
		UserName:   userName,
		Action:     "create_service",
		Resource:   "service",
		ResourceID: svc.ID,
		Details:    fmt.Sprintf("Created service %s", svc.Name),
	})

	return nil
}

func (s *ServiceManagementService) GetByID(id string) (*model.Service, error) {
	return s.repo.GetByID(id)
}

func (s *ServiceManagementService) List() ([]model.Service, error) {
	return s.repo.List()
}

func (s *ServiceManagementService) Update(svc *model.Service) error {
	return s.repo.Update(svc)
}

func (s *ServiceManagementService) Delete(id string) error {
	return s.repo.Delete(id)
}

func generateTimestamp() int64 {
	return int64(time.Now().UnixNano() / 1e6)
}
