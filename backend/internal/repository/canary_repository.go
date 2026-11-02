package repository

import (
	"database/sql"
	"time"

	"github.com/company/release-portal/internal/model"
)

type CanaryRepository struct {
	db *sql.DB
}

func NewCanaryRepository(db *sql.DB) *CanaryRepository {
	return &CanaryRepository{db: db}
}

func (r *CanaryRepository) Create(canary *model.CanaryDeployment) error {
	query := `INSERT INTO canary_deployments (
		id, release_id, service_id, status, percentage, 
		stable_version, canary_version, stable_revision, canary_revision,
		argo_app, git_tag, started_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	
	_, err := r.db.Exec(query,
		canary.ID,
		canary.ReleaseID,
		canary.ServiceID,
		canary.Status,
		canary.Percentage,
		canary.StableVersion,
		canary.CanaryVersion,
		canary.StableRevision,
		canary.CanaryRevision,
		canary.ArgoApp,
		canary.GitTag,
		canary.StartedAt,
		canary.UpdatedAt,
	)
	return err
}

func (r *CanaryRepository) GetByID(id string) (*model.CanaryDeployment, error) {
	query := `SELECT 
		id, release_id, service_id, status, percentage,
		stable_version, canary_version, stable_revision, canary_revision,
		argo_app, git_tag, started_at, updated_at
	FROM canary_deployments WHERE id = $1`
	
	canary := &model.CanaryDeployment{}
	err := r.db.QueryRow(query, id).Scan(
		&canary.ID,
		&canary.ReleaseID,
		&canary.ServiceID,
		&canary.Status,
		&canary.Percentage,
		&canary.StableVersion,
		&canary.CanaryVersion,
		&canary.StableRevision,
		&canary.CanaryRevision,
		&canary.ArgoApp,
		&canary.GitTag,
		&canary.StartedAt,
		&canary.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return canary, nil
}

func (r *CanaryRepository) GetByReleaseID(releaseID string) (*model.CanaryDeployment, error) {
	query := `SELECT 
		id, release_id, service_id, status, percentage,
		stable_version, canary_version, stable_revision, canary_revision,
		argo_app, git_tag, started_at, updated_at
	FROM canary_deployments WHERE release_id = $1 ORDER BY started_at DESC LIMIT 1`
	
	canary := &model.CanaryDeployment{}
	err := r.db.QueryRow(query, releaseID).Scan(
		&canary.ID,
		&canary.ReleaseID,
		&canary.ServiceID,
		&canary.Status,
		&canary.Percentage,
		&canary.StableVersion,
		&canary.CanaryVersion,
		&canary.StableRevision,
		&canary.CanaryRevision,
		&canary.ArgoApp,
		&canary.GitTag,
		&canary.StartedAt,
		&canary.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return canary, nil
}

func (r *CanaryRepository) GetActiveByServiceID(serviceID string) (*model.CanaryDeployment, error) {
	query := `SELECT 
		id, release_id, service_id, status, percentage,
		stable_version, canary_version, stable_revision, canary_revision,
		argo_app, git_tag, started_at, updated_at
	FROM canary_deployments 
	WHERE service_id = $1 AND status IN ('pending', 'running', 'promoting')
	ORDER BY started_at DESC LIMIT 1`
	
	canary := &model.CanaryDeployment{}
	err := r.db.QueryRow(query, serviceID).Scan(
		&canary.ID,
		&canary.ReleaseID,
		&canary.ServiceID,
		&canary.Status,
		&canary.Percentage,
		&canary.StableVersion,
		&canary.CanaryVersion,
		&canary.StableRevision,
		&canary.CanaryRevision,
		&canary.ArgoApp,
		&canary.GitTag,
		&canary.StartedAt,
		&canary.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return canary, nil
}

func (r *CanaryRepository) Update(canary *model.CanaryDeployment) error {
	query := `UPDATE canary_deployments SET
		status = $2,
		percentage = $3,
		updated_at = $4
	WHERE id = $1`
	
	canary.UpdatedAt = time.Now()
	_, err := r.db.Exec(query,
		canary.ID,
		canary.Status,
		canary.Percentage,
		canary.UpdatedAt,
	)
	return err
}

func (r *CanaryRepository) UpdateStatus(id string, status string) error {
	query := `UPDATE canary_deployments SET
		status = $2,
		updated_at = $3
	WHERE id = $1`
	
	_, err := r.db.Exec(query, id, status, time.Now())
	return err
}

func (r *CanaryRepository) UpdatePercentage(id string, percentage int) error {
	query := `UPDATE canary_deployments SET
		percentage = $2,
		updated_at = $3
	WHERE id = $1`
	
	_, err := r.db.Exec(query, id, percentage, time.Now())
	return err
}

func (r *CanaryRepository) ListByServiceID(serviceID string) ([]model.CanaryDeployment, error) {
	query := `SELECT 
		id, release_id, service_id, status, percentage,
		stable_version, canary_version, stable_revision, canary_revision,
		argo_app, git_tag, started_at, updated_at
	FROM canary_deployments 
	WHERE service_id = $1
	ORDER BY started_at DESC`
	
	rows, err := r.db.Query(query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var canaries []model.CanaryDeployment
	for rows.Next() {
		canary := model.CanaryDeployment{}
		err := rows.Scan(
			&canary.ID,
			&canary.ReleaseID,
			&canary.ServiceID,
			&canary.Status,
			&canary.Percentage,
			&canary.StableVersion,
			&canary.CanaryVersion,
			&canary.StableRevision,
			&canary.CanaryRevision,
			&canary.ArgoApp,
			&canary.GitTag,
			&canary.StartedAt,
			&canary.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		canaries = append(canaries, canary)
	}
	return canaries, nil
}

func (r *CanaryRepository) Delete(id string) error {
	query := `DELETE FROM canary_deployments WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
