package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/company/release-portal/internal/model"
)

type IntegrationRepository struct {
	db *sql.DB
}

func NewIntegrationRepository(db *sql.DB) *IntegrationRepository {
	return &IntegrationRepository{db: db}
}

func (r *IntegrationRepository) Create(integration *model.Integration) error {
	configJSON, err := json.Marshal(integration.Config)
	if err != nil {
		return err
	}

	query := `INSERT INTO integrations (id, type, name, url, enabled, config, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`
	_, err = r.db.Exec(query, integration.ID, integration.Type, integration.Name, integration.URL, integration.Enabled, configJSON)
	return err
}

func (r *IntegrationRepository) GetByID(id string) (*model.Integration, error) {
	integration := &model.Integration{}
	var configJSON []byte

	query := `SELECT id, type, name, url, enabled, config, created_at, updated_at 
	          FROM integrations WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&integration.ID, &integration.Type, &integration.Name, &integration.URL,
		&integration.Enabled, &configJSON, &integration.CreatedAt, &integration.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configJSON, &integration.Config); err != nil {
		return nil, err
	}

	return integration, nil
}

func (r *IntegrationRepository) GetByType(integrationType string) ([]model.Integration, error) {
	query := `SELECT id, type, name, url, enabled, config, created_at, updated_at 
	          FROM integrations WHERE type = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, integrationType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []model.Integration
	for rows.Next() {
		var integration model.Integration
		var configJSON []byte

		if err := rows.Scan(
			&integration.ID, &integration.Type, &integration.Name, &integration.URL,
			&integration.Enabled, &configJSON, &integration.CreatedAt, &integration.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(configJSON, &integration.Config); err != nil {
			return nil, err
		}

		integrations = append(integrations, integration)
	}

	return integrations, nil
}

func (r *IntegrationRepository) GetEnabledByType(integrationType string) ([]model.Integration, error) {
	query := `SELECT id, type, name, url, enabled, config, created_at, updated_at 
	          FROM integrations WHERE type = $1 AND enabled = true ORDER BY created_at DESC`
	rows, err := r.db.Query(query, integrationType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []model.Integration
	for rows.Next() {
		var integration model.Integration
		var configJSON []byte

		if err := rows.Scan(
			&integration.ID, &integration.Type, &integration.Name, &integration.URL,
			&integration.Enabled, &configJSON, &integration.CreatedAt, &integration.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(configJSON, &integration.Config); err != nil {
			return nil, err
		}

		integrations = append(integrations, integration)
	}

	return integrations, nil
}

func (r *IntegrationRepository) List() ([]model.Integration, error) {
	query := `SELECT id, type, name, url, enabled, config, created_at, updated_at 
	          FROM integrations ORDER BY type, created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []model.Integration
	for rows.Next() {
		var integration model.Integration
		var configJSON []byte

		if err := rows.Scan(
			&integration.ID, &integration.Type, &integration.Name, &integration.URL,
			&integration.Enabled, &configJSON, &integration.CreatedAt, &integration.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(configJSON, &integration.Config); err != nil {
			return nil, err
		}

		integrations = append(integrations, integration)
	}

	return integrations, nil
}

func (r *IntegrationRepository) Update(integration *model.Integration) error {
	configJSON, err := json.Marshal(integration.Config)
	if err != nil {
		return err
	}

	query := `UPDATE integrations SET type=$2, name=$3, url=$4, enabled=$5, config=$6, updated_at=NOW() WHERE id=$1`
	_, err = r.db.Exec(query, integration.ID, integration.Type, integration.Name, integration.URL, integration.Enabled, configJSON)
	return err
}

func (r *IntegrationRepository) Delete(id string) error {
	query := `DELETE FROM integrations WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
