package logs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository persists and retrieves deployment logs.
type Repository interface {
	Append(ctx context.Context, log *DeploymentLog) error
	ListByDeployment(ctx context.Context, deploymentID string) ([]*DeploymentLog, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Append(ctx context.Context, log *DeploymentLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO deployment_logs (id, deployment_id, phase, level, message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		log.ID, log.DeploymentID, string(log.Phase), string(log.Level), log.Message, log.CreatedAt,
	)
	return err
}

func (r *repository) ListByDeployment(ctx context.Context, deploymentID string) ([]*DeploymentLog, error) {
	query := `
		SELECT id, deployment_id, phase, level, message, created_at
		FROM deployment_logs
		WHERE deployment_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, deploymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*DeploymentLog
	for rows.Next() {
		l := &DeploymentLog{}
		var phase, level string
		err := rows.Scan(&l.ID, &l.DeploymentID, &phase, &level, &l.Message, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		l.Phase = LogPhase(phase)
		l.Level = LogLevel(level)
		logs = append(logs, l)
	}
	return logs, nil
}
