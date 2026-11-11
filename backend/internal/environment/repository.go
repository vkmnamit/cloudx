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
	Replace(ctx context.Context, serviceID string, vars []*EnvVar) error
	ListByService(ctx context.Context, serviceID string) ([]*EnvVar, error)
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
		INSERT INTO environment_variables (
id,
service_id,
variable_key,
variable_value,
is_secret,
created_at,
updated_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (service_id, variable_key)
DO UPDATE
SET
variable_value = EXCLUDED.variable_value,
is_secret = EXCLUDED.is_secret,
updated_at = NOW();
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
			v.ID, v.ServiceID, v.Key, v.Value, v.IsSecret, v.CreatedAt, v.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) Replace(ctx context.Context, serviceID string, vars []*EnvVar) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM environment_variables WHERE service_id = $1`, serviceID); err != nil {
		return err
	}
	for _, variable := range vars {
		if variable.ID == "" {
			variable.ID = uuid.New().String()
		}
		now := time.Now()
		if variable.CreatedAt.IsZero() {
			variable.CreatedAt = now
		}
		variable.UpdatedAt = now
		_, err := tx.Exec(ctx, `INSERT INTO environment_variables (id, service_id, variable_key, variable_value, is_secret, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, variable.ID, serviceID, variable.Key, variable.Value, variable.IsSecret, variable.CreatedAt, variable.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *repository) ListByService(ctx context.Context, serviceID string) ([]*EnvVar, error) {
	query := `
		SELECT
id,
service_id,
variable_key,
variable_value,
is_secret,
created_at,
updated_at
FROM environment_variables
WHERE service_id=$1
ORDER BY variable_key;
	`
	rows, err := r.db.Query(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vars []*EnvVar
	for rows.Next() {
		v := &EnvVar{}
		if err := rows.Scan(&v.ID, &v.ServiceID, &v.Key, &v.Value, &v.IsSecret, &v.CreatedAt, &v.UpdatedAt); err != nil {
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
