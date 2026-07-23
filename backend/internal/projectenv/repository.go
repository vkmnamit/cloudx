package projectenv

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository interface {
	Create(ctx context.Context, env *Environment) error
	GetByID(ctx context.Context, id string) (*Environment, error)
	ListByProject(ctx context.Context, projectID string) ([]*Environment, error)
	ListAll(ctx context.Context) ([]*Environment, error)
	Update(ctx context.Context, env *Environment) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *pgx.Conn
}

func NewRepository(db *pgx.Conn) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, env *Environment) error {
	if env.ID == "" {
		env.ID = uuid.New().String()
	}
	now := time.Now()
	env.CreatedAt = now
	env.UpdatedAt = now

	query := `
		INSERT INTO environments (id, project_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, env.ID, env.ProjectID, env.Name, env.CreatedAt, env.UpdatedAt)
	return err
}

func (r *repository) GetByID(ctx context.Context, id string) (*Environment, error) {
	query := `
		SELECT id::text, project_id::text, name, created_at, updated_at
		FROM environments
		WHERE id = $1
	`
	env := &Environment{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&env.ID,
		&env.ProjectID,
		&env.Name,
		&env.CreatedAt,
		&env.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func (r *repository) ListByProject(ctx context.Context, projectID string) ([]*Environment, error) {
	query := `
		SELECT id::text, project_id::text, name, created_at, updated_at
		FROM environments
		WHERE project_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envs []*Environment
	for rows.Next() {
		env := &Environment{}
		if err := rows.Scan(
			&env.ID,
			&env.ProjectID,
			&env.Name,
			&env.CreatedAt,
			&env.UpdatedAt,
		); err != nil {
			return nil, err
		}
		envs = append(envs, env)
	}
	return envs, nil
}

func (r *repository) ListAll(ctx context.Context) ([]*Environment, error) {
	query := `
		SELECT id::text, project_id::text, name, created_at, updated_at
		FROM environments
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envs []*Environment
	for rows.Next() {
		env := &Environment{}
		if err := rows.Scan(
			&env.ID,
			&env.ProjectID,
			&env.Name,
			&env.CreatedAt,
			&env.UpdatedAt,
		); err != nil {
			return nil, err
		}
		envs = append(envs, env)
	}
	return envs, nil
}

func (r *repository) Update(ctx context.Context, env *Environment) error {
	env.UpdatedAt = time.Now()
	query := `
		UPDATE environments
		SET name = COALESCE(NULLIF($1, ''), name),
		    updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, env.Name, env.UpdatedAt, env.ID)
	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM environments WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
