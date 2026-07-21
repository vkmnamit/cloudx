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
	Clone(repoURL string, destDir string) (string, error)

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
func (s *gitService) Clone(repoURL string, destDir string) (string, error) {
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
	cmd := exec.Command("git", "clone", "--depth=1", repoURL, target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git clone failed: %w", err)
	}

	return target, nil
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
