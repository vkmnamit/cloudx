package logs

import (
	"bufio"
	"context"
	"fmt"
	"io"
)

// LogService writes and retrieves deployment logs.
// It also provides a Writer that can be attached to os/exec commands
// so every line of build output is persisted in real time.
type LogService interface {
	// Append persists a single log line.
	Append(ctx context.Context, deploymentID string, phase LogPhase, level LogLevel, message string) error

	// GetLogs returns all logs for a deployment, ordered ascending.
	GetLogs(ctx context.Context, deploymentID string) ([]*DeploymentLog, error)

	// NewWriter returns an io.Writer that streams each newline-delimited
	// line from an exec.Command into deployment_logs.
	NewWriter(ctx context.Context, deploymentID string, phase LogPhase, level LogLevel) io.Writer
}

type logService struct {
	repo Repository
}

func NewLogService(repo Repository) LogService {
	return &logService{repo: repo}
}

func (s *logService) Append(ctx context.Context, deploymentID string, phase LogPhase, level LogLevel, message string) error {
	return s.repo.Append(ctx, &DeploymentLog{
		DeploymentID: deploymentID,
		Phase:        phase,
		Level:        level,
		Message:      message,
	})
}

func (s *logService) GetLogs(ctx context.Context, deploymentID string) ([]*DeploymentLog, error) {
	return s.repo.ListByDeployment(ctx, deploymentID)
}

// NewWriter returns an io.Writer backed by a goroutine that reads lines
// and persists each one via Append. The writer is safe to pass directly
// to exec.Cmd.Stdout or exec.Cmd.Stderr.
func (s *logService) NewWriter(ctx context.Context, deploymentID string, phase LogPhase, level LogLevel) io.Writer {
	pr, pw := io.Pipe()

	go func() {
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if err := s.Append(ctx, deploymentID, phase, level, line); err != nil {
				// Best-effort: log to stderr and continue so the build isn't blocked
				fmt.Printf("[cloudx/logs] failed to persist log: %v\n", err)
			}
		}
	}()

	return pw
}
