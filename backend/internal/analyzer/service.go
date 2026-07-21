package analyzer

import (
	"os"
	"path/filepath"
)

// Analysis holds the result of analyzing a project directory.
type Analysis struct {
	Language         string
	Framework        string
	PackageManager   string
	BuildCommand     string
	StartCommand     string
	ExposedPort      int
	DockerfileExists bool
	ComposeExists    bool
}

// AnalyzerService analyzes a project directory and returns an Analysis.
type AnalyzerService interface {
	Analyze(projectPath string) (*Analysis, error)
}

type analyzerService struct{}

func NewAnalyzerService() AnalyzerService {
	return &analyzerService{}
}

func (s *analyzerService) Analyze(projectPath string) (*Analysis, error) {
	result := &Analysis{
		Language:       "unknown",
		Framework:      "unknown",
		PackageManager: "",
		BuildCommand:   "",
		StartCommand:   "",
		ExposedPort:    8080,
	}

	// Check for docker-compose.yml
	if _, err := os.Stat(filepath.Join(projectPath, "docker-compose.yml")); err == nil {
		result.ComposeExists = true
	}

	// Check for existing Dockerfile
	if _, err := os.Stat(filepath.Join(projectPath, "Dockerfile")); err == nil {
		result.DockerfileExists = true
		return result, nil
	}

	// Node.js
	if _, err := os.Stat(filepath.Join(projectPath, "package.json")); err == nil {
		result.Language = "Node.js"
		result.ExposedPort = 3000

		// Detect package manager
		if _, err := os.Stat(filepath.Join(projectPath, "yarn.lock")); err == nil {
			result.PackageManager = "yarn"
			result.BuildCommand = "yarn install"
			result.StartCommand = "yarn start"
		} else if _, err := os.Stat(filepath.Join(projectPath, "pnpm-lock.yaml")); err == nil {
			result.PackageManager = "pnpm"
			result.BuildCommand = "pnpm install"
			result.StartCommand = "pnpm start"
		} else {
			result.PackageManager = "npm"
			result.BuildCommand = "npm install"
			result.StartCommand = "npm start"
		}
		return result, nil
	}

	// Go
	if _, err := os.Stat(filepath.Join(projectPath, "go.mod")); err == nil {
		result.Language = "Go"
		result.PackageManager = "go modules"
		result.BuildCommand = "go build -o main ."
		result.StartCommand = "./main"
		result.ExposedPort = 8080
		return result, nil
	}

	// Python
	if _, err := os.Stat(filepath.Join(projectPath, "requirements.txt")); err == nil {
		result.Language = "Python"
		result.PackageManager = "pip"
		result.BuildCommand = "pip install -r requirements.txt"
		result.StartCommand = "python app.py"
		result.ExposedPort = 5000
		return result, nil
	}

	// Rust
	if _, err := os.Stat(filepath.Join(projectPath, "Cargo.toml")); err == nil {
		result.Language = "Rust"
		result.PackageManager = "cargo"
		result.BuildCommand = "cargo build --release"
		result.StartCommand = "./target/release/app"
		result.ExposedPort = 8080
		return result, nil
	}

	return result, nil
}
