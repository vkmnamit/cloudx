package environment

import (
	"context"
	"fmt"
)

// EnvironmentService manages environment variables for deployments.
type EnvironmentService interface {
	// SetVars persists a batch of env vars for a deployment.
	SetVars(ctx context.Context, deploymentID string, inputs []EnvVarInput) error

	// GetVars returns all env vars for a deployment.
	// Secret values are included (the handler sanitises them for clients).
	GetVars(ctx context.Context, deploymentID string) ([]*EnvVar, error)

	// ToDockerArgs converts a deployment's env vars into the slice of
	// "-e KEY=VALUE" arguments required by `docker run`.
	ToDockerArgs(ctx context.Context, deploymentID string) ([]string, error)
}

type envService struct {
	repo Repository
}

func NewEnvironmentService(repo Repository) EnvironmentService {
	return &envService{repo: repo}
}

func (s *envService) SetVars(ctx context.Context, deploymentID string, inputs []EnvVarInput) error {
	vars := make([]*EnvVar, 0, len(inputs))
	for _, in := range inputs {
		vars = append(vars, &EnvVar{
			DeploymentID: deploymentID,
			Key:          in.Key,
			Value:        in.Value,
			IsSecret:     in.IsSecret,
		})
	}
	return s.repo.Set(ctx, vars)
}

func (s *envService) GetVars(ctx context.Context, deploymentID string) ([]*EnvVar, error) {
	return s.repo.ListByDeployment(ctx, deploymentID)
}

// ToDockerArgs fetches all env vars for the deployment and returns them
// in the format docker expects: ["-e", "KEY=VALUE", "-e", "KEY2=VALUE2"]
func (s *envService) ToDockerArgs(ctx context.Context, deploymentID string) ([]string, error) {
	vars, err := s.repo.ListByDeployment(ctx, deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load env vars: %w", err)
	}

	args := make([]string, 0, len(vars)*2)
	for _, v := range vars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", v.Key, v.Value))
	}
	return args, nil
}
