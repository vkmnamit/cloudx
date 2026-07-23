package project

import (
	"context"
	"errors"
	"testing"
)

type mockRepo struct {
	projects map[string]*Project
}

func (m *mockRepo) Create(ctx context.Context, p *Project) error {
	p.ID = "proj-1"
	m.projects[p.ID] = p
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*Project, error) {
	p, ok := m.projects[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return p, nil
}

func (m *mockRepo) ListByOrganization(ctx context.Context, orgID string) ([]*Project, error) {
	var res []*Project
	for _, p := range m.projects {
		if p.OrganizationID == orgID {
			res = append(res, p)
		}
	}
	return res, nil
}

func (m *mockRepo) ListAll(ctx context.Context) ([]*Project, error) {
	var res []*Project
	for _, p := range m.projects {
		res = append(res, p)
	}
	return res, nil
}

func (m *mockRepo) Update(ctx context.Context, p *Project) error {
	m.projects[p.ID] = p
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	delete(m.projects, id)
	return nil
}

func TestProjectService(t *testing.T) {
	repo := &mockRepo{projects: make(map[string]*Project)}
	svc := NewService(repo)
	ctx := context.Background()

	p, err := svc.Create(ctx, CreateProjectRequest{
		OrganizationID: "org-1",
		Name:           "CloudX Web",
		Slug:           "cloudx-web",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if p.Name != "CloudX Web" {
		t.Errorf("expected 'CloudX Web', got '%s'", p.Name)
	}
}
