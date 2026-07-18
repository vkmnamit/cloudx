package auth

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
	DB *pgx.Conn
}

func NewRepository(db *pgx.Conn) *Repository {
	return &Repository{
		DB: db,
	}
}
func (r *Repository) FindByEmail(email string) (*User, error) {

	var user User

	err := r.DB.QueryRow(
		context.Background(),
		`
		SELECT
		    id,
		    name,
		    email,
		    password_hash,
		    avatar_url,
		    email_verified,
		    created_at,
		    updated_at
		FROM users
		WHERE email=$1
		`,
		email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) CreateUser(user *User) error {

	_, err := r.DB.Exec(
		context.Background(),
		`
		INSERT INTO users
		(name,email,password_hash)
		VALUES($1,$2,$3)
		`,
		user.Name,
		user.Email,
		user.PasswordHash,
	)

	return err
}
