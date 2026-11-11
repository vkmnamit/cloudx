package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitService handles cloning and pulling repositories.
// GitHub OAuth and webhooks are intentionally deferred to a later milestone.
type GitService interface {
	// Clone clones a remote repository into a local directory.
	// Returns the absolute path to the cloned project.
	Clone(repoURL string, branch string, destDir string) (string, error)
	CloneWithToken(repoURL string, branch string, destDir, token string) (string, error)

	// CommitMetadata returns the current commit SHA and message from a local checkout.
	CommitMetadata(projectPath string) (sha string, message string, err error)

	// Pull pulls the latest changes for a repo that was already cloned.
	Pull(projectPath string) error
}

type gitService struct {
	// baseDir is the root directory where repos are cloned.
	// e.g. /tmp/cloudx/repos
	baseDir string
}

func NewGitService(baseDir string) GitService {
	return &gitService{baseDir: baseDir}
}

// Clone performs a shallow `git clone` of repoURL into baseDir/<slug>.
// If the directory already exists, it pulls instead of cloning.
func (s *gitService) Clone(repoURL string, branch string, destDir string) (string, error) {
	return s.CloneWithToken(repoURL, branch, destDir, "")
}

func (s *gitService) CloneWithToken(repoURL string, branch string, destDir, token string) (string, error) {
	if err := os.MkdirAll(s.baseDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create base dir: %w", err)
	}

	// Build a filesystem-safe name from the repo URL
	slug := repoSlug(repoURL)
	target := filepath.Join(s.baseDir, destDir, slug)

	// If already cloned, pull instead
	if _, err := os.Stat(filepath.Join(target, ".git")); err == nil {
		if pullErr := s.Pull(target); pullErr != nil {
			return target, fmt.Errorf("pull failed: %w", pullErr)
		}
		return target, nil
	}

	// Shallow clone (depth=1) for speed
	args := []string{"clone", "--depth=1"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, repoURL, target)
	cmd := exec.Command("git", args...)
	if token != "" {
		askpassFile, err := os.CreateTemp(s.baseDir, ".cloudx-git-askpass-*")
		if err != nil {
			return "", fmt.Errorf("prepare temporary git credentials: %w", err)
		}
		askpass := askpassFile.Name()
		if _, err := askpassFile.Write([]byte("#!/bin/sh\ncase \"$1\" in *Username*) echo x-access-token;; *Password*) printf '%s' \"$CLOUDX_GIT_TOKEN\";; esac\n")); err != nil {
			askpassFile.Close()
			os.Remove(askpass)
			return "", fmt.Errorf("write temporary git credentials: %w", err)
		}
		if err := askpassFile.Chmod(0o700); err != nil {
			askpassFile.Close()
			os.Remove(askpass)
			return "", fmt.Errorf("protect temporary git credentials: %w", err)
		}
		if err := askpassFile.Close(); err != nil {
			os.Remove(askpass)
			return "", fmt.Errorf("close temporary git credentials: %w", err)
		}
		defer os.Remove(askpass)
		cmd.Env = append(os.Environ(), "GIT_ASKPASS="+askpass, "GIT_TERMINAL_PROMPT=0", "CLOUDX_GIT_TOKEN="+token)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git clone failed: %w", err)
	}

	return target, nil
}

// CommitMetadata returns the HEAD SHA and commit message for a cloned repository.
func (s *gitService) CommitMetadata(projectPath string) (string, string, error) {
	shaCmd := exec.Command("git", "rev-parse", "HEAD")
	shaCmd.Dir = projectPath
	shaOut, err := shaCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("git rev-parse failed: %w", err)
	}

	msgCmd := exec.Command("git", "log", "-1", "--pretty=%B")
	msgCmd.Dir = projectPath
	msgOut, err := msgCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("git log failed: %w", err)
	}

	return strings.TrimSpace(string(shaOut)), strings.TrimSpace(string(msgOut)), nil
}

// Pull runs `git pull` in the given project directory.
func (s *gitService) Pull(projectPath string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// repoSlug converts a repo URL to a safe directory name.
// "https://github.com/user/my-app.git" → "user-my-app"
func repoSlug(repoURL string) string {
	slug := repoURL
	slug = strings.TrimSuffix(slug, ".git")
	slug = strings.TrimPrefix(slug, "https://")
	slug = strings.TrimPrefix(slug, "http://")
	slug = strings.TrimPrefix(slug, "git@")

	// Replace colons, slashes, dots with dashes
	replacer := strings.NewReplacer(
		"github.com/", "",
		"gitlab.com/", "",
		"/", "-",
		":", "-",
		".", "-",
	)
	return replacer.Replace(slug)
}
