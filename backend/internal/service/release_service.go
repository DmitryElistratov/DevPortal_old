package service

import (
	"fmt"

	"github.com/company/release-portal/internal/model"
	"github.com/company/release-portal/internal/repository"
)

type ReleaseService struct {
	repo      *repository.ReleaseRepository
	auditRepo *repository.AuditRepository
}

func NewReleaseService(repo *repository.ReleaseRepository, auditRepo *repository.AuditRepository) *ReleaseService {
	return &ReleaseService{repo: repo, auditRepo: auditRepo}
}

func (s *ReleaseService) Create(rel *model.Release, userID, userName string) error {
	rel.ID = fmt.Sprintf("rel-%d", generateTimestamp())
	if err := s.repo.Create(rel); err != nil {
		return err
	}

	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", generateTimestamp()),
		UserID:     userID,
		UserName:   userName,
		Action:     "create_release",
		Resource:   "release",
		ResourceID: rel.ID,
		Details:    fmt.Sprintf("Created release %s for service %s", rel.Version, rel.ServiceID),
	})

	return nil
}

func (s *ReleaseService) GetByID(id string) (*model.Release, error) {
	return s.repo.GetByID(id)
}

func (s *ReleaseService) List(serviceID string) ([]model.Release, error) {
	return s.repo.List(serviceID)
}

func (s *ReleaseService) Update(rel *model.Release) error {
	return s.repo.Update(rel)
}

func (s *ReleaseService) Deploy(id, environment, userID, userName string) error {
	rel, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	var newStatus string
	switch environment {
	case "staging":
		newStatus = "staging"
	case "production":
		newStatus = "production"
	default:
		return fmt.Errorf("invalid environment: %s", environment)
	}

	if err := s.repo.UpdateStatus(id, newStatus); err != nil {
		return err
	}

	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", generateTimestamp()),
		UserID:     userID,
		UserName:   userName,
		Action:     "deploy",
		Resource:   "release",
		ResourceID: id,
		Details:    fmt.Sprintf("Deployed release %s to %s", rel.Version, environment),
	})

	return nil
}

func (s *ReleaseService) Rollback(id, userID, userName string) error {
	rel, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(id, "rolled_back"); err != nil {
		return err
	}

	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", generateTimestamp()),
		UserID:     userID,
		UserName:   userName,
		Action:     "rollback",
		Resource:   "release",
		ResourceID: id,
		Details:    fmt.Sprintf("Rolled back release %s", rel.Version),
	})

	return nil
}

func (s *ReleaseService) Approve(id, userID, userName string) error {
	if err := s.repo.IncrementApproval(id); err != nil {
		return err
	}

	s.auditRepo.Create(&model.AuditEntry{
		ID:         fmt.Sprintf("audit-%d", generateTimestamp()),
		UserID:     userID,
		UserName:   userName,
		Action:     "approve",
		Resource:   "release",
		ResourceID: id,
		Details:    "Approved release",
	})

	return nil
}
