package environment

import (
	"context"
	"fmt"
	"strings"
)

// EnvironmentService manages environment variables for deployments.
type EnvironmentService interface {
	// SetVars persists a batch of env vars for a deployment.
	SetVars(ctx context.Context, deploymentID string, inputs []EnvVarInput) error
	ReplaceVars(ctx context.Context, serviceID string, inputs []EnvVarInput) error

	// GetVars returns all env vars for a deployment.
	// Secret values are included (the handler sanitises them for clients).
	GetVars(ctx context.Context, deploymentID string) ([]*EnvVar, error)

	// ToDockerArgs converts a deployment's env vars into the slice of
	// "-e KEY=VALUE" arguments required by `docker run`.
	ToDockerArgs(ctx context.Context, deploymentID string) ([]string, error)
}

type envService struct {
	repo   Repository
	cipher *secretCipher
}

func NewEnvironmentService(repo Repository, applicationSecrets ...string) EnvironmentService {
	cipher, _ := newSecretCipher(applicationSecrets...)
	return &envService{repo: repo, cipher: cipher}
}

func (s *envService) SetVars(ctx context.Context, deploymentID string, inputs []EnvVarInput) error {
	return s.saveVars(ctx, deploymentID, inputs, false)
}

func (s *envService) ReplaceVars(ctx context.Context, serviceID string, inputs []EnvVarInput) error {
	return s.saveVars(ctx, serviceID, inputs, true)
}

func (s *envService) saveVars(ctx context.Context, serviceID string, inputs []EnvVarInput, replace bool) error {
	vars := make([]*EnvVar, 0, len(inputs))
	existing := make(map[string]*EnvVar)
	if replace {
		stored, err := s.repo.ListByService(ctx, serviceID)
		if err != nil {
			return fmt.Errorf("load current environment variables: %w", err)
		}
		for _, variable := range stored {
			existing[variable.Key] = variable
		}
	}
	for _, in := range inputs {
		if in.Key == "" || strings.TrimSpace(in.Key) != in.Key || !validEnvKey(in.Key) {
			return fmt.Errorf("invalid environment variable key %q", in.Key)
		}
		value := in.Value
		if in.IsSecret {
			if value == "" {
				if current, ok := existing[in.Key]; ok && current.IsSecret {
					value = current.Value
					if !strings.HasPrefix(value, encryptedValuePrefix) && s.cipher != nil {
						var err error
						value, err = s.cipher.encrypt(value)
						if err != nil {
							return err
						}
					}
					vars = append(vars, &EnvVar{ServiceID: serviceID, Key: in.Key, Value: value, IsSecret: true})
					continue
				}
				return fmt.Errorf("environment variable %s needs a value", in.Key)
			}
			if s.cipher == nil {
				return fmt.Errorf("secret environment variables require JWT_SECRET to be configured")
			}
			var err error
			value, err = s.cipher.encrypt(value)
			if err != nil {
				return err
			}
		}
		if value == "" {
			return fmt.Errorf("environment variable %s has an empty value", in.Key)
		}
		vars = append(vars, &EnvVar{
			ServiceID: serviceID,
			Key:       in.Key,
			Value:     value,
			IsSecret:  in.IsSecret,
		})
	}
	if replace {
		return s.repo.Replace(ctx, serviceID, vars)
	}
	return s.repo.Set(ctx, vars)
}

func validEnvKey(key string) bool {
	for index, char := range key {
		valid := char == '_' || char >= 'A' && char <= 'Z' || index > 0 && (char >= 'a' && char <= 'z' || char >= '0' && char <= '9')
		if !valid {
			return false
		}
	}
	return len(key) > 0
}

func (s *envService) GetVars(ctx context.Context, deploymentID string) ([]*EnvVar, error) {
	vars, err := s.repo.ListByService(ctx, deploymentID)
	if err != nil {
		return nil, err
	}
	for _, value := range vars {
		if value.IsSecret {
			if s.cipher == nil && strings.HasPrefix(value.Value, encryptedValuePrefix) {
				return nil, fmt.Errorf("environment encryption is unavailable")
			}
			decrypted, err := s.cipher.decrypt(value.Value)
			if err != nil {
				return nil, fmt.Errorf("decrypt environment variable %s: %w", value.Key, err)
			}
			value.Value = decrypted
		}
	}
	return vars, nil
}

// ToDockerArgs fetches all env vars for the deployment and returns them
// in the format docker expects: ["-e", "KEY=VALUE", "-e", "KEY2=VALUE2"]
func (s *envService) ToDockerArgs(ctx context.Context, deploymentID string) ([]string, error) {
	vars, err := s.GetVars(ctx, deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load env vars: %w", err)
	}

	args := make([]string, 0, len(vars)*2)
	for _, v := range vars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", v.Key, v.Value))
	}
	return args, nil
}
