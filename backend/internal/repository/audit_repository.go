package repository

import (
	"database/sql"

	"github.com/company/release-portal/internal/model"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(entry *model.AuditEntry) error {
	query := `INSERT INTO audit_log (id, user_id, user_name, action, resource, resource_id, details, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`
	_, err := r.db.Exec(query, entry.ID, entry.UserID, entry.UserName, entry.Action, entry.Resource, entry.ResourceID, entry.Details)
	return err
}

func (r *AuditRepository) List(limit, offset int) ([]model.AuditEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `SELECT id, user_id, user_name, action, resource, resource_id, details, created_at 
	          FROM audit_log ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.AuditEntry
	for rows.Next() {
		var e model.AuditEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.UserName, &e.Action, &e.Resource, &e.ResourceID, &e.Details, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (r *AuditRepository) ListByResource(resource, resourceID string) ([]model.AuditEntry, error) {
	query := `SELECT id, user_id, user_name, action, resource, resource_id, details, created_at 
	          FROM audit_log WHERE resource = $1 AND resource_id = $2 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, resource, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.AuditEntry
	for rows.Next() {
		var e model.AuditEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.UserName, &e.Action, &e.Resource, &e.ResourceID, &e.Details, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
