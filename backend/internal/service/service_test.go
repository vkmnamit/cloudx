package service

import (
	"context"
	"errors"
	"testing"
)

type mockRepo struct {
	records map[string]*ServiceRecord
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*ServiceRecord, error) {
	r, ok := m.records[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return r, nil
}

func (m *mockRepo) Create(ctx context.Context, s *ServiceRecord) error {
	s.ID = "svc-1"
	m.records[s.ID] = s
	return nil
}

func (m *mockRepo) ListByEnvironment(ctx context.Context, envID string) ([]*ServiceRecord, error) {
	var res []*ServiceRecord
	for _, r := range m.records {
		if r.EnvironmentID == envID {
			res = append(res, r)
		}
	}
	return res, nil
}

func (m *mockRepo) ListAll(ctx context.Context) ([]*ServiceRecord, error) {
	var res []*ServiceRecord
	for _, r := range m.records {
		res = append(res, r)
	}
	return res, nil
}

func (m *mockRepo) Update(ctx context.Context, s *ServiceRecord) error {
	m.records[s.ID] = s
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	delete(m.records, id)
	return nil
}

func TestServicePackage(t *testing.T) {
	repo := &mockRepo{records: make(map[string]*ServiceRecord)}
	svc := NewService(repo)
	ctx := context.Background()

	rec, err := svc.Create(ctx, CreateServiceRequest{
		EnvironmentID: "env-1",
		Name:          "api-server",
		Runtime:       "Go",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	fetched, err := svc.GetByID(ctx, rec.ID)
	if err != nil {
		t.Fatalf("expected GetByID to succeed, got %v", err)
	}

	if fetched.Name != "api-server" {
		t.Errorf("expected 'api-server', got '%s'", fetched.Name)
	}
}
