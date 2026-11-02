package repository

import (
	"database/sql"

	"github.com/company/release-portal/internal/model"
)

type ServiceRepository struct {
	db *sql.DB
}

func NewServiceRepository(db *sql.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) Create(svc *model.Service) error {
	query := `INSERT INTO services (id, name, team, type, lifecycle, description, language, repository, owner, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())`
	_, err := r.db.Exec(query, svc.ID, svc.Name, svc.Team, svc.Type, svc.Lifecycle, svc.Description, svc.Language, svc.Repository, svc.Owner)
	return err
}

func (r *ServiceRepository) GetByID(id string) (*model.Service, error) {
	svc := &model.Service{}
	query := `SELECT id, name, team, type, lifecycle, description, language, repository, owner, created_at, updated_at 
	          FROM services WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&svc.ID, &svc.Name, &svc.Team, &svc.Type, &svc.Lifecycle, &svc.Description, &svc.Language, &svc.Repository, &svc.Owner, &svc.CreatedAt, &svc.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (r *ServiceRepository) GetByName(name string) (*model.Service, error) {
	svc := &model.Service{}
	query := `SELECT id, name, team, type, lifecycle, description, language, repository, owner, created_at, updated_at 
	          FROM services WHERE name = $1`
	err := r.db.QueryRow(query, name).Scan(&svc.ID, &svc.Name, &svc.Team, &svc.Type, &svc.Lifecycle, &svc.Description, &svc.Language, &svc.Repository, &svc.Owner, &svc.CreatedAt, &svc.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (r *ServiceRepository) GetByRepository(repoURL string) (*model.Service, error) {
	svc := &model.Service{}
	query := `SELECT id, name, team, type, lifecycle, description, language, repository, owner, created_at, updated_at 
	          FROM services WHERE repository = $1`
	err := r.db.QueryRow(query, repoURL).Scan(&svc.ID, &svc.Name, &svc.Team, &svc.Type, &svc.Lifecycle, &svc.Description, &svc.Language, &svc.Repository, &svc.Owner, &svc.CreatedAt, &svc.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (r *ServiceRepository) List() ([]model.Service, error) {
	query := `SELECT id, name, team, type, lifecycle, description, language, repository, owner, created_at, updated_at 
	          FROM services ORDER BY name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []model.Service
	for rows.Next() {
		var s model.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.Team, &s.Type, &s.Lifecycle, &s.Description, &s.Language, &s.Repository, &s.Owner, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, nil
}

func (r *ServiceRepository) Update(svc *model.Service) error {
	query := `UPDATE services SET name=$2, team=$3, type=$4, lifecycle=$5, description=$6, language=$7, repository=$8, owner=$9, updated_at=NOW() WHERE id=$1`
	_, err := r.db.Exec(query, svc.ID, svc.Name, svc.Team, svc.Type, svc.Lifecycle, svc.Description, svc.Language, svc.Repository, svc.Owner)
	return err
}

func (r *ServiceRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM services WHERE id = $1`, id)
	return err
}
