package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository persists and retrieves service records.
type Repository interface {
	GetByID(ctx context.Context, id string) (*ServiceRecord, error)
	Create(ctx context.Context, s *ServiceRecord) error
	ListByEnvironment(ctx context.Context, envID string) ([]*ServiceRecord, error)
	ListAll(ctx context.Context) ([]*ServiceRecord, error)
	Update(ctx context.Context, s *ServiceRecord) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id string) (*ServiceRecord, error) {
	query := `
		SELECT
			id::text,
			environment_id::text,
			name,
			runtime,
			COALESCE(repository_url, ''),
			COALESCE(branch, 'main'),
			COALESCE(dockerfile_path, './Dockerfile'),
			COALESCE(build_context, '.'),
			COALESCE(port, 0),
			COALESCE(health_check, ''),
			COALESCE(cpu_limit, ''),
			COALESCE(memory_limit, ''),
			COALESCE(replicas, 1),
			COALESCE(auto_deploy, TRUE),
			created_at,
			updated_at
		FROM services
		WHERE id = $1
	`

	serviceRecord := &ServiceRecord{}
	if err := r.db.QueryRow(ctx, query, id).Scan(
		&serviceRecord.ID,
		&serviceRecord.EnvironmentID,
		&serviceRecord.Name,
		&serviceRecord.Runtime,
		&serviceRecord.RepositoryURL,
		&serviceRecord.Branch,
		&serviceRecord.DockerfilePath,
		&serviceRecord.BuildContext,
		&serviceRecord.Port,
		&serviceRecord.HealthCheck,
		&serviceRecord.CPULimit,
		&serviceRecord.MemoryLimit,
		&serviceRecord.Replicas,
		&serviceRecord.AutoDeploy,
		&serviceRecord.CreatedAt,
		&serviceRecord.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return serviceRecord, nil
}

func (r *repository) Create(ctx context.Context, s *ServiceRecord) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now

	if s.Branch == "" {
		s.Branch = "main"
	}
	if s.DockerfilePath == "" {
		s.DockerfilePath = "./Dockerfile"
	}
	if s.BuildContext == "" {
		s.BuildContext = "."
	}
	if s.Replicas <= 0 {
		s.Replicas = 1
	}

	query := `
		INSERT INTO services (
			id, environment_id, name, runtime, repository_url, branch, dockerfile_path, build_context,
			port, health_check, cpu_limit, memory_limit, replicas, auto_deploy, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`
	_, err := r.db.Exec(ctx, query,
		s.ID, s.EnvironmentID, s.Name, s.Runtime, s.RepositoryURL, s.Branch, s.DockerfilePath, s.BuildContext,
		s.Port, s.HealthCheck, s.CPULimit, s.MemoryLimit, s.Replicas, s.AutoDeploy, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *repository) ListByEnvironment(ctx context.Context, envID string) ([]*ServiceRecord, error) {
	query := `
		SELECT
			id::text, environment_id::text, name, runtime, COALESCE(repository_url, ''),
			COALESCE(branch, 'main'), COALESCE(dockerfile_path, './Dockerfile'), COALESCE(build_context, '.'),
			COALESCE(port, 0), COALESCE(health_check, ''), COALESCE(cpu_limit, ''), COALESCE(memory_limit, ''),
			COALESCE(replicas, 1), COALESCE(auto_deploy, TRUE), created_at, updated_at
		FROM services
		WHERE environment_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, envID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*ServiceRecord
	for rows.Next() {
		s := &ServiceRecord{}
		if err := rows.Scan(
			&s.ID, &s.EnvironmentID, &s.Name, &s.Runtime, &s.RepositoryURL,
			&s.Branch, &s.DockerfilePath, &s.BuildContext, &s.Port, &s.HealthCheck,
			&s.CPULimit, &s.MemoryLimit, &s.Replicas, &s.AutoDeploy, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, s)
	}
	return records, nil
}

func (r *repository) ListAll(ctx context.Context) ([]*ServiceRecord, error) {
	query := `
		SELECT
			id::text, environment_id::text, name, runtime, COALESCE(repository_url, ''),
			COALESCE(branch, 'main'), COALESCE(dockerfile_path, './Dockerfile'), COALESCE(build_context, '.'),
			COALESCE(port, 0), COALESCE(health_check, ''), COALESCE(cpu_limit, ''), COALESCE(memory_limit, ''),
			COALESCE(replicas, 1), COALESCE(auto_deploy, TRUE), created_at, updated_at
		FROM services
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*ServiceRecord
	for rows.Next() {
		s := &ServiceRecord{}
		if err := rows.Scan(
			&s.ID, &s.EnvironmentID, &s.Name, &s.Runtime, &s.RepositoryURL,
			&s.Branch, &s.DockerfilePath, &s.BuildContext, &s.Port, &s.HealthCheck,
			&s.CPULimit, &s.MemoryLimit, &s.Replicas, &s.AutoDeploy, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, s)
	}
	return records, nil
}

func (r *repository) ListByOwner(ctx context.Context, ownerID string) ([]*ServiceRecord, error) {
	query := `SELECT s.id::text,s.environment_id::text,s.name,s.runtime,COALESCE(s.repository_url,''),COALESCE(s.branch,'main'),COALESCE(s.dockerfile_path,'./Dockerfile'),COALESCE(s.build_context,'.'),COALESCE(s.port,0),COALESCE(s.health_check,''),COALESCE(s.cpu_limit,''),COALESCE(s.memory_limit,''),COALESCE(s.replicas,1),COALESCE(s.auto_deploy,TRUE),s.created_at,s.updated_at FROM services s JOIN environments e ON e.id=s.environment_id JOIN projects p ON p.id=e.project_id JOIN organizations o ON o.id=p.organization_id WHERE o.owner_id=$1 ORDER BY s.created_at DESC`
	rows, err := r.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]*ServiceRecord, 0)
	for rows.Next() {
		s := &ServiceRecord{}
		if err := rows.Scan(&s.ID, &s.EnvironmentID, &s.Name, &s.Runtime, &s.RepositoryURL, &s.Branch, &s.DockerfilePath, &s.BuildContext, &s.Port, &s.HealthCheck, &s.CPULimit, &s.MemoryLimit, &s.Replicas, &s.AutoDeploy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, s)
	}
	return records, rows.Err()
}

func (r *repository) Update(ctx context.Context, s *ServiceRecord) error {
	s.UpdatedAt = time.Now()
	query := `
		UPDATE services
		SET name = COALESCE(NULLIF($1, ''), name),
		    runtime = COALESCE(NULLIF($2, ''), runtime),
		    repository_url = $3,
		    branch = COALESCE(NULLIF($4, ''), branch),
		    dockerfile_path = COALESCE(NULLIF($5, ''), dockerfile_path),
		    build_context = COALESCE(NULLIF($6, ''), build_context),
		    port = $7,
		    health_check = $8,
		    cpu_limit = $9,
		    memory_limit = $10,
		    replicas = $11,
		    auto_deploy = $12,
		    updated_at = $13
		WHERE id = $14
	`
	_, err := r.db.Exec(ctx, query,
		s.Name, s.Runtime, s.RepositoryURL, s.Branch, s.DockerfilePath, s.BuildContext,
		s.Port, s.HealthCheck, s.CPULimit, s.MemoryLimit, s.Replicas, s.AutoDeploy, s.UpdatedAt, s.ID,
	)
	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM services WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
