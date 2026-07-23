package projectenv

import (
	"context"
	"errors"
)

type Service interface {
	Create(ctx context.Context, req CreateEnvironmentRequest) (*Environment, error)
	GetByID(ctx context.Context, id string) (*Environment, error)
	List(ctx context.Context, projectID string) ([]*Environment, error)
	Update(ctx context.Context, id string, req UpdateEnvironmentRequest) (*Environment, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateEnvironmentRequest) (*Environment, error) {
	if req.ProjectID == "" {
		return nil, errors.New("project_id is required")
	}

	env := &Environment{
		ProjectID: req.ProjectID,
		Name:      req.Name,
	}

	if err := s.repo.Create(ctx, env); err != nil {
		return nil, err
	}
	return env, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*Environment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, projectID string) ([]*Environment, error) {
	if projectID != "" {
		return s.repo.ListByProject(ctx, projectID)
	}
	return s.repo.ListAll(ctx)
}

func (s *service) Update(ctx context.Context, id string, req UpdateEnvironmentRequest) (*Environment, error) {
	env, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		env.Name = req.Name
	}

	if err := s.repo.Update(ctx, env); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
