package repository

import (
	"database/sql"

	"github.com/company/release-portal/internal/model"
)

type ActiveVersionRepository struct {
	db *sql.DB
}

func NewActiveVersionRepository(db *sql.DB) *ActiveVersionRepository {
	return &ActiveVersionRepository{db: db}
}

func (r *ActiveVersionRepository) Create(version *model.ActiveVersion) error {
	query := `INSERT INTO active_versions (id, service_id, release_id, version, branch, git_commit, environment, traffic_percentage, status, deployed_at, deployed_by)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.db.Exec(query, version.ID, version.ServiceID, version.ReleaseID, version.Version, version.Branch, version.GitCommit, version.Environment, version.TrafficPercentage, version.Status, version.DeployedAt, version.DeployedBy)
	return err
}

func (r *ActiveVersionRepository) GetByServiceID(serviceID string, environment string) ([]model.ActiveVersion, error) {
	query := `SELECT id, service_id, release_id, version, branch, git_commit, environment, traffic_percentage, status, deployed_at, deployed_by 
	          FROM active_versions WHERE service_id = $1 AND environment = $2 ORDER BY deployed_at DESC`
	rows, err := r.db.Query(query, serviceID, environment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []model.ActiveVersion
	for rows.Next() {
		var v model.ActiveVersion
		if err := rows.Scan(&v.ID, &v.ServiceID, &v.ReleaseID, &v.Version, &v.Branch, &v.GitCommit, &v.Environment, &v.TrafficPercentage, &v.Status, &v.DeployedAt, &v.DeployedBy); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}

func (r *ActiveVersionRepository) GetByID(id string) (*model.ActiveVersion, error) {
	query := `SELECT id, service_id, release_id, version, branch, git_commit, environment, traffic_percentage, status, deployed_at, deployed_by 
	          FROM active_versions WHERE id = $1`
	v := &model.ActiveVersion{}
	err := r.db.QueryRow(query, id).Scan(&v.ID, &v.ServiceID, &v.ReleaseID, &v.Version, &v.Branch, &v.GitCommit, &v.Environment, &v.TrafficPercentage, &v.Status, &v.DeployedAt, &v.DeployedBy)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *ActiveVersionRepository) GetActiveByServiceID(serviceID string, environment string) (*model.ActiveVersion, error) {
	v := &model.ActiveVersion{}
	query := `SELECT id, service_id, release_id, version, branch, git_commit, environment, traffic_percentage, status, deployed_at, deployed_by 
	          FROM active_versions WHERE service_id = $1 AND environment = $2 AND status = 'active' ORDER BY deployed_at DESC LIMIT 1`
	err := r.db.QueryRow(query, serviceID, environment).Scan(&v.ID, &v.ServiceID, &v.ReleaseID, &v.Version, &v.Branch, &v.GitCommit, &v.Environment, &v.TrafficPercentage, &v.Status, &v.DeployedAt, &v.DeployedBy)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *ActiveVersionRepository) UpdateTrafficPercentage(id string, percentage int) error {
	query := `UPDATE active_versions SET traffic_percentage = $2 WHERE id = $1`
	_, err := r.db.Exec(query, id, percentage)
	return err
}

func (r *ActiveVersionRepository) UpdateStatus(id string, status string) error {
	query := `UPDATE active_versions SET status = $2 WHERE id = $1`
	_, err := r.db.Exec(query, id, status)
	return err
}

func (r *ActiveVersionRepository) Delete(id string) error {
	query := `DELETE FROM active_versions WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ActiveVersionRepository) DeactivateAllExcept(serviceID string, environment string, excludeID string) error {
	query := `UPDATE active_versions SET status = 'inactive' WHERE service_id = $1 AND environment = $2 AND id != $3`
	_, err := r.db.Exec(query, serviceID, environment, excludeID)
	return err
}
