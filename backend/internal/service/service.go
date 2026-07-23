package service

import (
	"context"
	"errors"
)

// Service exposes service record lookups and CRUD for deployment orchestration and management.
type Service interface {
	GetByID(ctx context.Context, id string) (*ServiceRecord, error)
	Create(ctx context.Context, req CreateServiceRequest) (*ServiceRecord, error)
	List(ctx context.Context, envID string) ([]*ServiceRecord, error)
	Update(ctx context.Context, id string, req UpdateServiceRequest) (*ServiceRecord, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(ctx context.Context, id string) (*ServiceRecord, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) Create(ctx context.Context, req CreateServiceRequest) (*ServiceRecord, error) {
	if req.EnvironmentID == "" {
		return nil, errors.New("environment_id is required")
	}
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.Runtime == "" {
		return nil, errors.New("runtime is required")
	}

	autoDeploy := true
	if req.AutoDeploy != nil {
		autoDeploy = *req.AutoDeploy
	}

	branch := req.Branch
	if branch == "" {
		branch = "main"
	}

	dockerfilePath := req.DockerfilePath
	if dockerfilePath == "" {
		dockerfilePath = "./Dockerfile"
	}

	buildContext := req.BuildContext
	if buildContext == "" {
		buildContext = "."
	}

	replicas := req.Replicas
	if replicas <= 0 {
		replicas = 1
	}

	record := &ServiceRecord{
		EnvironmentID:  req.EnvironmentID,
		Name:           req.Name,
		Runtime:        req.Runtime,
		RepositoryURL:  req.RepositoryURL,
		Branch:         branch,
		DockerfilePath: dockerfilePath,
		BuildContext:   buildContext,
		Port:           req.Port,
		HealthCheck:    req.HealthCheck,
		CPULimit:       req.CPULimit,
		MemoryLimit:    req.MemoryLimit,
		Replicas:       replicas,
		AutoDeploy:     autoDeploy,
	}

	if err := s.repo.Create(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *service) List(ctx context.Context, envID string) ([]*ServiceRecord, error) {
	if envID != "" {
		return s.repo.ListByEnvironment(ctx, envID)
	}
	return s.repo.ListAll(ctx)
}

func (s *service) Update(ctx context.Context, id string, req UpdateServiceRequest) (*ServiceRecord, error) {
	record, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		record.Name = req.Name
	}
	if req.Runtime != "" {
		record.Runtime = req.Runtime
	}
	if req.RepositoryURL != "" {
		record.RepositoryURL = req.RepositoryURL
	}
	if req.Branch != "" {
		record.Branch = req.Branch
	}
	if req.DockerfilePath != "" {
		record.DockerfilePath = req.DockerfilePath
	}
	if req.BuildContext != "" {
		record.BuildContext = req.BuildContext
	}
	if req.Port > 0 {
		record.Port = req.Port
	}
	if req.HealthCheck != "" {
		record.HealthCheck = req.HealthCheck
	}
	if req.CPULimit != "" {
		record.CPULimit = req.CPULimit
	}
	if req.MemoryLimit != "" {
		record.MemoryLimit = req.MemoryLimit
	}
	if req.Replicas > 0 {
		record.Replicas = req.Replicas
	}
	if req.AutoDeploy != nil {
		record.AutoDeploy = *req.AutoDeploy
	}

	if err := s.repo.Update(ctx, record); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
