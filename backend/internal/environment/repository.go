package environment

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Repository persists and retrieves environment variables.
type Repository interface {
	Set(ctx context.Context, vars []*EnvVar) error
	ListByDeployment(ctx context.Context, deploymentID string) ([]*EnvVar, error)
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *pgx.Conn
}

func NewRepository(db *pgx.Conn) Repository {
	return &repository{db: db}
}

// Set upserts a batch of environment variables for a deployment.
// Uses ON CONFLICT to update existing keys.
func (r *repository) Set(ctx context.Context, vars []*EnvVar) error {
	query := `
		INSERT INTO environment_variables (id, deployment_id, key, value, is_secret, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (deployment_id, key)
		DO UPDATE SET value = EXCLUDED.value, is_secret = EXCLUDED.is_secret, updated_at = NOW()
	`
	for _, v := range vars {
		if v.ID == "" {
			v.ID = uuid.New().String()
		}
		now := time.Now()
		if v.CreatedAt.IsZero() {
			v.CreatedAt = now
		}
		v.UpdatedAt = now

		_, err := r.db.Exec(ctx, query,
			v.ID, v.DeploymentID, v.Key, v.Value, v.IsSecret, v.CreatedAt, v.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) ListByDeployment(ctx context.Context, deploymentID string) ([]*EnvVar, error) {
	query := `
		SELECT id, deployment_id, key, value, is_secret, created_at, updated_at
		FROM environment_variables
		WHERE deployment_id = $1
		ORDER BY key ASC
	`
	rows, err := r.db.Query(ctx, query, deploymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vars []*EnvVar
	for rows.Next() {
		v := &EnvVar{}
		if err := rows.Scan(&v.ID, &v.DeploymentID, &v.Key, &v.Value, &v.IsSecret, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		vars = append(vars, v)
	}
	return vars, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM environment_variables WHERE id = $1`, id)
	return err
}
