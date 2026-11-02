package repository

import (
	"database/sql"

	"github.com/company/release-portal/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	query := `INSERT INTO users (id, username, email, full_name, password_hash, role, auth_provider, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`
	_, err := r.db.Exec(query, user.ID, user.Username, user.Email, user.FullName, user.Password, user.Role, user.AuthProvider)
	return err
}

func (r *UserRepository) GetByID(id string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, email, full_name, password_hash, role, COALESCE(auth_provider, 'local'), COALESCE(external_id, ''), created_at, updated_at FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email, &user.FullName, &user.Password, &user.Role, &user.AuthProvider, &user.ExternalID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, email, full_name, password_hash, role, COALESCE(auth_provider, 'local'), COALESCE(external_id, ''), created_at, updated_at FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.FullName, &user.Password, &user.Role, &user.AuthProvider, &user.ExternalID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByExternalID(provider, externalID string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, email, full_name, password_hash, role, COALESCE(auth_provider, 'local'), COALESCE(external_id, ''), created_at, updated_at 
	          FROM users WHERE auth_provider = $1 AND external_id = $2`
	err := r.db.QueryRow(query, provider, externalID).Scan(&user.ID, &user.Username, &user.Email, &user.FullName, &user.Password, &user.Role, &user.AuthProvider, &user.ExternalID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Update(user *model.User) error {
	query := `UPDATE users SET username=$2, email=$3, full_name=$4, role=$5, auth_provider=$6, external_id=$7, updated_at=NOW() WHERE id=$1`
	_, err := r.db.Exec(query, user.ID, user.Username, user.Email, user.FullName, user.Role, user.AuthProvider, user.ExternalID)
	return err
}

func (r *UserRepository) List() ([]model.User, error) {
	query := `SELECT id, username, email, full_name, password_hash, role, COALESCE(auth_provider, 'local'), COALESCE(external_id, ''), created_at, updated_at FROM users ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.Password, &u.Role, &u.AuthProvider, &u.ExternalID, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	return err
}
