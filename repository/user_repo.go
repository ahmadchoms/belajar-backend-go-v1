package repository

import (
	"database/sql"
	"phase3-api-architecture/models"
)

type UserRepository struct {
	DB *sql.DB
}

// Testing Interface
type UserRepoInterface interface {
	Register(u models.User) error
	GetByEmail(email string) (models.User, error)
}

func (r *UserRepository) Register(u models.User) error {
	query := "INSERT INTO users (email, password) VALUES ($1, $2)"
	_, err := r.DB.Exec(query, u.Email, u.Password)
	return err
}

func (r *UserRepository) GetByEmail(email string) (models.User, error) {
	var u models.User
	query := "SELECT id, email, password, role FROM users WHERE email = $1"

	err := r.DB.QueryRow(query, email).Scan(&u.ID, &u.Email, &u.Password, &u.Role)
	return u, err
}
