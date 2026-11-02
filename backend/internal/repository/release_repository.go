package repository

import (
	"database/sql"

	"github.com/company/release-portal/internal/model"
)

type ReleaseRepository struct {
	db *sql.DB
}

func NewReleaseRepository(db *sql.DB) *ReleaseRepository {
	return &ReleaseRepository{db: db}
}

func (r *ReleaseRepository) Create(rel *model.Release) error {
	query := `INSERT INTO releases (id, service_id, version, status, description, branch, git_commit, created_by, approvals, required_approvals, is_primary, traffic_percentage, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())`
	_, err := r.db.Exec(query, rel.ID, rel.ServiceID, rel.Version, rel.Status, rel.Description, rel.Branch, rel.GitCommit, rel.CreatedBy, rel.Approvals, rel.RequiredApprovals, rel.IsPrimary, rel.TrafficPercentage)
	return err
}

func (r *ReleaseRepository) GetByID(id string) (*model.Release, error) {
	rel := &model.Release{}
	query := `SELECT id, service_id, version, status, description, branch, git_commit, created_by, approvals, required_approvals, is_primary, traffic_percentage, created_at, updated_at 
	          FROM releases WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&rel.ID, &rel.ServiceID, &rel.Version, &rel.Status, &rel.Description, &rel.Branch, &rel.GitCommit, &rel.CreatedBy, &rel.Approvals, &rel.RequiredApprovals, &rel.IsPrimary, &rel.TrafficPercentage, &rel.CreatedAt, &rel.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return rel, nil
}

func (r *ReleaseRepository) List(serviceID string) ([]model.Release, error) {
	var rows *sql.Rows
	var err error

	if serviceID != "" {
		query := `SELECT id, service_id, version, status, description, branch, git_commit, created_by, approvals, required_approvals, is_primary, traffic_percentage, created_at, updated_at 
		          FROM releases WHERE service_id = $1 ORDER BY created_at DESC`
		rows, err = r.db.Query(query, serviceID)
	} else {
		query := `SELECT id, service_id, version, status, description, branch, git_commit, created_by, approvals, required_approvals, is_primary, traffic_percentage, created_at, updated_at 
		          FROM releases ORDER BY created_at DESC`
		rows, err = r.db.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var releases []model.Release
	for rows.Next() {
		var rel model.Release
		if err := rows.Scan(&rel.ID, &rel.ServiceID, &rel.Version, &rel.Status, &rel.Description, &rel.Branch, &rel.GitCommit, &rel.CreatedBy, &rel.Approvals, &rel.RequiredApprovals, &rel.IsPrimary, &rel.TrafficPercentage, &rel.CreatedAt, &rel.UpdatedAt); err != nil {
			return nil, err
		}
		releases = append(releases, rel)
	}
	return releases, nil
}

func (r *ReleaseRepository) ListByServiceID(serviceID string) ([]model.Release, error) {
	return r.List(serviceID)
}

func (r *ReleaseRepository) Update(rel *model.Release) error {
	query := `UPDATE releases SET version=$2, status=$3, description=$4, branch=$5, git_commit=$6, approvals=$7, is_primary=$8, traffic_percentage=$9, updated_at=NOW() WHERE id=$1`
	_, err := r.db.Exec(query, rel.ID, rel.Version, rel.Status, rel.Description, rel.Branch, rel.GitCommit, rel.Approvals, rel.IsPrimary, rel.TrafficPercentage)
	return err
}

func (r *ReleaseRepository) UpdateStatus(id, status string) error {
	query := `UPDATE releases SET status=$2, updated_at=NOW() WHERE id=$1`
	_, err := r.db.Exec(query, id, status)
	return err
}

func (r *ReleaseRepository) IncrementApproval(id string) error {
	query := `UPDATE releases SET approvals = approvals + 1, updated_at=NOW() WHERE id=$1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ReleaseRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM releases WHERE id = $1`, id)
	return err
}
