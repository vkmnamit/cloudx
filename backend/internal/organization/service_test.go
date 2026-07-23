package organization

import (
	"context"
	"errors"
	"testing"
)

type mockRepo struct {
	orgs map[string]*Organization
}

func (m *mockRepo) Create(ctx context.Context, org *Organization) error {
	org.ID = "test-org-id"
	m.orgs[org.ID] = org
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*Organization, error) {
	org, ok := m.orgs[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return org, nil
}

func (m *mockRepo) ListByOwner(ctx context.Context, ownerID string) ([]*Organization, error) {
	var list []*Organization
	for _, o := range m.orgs {
		if o.OwnerID == ownerID {
			list = append(list, o)
		}
	}
	return list, nil
}

func (m *mockRepo) Update(ctx context.Context, org *Organization) error {
	m.orgs[org.ID] = org
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	delete(m.orgs, id)
	return nil
}

func TestOrganizationService(t *testing.T) {
	repo := &mockRepo{orgs: make(map[string]*Organization)}
	svc := NewService(repo)

	ctx := context.Background()
	org, err := svc.Create(ctx, "user-1", CreateOrganizationRequest{
		Name: "Test Org",
		Slug: "test-org",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if org.Name != "Test Org" {
		t.Errorf("expected name 'Test Org', got '%s'", org.Name)
	}

	list, err := svc.ListByOwner(ctx, "user-1")
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 org in list, got %d (err: %v)", len(list), err)
	}
}
