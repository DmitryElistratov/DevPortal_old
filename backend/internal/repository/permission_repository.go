package repository

import (
	"database/sql"

	"github.com/company/release-portal/internal/model"
)

type PermissionRepository struct {
	db *sql.DB
}

func NewPermissionRepository(db *sql.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) GetByUserAndService(userID, serviceID string) (*model.ProjectPermission, error) {
	perm := &model.ProjectPermission{}
	query := `SELECT id, user_id, service_id, role, granted_by, created_at 
	          FROM project_permissions WHERE user_id = $1 AND service_id = $2`
	err := r.db.QueryRow(query, userID, serviceID).Scan(
		&perm.ID, &perm.UserID, &perm.ServiceID, &perm.Role, &perm.GrantedBy, &perm.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return perm, nil
}

func (r *PermissionRepository) Upsert(perm *model.ProjectPermission) error {
	query := `INSERT INTO project_permissions (id, user_id, service_id, role, granted_by, created_at)
	          VALUES ($1, $2, $3, $4, $5, NOW())
	          ON CONFLICT (user_id, service_id) 
	          DO UPDATE SET role = $4, granted_by = $5`
	_, err := r.db.Exec(query, perm.ID, perm.UserID, perm.ServiceID, perm.Role, perm.GrantedBy)
	return err
}

func (r *PermissionRepository) Delete(userID, serviceID string) error {
	_, err := r.db.Exec(`DELETE FROM project_permissions WHERE user_id = $1 AND service_id = $2`, userID, serviceID)
	return err
}

func (r *PermissionRepository) ListByUser(userID string) ([]model.ProjectPermission, error) {
	query := `SELECT id, user_id, service_id, role, granted_by, created_at 
	          FROM project_permissions WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []model.ProjectPermission
	for rows.Next() {
		var p model.ProjectPermission
		if err := rows.Scan(&p.ID, &p.UserID, &p.ServiceID, &p.Role, &p.GrantedBy, &p.CreatedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}
	return permissions, nil
}

func (r *PermissionRepository) ListByService(serviceID string) ([]model.ProjectPermission, error) {
	query := `SELECT id, user_id, service_id, role, granted_by, created_at 
	          FROM project_permissions WHERE service_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []model.ProjectPermission
	for rows.Next() {
		var p model.ProjectPermission
		if err := rows.Scan(&p.ID, &p.UserID, &p.ServiceID, &p.Role, &p.GrantedBy, &p.CreatedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}
	return permissions, nil
}
