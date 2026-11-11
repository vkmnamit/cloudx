package project

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id string) (*Project, error)
	ListByOrganization(ctx context.Context, orgID string) ([]*Project, error)
	ListAll(ctx context.Context) ([]*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, project *Project) error {
	if project.ID == "" {
		project.ID = uuid.New().String()
	}
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	if project.Visibility == "" {
		project.Visibility = "PRIVATE"
	}

	query := `
		INSERT INTO projects (id, organization_id, name, slug, description, visibility, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::project_visibility, $7, $8)
	`
	_, err := r.db.Exec(ctx, query, project.ID, project.OrganizationID, project.Name, project.Slug, project.Description, project.Visibility, project.CreatedAt, project.UpdatedAt)
	return err
}

func (r *repository) GetByID(ctx context.Context, id string) (*Project, error) {
	query := `
		SELECT id::text, organization_id::text, name, slug, description, visibility::text, created_at, updated_at
		FROM projects
		WHERE id = $1
	`
	p := &Project{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.OrganizationID,
		&p.Name,
		&p.Slug,
		&p.Description,
		&p.Visibility,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *repository) ListByOrganization(ctx context.Context, orgID string) ([]*Project, error) {
	query := `
		SELECT id::text, organization_id::text, name, slug, description, visibility::text, created_at, updated_at
		FROM projects
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		p := &Project{}
		if err := rows.Scan(
			&p.ID,
			&p.OrganizationID,
			&p.Name,
			&p.Slug,
			&p.Description,
			&p.Visibility,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *repository) ListAll(ctx context.Context) ([]*Project, error) {
	query := `
		SELECT id::text, organization_id::text, name, slug, description, visibility::text, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		p := &Project{}
		if err := rows.Scan(
			&p.ID,
			&p.OrganizationID,
			&p.Name,
			&p.Slug,
			&p.Description,
			&p.Visibility,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *repository) ListByOwner(ctx context.Context, ownerID string) ([]*Project, error) {
	query := `SELECT p.id::text,p.organization_id::text,p.name,p.slug,p.description,p.visibility::text,p.created_at,p.updated_at FROM projects p JOIN organizations o ON o.id=p.organization_id WHERE o.owner_id=$1 ORDER BY p.created_at DESC`
	rows, err := r.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := make([]*Project, 0)
	for rows.Next() {
		p := &Project{}
		if err := rows.Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Slug, &p.Description, &p.Visibility, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *repository) Update(ctx context.Context, project *Project) error {
	project.UpdatedAt = time.Now()
	query := `
		UPDATE projects
		SET name = COALESCE(NULLIF($1, ''), name),
		    slug = COALESCE(NULLIF($2, ''), slug),
		    description = $3,
		    visibility = CASE WHEN $4 = '' THEN visibility ELSE $4::project_visibility END,
		    updated_at = $5
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query, project.Name, project.Slug, project.Description, project.Visibility, project.UpdatedAt, project.ID)
	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM projects WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
