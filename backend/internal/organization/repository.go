package organization

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, org *Organization) error
	GetByID(ctx context.Context, id string) (*Organization, error)
	ListByOwner(ctx context.Context, ownerID string) ([]*Organization, error)
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, org *Organization) error {
	if org.ID == "" {
		org.ID = uuid.New().String()
	}
	now := time.Now()
	org.CreatedAt = now
	org.UpdatedAt = now

	if org.Plan == "" {
		org.Plan = "FREE"
	}

	query := `
		INSERT INTO organizations (id, owner_id, name, slug, plan, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::organization_plan, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, org.ID, org.OwnerID, org.Name, org.Slug, org.Plan, org.CreatedAt, org.UpdatedAt)
	return err
}

func (r *repository) GetByID(ctx context.Context, id string) (*Organization, error) {
	query := `
		SELECT id::text, owner_id::text, name, slug, plan::text, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`
	org := &Organization{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&org.ID,
		&org.OwnerID,
		&org.Name,
		&org.Slug,
		&org.Plan,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return org, nil
}

func (r *repository) ListByOwner(ctx context.Context, ownerID string) ([]*Organization, error) {
	query := `
		SELECT id::text, owner_id::text, name, slug, plan::text, created_at, updated_at
		FROM organizations
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*Organization
	for rows.Next() {
		org := &Organization{}
		if err := rows.Scan(
			&org.ID,
			&org.OwnerID,
			&org.Name,
			&org.Slug,
			&org.Plan,
			&org.CreatedAt,
			&org.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}

func (r *repository) Update(ctx context.Context, org *Organization) error {
	org.UpdatedAt = time.Now()
	query := `
		UPDATE organizations
		SET name = COALESCE(NULLIF($1, ''), name),
		    slug = COALESCE(NULLIF($2, ''), slug),
		    plan = CASE WHEN $3 = '' THEN plan ELSE $3::organization_plan END,
		    updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, org.Name, org.Slug, org.Plan, org.UpdatedAt, org.ID)
	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM organizations WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
