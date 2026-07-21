package deployment

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Repository defines the full persistence interface for deployments.
type Repository interface {
	Create(ctx context.Context, dep *Deployment) error
	GetByID(ctx context.Context, id string) (*Deployment, error)
	ListByUser(ctx context.Context, userID string) ([]*Deployment, error)
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
			(id, user_id, project_name, repository, framework, language, image_name, container_id, host_port, container_port, status, url, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.Exec(ctx, query,
		dep.ID, dep.UserID, dep.ProjectName, dep.Repository, dep.Framework,
		dep.Language, dep.ImageName, dep.ContainerID, dep.HostPort, dep.ContainerPort,
		dep.Status, dep.URL, dep.CreatedAt, dep.UpdatedAt,
	)
	return err
}

func (r *repository) GetByID(ctx context.Context, id string) (*Deployment, error) {
	query := `
		SELECT id, user_id, project_name, repository, framework, language, image_name, container_id, host_port, container_port, status, url, created_at, updated_at
		FROM deployments WHERE id = $1
	`
	dep := &Deployment{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&dep.ID, &dep.UserID, &dep.ProjectName, &dep.Repository, &dep.Framework,
		&dep.Language, &dep.ImageName, &dep.ContainerID, &dep.HostPort, &dep.ContainerPort,
		&dep.Status, &dep.URL, &dep.CreatedAt, &dep.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return dep, nil
}

func (r *repository) ListByUser(ctx context.Context, userID string) ([]*Deployment, error) {
	query := `
		SELECT id, user_id, project_name, repository, framework, language, image_name, container_id, host_port, container_port, status, url, created_at, updated_at
		FROM deployments WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*Deployment
	for rows.Next() {
		dep := &Deployment{}
		err := rows.Scan(
			&dep.ID, &dep.UserID, &dep.ProjectName, &dep.Repository, &dep.Framework,
			&dep.Language, &dep.ImageName, &dep.ContainerID, &dep.HostPort, &dep.ContainerPort,
			&dep.Status, &dep.URL, &dep.CreatedAt, &dep.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		deployments = append(deployments, dep)
	}
	return deployments, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE deployments SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM deployments WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
