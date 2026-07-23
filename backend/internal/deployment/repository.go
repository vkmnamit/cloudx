package deployment

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
)

// Repository defines the full persistence interface for deployments.
type Repository interface {
	Create(ctx context.Context, dep *Deployment) error
	GetByID(ctx context.Context, id string) (*Deployment, error)
	ListByDeployedBy(ctx context.Context, deployedBy string) ([]*Deployment, error)
	UpdateMetadata(ctx context.Context, id string, commitSHA, commitMessage string) error
	UpdateStatus(ctx context.Context, id string, status string) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *pgx.Conn
}

func NewRepository(db *pgx.Conn) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, dep *Deployment) error {
	query := `
		INSERT INTO deployments
			(id, service_id, commit_sha, commit_message, deployed_by, status, image_tag, started_at, finished_at, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		dep.ID, dep.ServiceID, dep.CommitSHA, dep.CommitMessage, dep.DeployedBy,
		dep.Status, dep.ImageTag, dep.StartedAt, dep.FinishedAt, dep.CreatedAt, dep.UpdatedAt,
	)
	return err
}

func (r *repository) GetByID(ctx context.Context, id string) (*Deployment, error) {
	query := `
		SELECT id::text, service_id::text, COALESCE(commit_sha, ''), COALESCE(commit_message, ''), COALESCE(image_tag, ''),
			status, started_at, finished_at, deployed_by::text, created_at, updated_at
		FROM deployments WHERE id = $1
	`
	dep := &Deployment{}
	var finishedAt sql.NullTime
	err := r.db.QueryRow(ctx, query, id).Scan(
		&dep.ID, &dep.ServiceID, &dep.CommitSHA, &dep.CommitMessage, &dep.ImageTag,
		&dep.Status, &dep.StartedAt, &finishedAt, &dep.DeployedBy, &dep.CreatedAt, &dep.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if finishedAt.Valid {
		dep.FinishedAt = &finishedAt.Time
	}
	return dep, nil
}

func (r *repository) ListByDeployedBy(ctx context.Context, deployedBy string) ([]*Deployment, error) {
	query := `
		SELECT id::text, service_id::text, COALESCE(commit_sha, ''), COALESCE(commit_message, ''), COALESCE(image_tag, ''),
			status, started_at, finished_at, deployed_by::text, created_at, updated_at
		FROM deployments WHERE deployed_by = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, deployedBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*Deployment
	for rows.Next() {
		dep := &Deployment{}
		var finishedAt sql.NullTime
		err := rows.Scan(
			&dep.ID, &dep.ServiceID, &dep.CommitSHA, &dep.CommitMessage, &dep.ImageTag,
			&dep.Status, &dep.StartedAt, &finishedAt, &dep.DeployedBy, &dep.CreatedAt, &dep.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if finishedAt.Valid {
			dep.FinishedAt = &finishedAt.Time
		}
		deployments = append(deployments, dep)
	}
	return deployments, nil
}

func (r *repository) UpdateMetadata(ctx context.Context, id string, commitSHA, commitMessage string) error {
	query := `
		UPDATE deployments
		SET commit_sha = $1, commit_message = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, commitSHA, commitMessage, id)
	return err
}

func (r *repository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE deployments
		SET status = $1,
			finished_at = CASE
				WHEN $1 IN ('RUNNING', 'FAILED', 'CANCELLED') THEN COALESCE(finished_at, NOW())
				ELSE finished_at
			END,
			updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM deployments WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
