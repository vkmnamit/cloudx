package organization

import (
	"context"
	"errors"
)

type Service interface {
	Create(ctx context.Context, ownerID string, req CreateOrganizationRequest) (*Organization, error)
	GetByID(ctx context.Context, id string) (*Organization, error)
	ListByOwner(ctx context.Context, ownerID string) ([]*Organization, error)
	Update(ctx context.Context, id string, req UpdateOrganizationRequest) (*Organization, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, ownerID string, req CreateOrganizationRequest) (*Organization, error) {
	if ownerID == "" {
		return nil, errors.New("owner_id is required")
	}

	org := &Organization{
		OwnerID: ownerID,
		Name:    req.Name,
		Slug:    req.Slug,
		Plan:    req.Plan,
	}

	if err := s.repo.Create(ctx, org); err != nil {
		return nil, err
	}
	return org, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*Organization, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListByOwner(ctx context.Context, ownerID string) ([]*Organization, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}

func (s *service) Update(ctx context.Context, id string, req UpdateOrganizationRequest) (*Organization, error) {
	org, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		org.Name = req.Name
	}
	if req.Slug != "" {
		org.Slug = req.Slug
	}
	if req.Plan != "" {
		org.Plan = req.Plan
	}

	if err := s.repo.Update(ctx, org); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
