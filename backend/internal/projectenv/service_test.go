package projectenv

import (
	"context"
	"errors"
	"testing"
)

type mockRepo struct {
	envs map[string]*Environment
}

func (m *mockRepo) Create(ctx context.Context, env *Environment) error {
	env.ID = "env-1"
	m.envs[env.ID] = env
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*Environment, error) {
	e, ok := m.envs[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return e, nil
}

func (m *mockRepo) ListByProject(ctx context.Context, projectID string) ([]*Environment, error) {
	var res []*Environment
	for _, e := range m.envs {
		if e.ProjectID == projectID {
			res = append(res, e)
		}
	}
	return res, nil
}

func (m *mockRepo) ListAll(ctx context.Context) ([]*Environment, error) {
	var res []*Environment
	for _, e := range m.envs {
		res = append(res, e)
	}
	return res, nil
}

func (m *mockRepo) Update(ctx context.Context, env *Environment) error {
	m.envs[env.ID] = env
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	delete(m.envs, id)
	return nil
}

func TestProjectEnvService(t *testing.T) {
	repo := &mockRepo{envs: make(map[string]*Environment)}
	svc := NewService(repo)
	ctx := context.Background()

	e, err := svc.Create(ctx, CreateEnvironmentRequest{
		ProjectID: "proj-1",
		Name:      "production",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if e.Name != "production" {
		t.Errorf("expected 'production', got '%s'", e.Name)
	}
}
