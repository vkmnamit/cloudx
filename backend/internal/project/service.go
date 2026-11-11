package project

import (
	"context"
	"errors"
)

type Service interface {
	Create(ctx context.Context, req CreateProjectRequest) (*Project, error)
	GetByID(ctx context.Context, id string) (*Project, error)
	List(ctx context.Context, orgID string) ([]*Project, error)
	ListByOwner(ctx context.Context, ownerID string) ([]*Project, error)
	Update(ctx context.Context, id string, req UpdateProjectRequest) (*Project, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateProjectRequest) (*Project, error) {
	if req.OrganizationID == "" {
		return nil, errors.New("organization_id is required")
	}

	p := &Project{
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		Slug:           req.Slug,
		Description:    req.Description,
		Visibility:     req.Visibility,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*Project, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, orgID string) ([]*Project, error) {
	if orgID != "" {
		return s.repo.ListByOrganization(ctx, orgID)
	}
	return s.repo.ListAll(ctx)
}

func (s *service) ListByOwner(ctx context.Context, ownerID string) ([]*Project, error) {
	ownerLister, ok := s.repo.(interface {
		ListByOwner(context.Context, string) ([]*Project, error)
	})
	if !ok {
		return nil, errors.New("owner-scoped project listing is unavailable")
	}
	return ownerLister.ListByOwner(ctx, ownerID)
}

func (s *service) Update(ctx context.Context, id string, req UpdateProjectRequest) (*Project, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Slug != "" {
		p.Slug = req.Slug
	}
	if req.Description != nil {
		p.Description = req.Description
	}
	if req.Visibility != "" {
		p.Visibility = req.Visibility
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
