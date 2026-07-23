package analyzer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	ModuleName       string
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

	if _, err := os.Stat(filepath.Join(projectPath, "docker-compose.yml")); err == nil {
		result.ComposeExists = true
	}

	// 1. Dockerfile has priority
	if _, err := os.Stat(filepath.Join(projectPath, "Dockerfile")); err == nil {
		result.DockerfileExists = true
		// Minimal parsing just to extract a port if possible (Optional but good)
		extractPortFromDockerfile(filepath.Join(projectPath, "Dockerfile"), result)
		return result, nil
	}

	// 2. Node.js
	packageJsonPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(packageJsonPath); err == nil {
		result.Language = "Node.js"
		result.ExposedPort = 3000 // default node port

		if _, err := os.Stat(filepath.Join(projectPath, "yarn.lock")); err == nil {
			result.PackageManager = "yarn"
		} else if _, err := os.Stat(filepath.Join(projectPath, "pnpm-lock.yaml")); err == nil {
			result.PackageManager = "pnpm"
		} else {
			result.PackageManager = "npm"
		}

		result.BuildCommand = result.PackageManager + " install"
		result.StartCommand = result.PackageManager + " start"

		if data, err := os.ReadFile(packageJsonPath); err == nil {
			var pkg struct {
				Dependencies    map[string]string `json:"dependencies"`
				DevDependencies map[string]string `json:"devDependencies"`
				Scripts         struct {
					Start string `json:"start"`
					Build string `json:"build"`
				} `json:"scripts"`
			}
			if err := json.Unmarshal(data, &pkg); err == nil {
				// Detect Framework
				if hasDep(pkg.Dependencies, "next") {
					result.Framework = "Next.js"
					result.ExposedPort = 3000
				} else if hasDep(pkg.Dependencies, "@nestjs/core") {
					result.Framework = "NestJS"
					result.ExposedPort = 3000
				} else if hasDep(pkg.Dependencies, "express") {
					result.Framework = "Express"
					result.ExposedPort = 3000
				} else if hasDep(pkg.Dependencies, "react") && hasDep(pkg.Dependencies, "react-scripts") {
					result.Framework = "React"
					result.ExposedPort = 3000
				} else if hasDep(pkg.DevDependencies, "vite") {
					result.Framework = "Vite"
					result.ExposedPort = 5173 // Default vite port, though typically dev only. Assuming build output is served.
				}

				if pkg.Scripts.Build != "" {
					result.BuildCommand = result.PackageManager + " run build"
				}
				if pkg.Scripts.Start != "" {
					result.StartCommand = result.PackageManager + " run start"
					if strings.Contains(pkg.Scripts.Start, "PORT=") {
						// Simple heuristic to extract PORT=xxxx
						// (A more complex regex might be needed for production)
					}
				}
			}
		}
		return result, nil
	}

	// 3. Go
	goModPath := filepath.Join(projectPath, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		result.Language = "Go"
		result.PackageManager = "go modules"
		result.ExposedPort = 8080

		result.BuildCommand = "go build -o app ."
		result.StartCommand = "./app"

		if data, err := os.ReadFile(goModPath); err == nil {
			content := string(data)
			scanner := bufio.NewScanner(strings.NewReader(content))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "module ") {
					result.ModuleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
				}
			}
			
			if strings.Contains(content, "github.com/gin-gonic/gin") {
				result.Framework = "Gin"
				result.ExposedPort = 8080
			} else if strings.Contains(content, "github.com/gofiber/fiber") {
				result.Framework = "Fiber"
				result.ExposedPort = 3000
			}
		}

		if _, err := os.Stat(filepath.Join(projectPath, "cmd", "server")); err == nil {
			result.BuildCommand = "go build -o app ./cmd/server"
		}

		return result, nil
	}

	// 4. Python
	reqTxtPath := filepath.Join(projectPath, "requirements.txt")
	if _, err := os.Stat(reqTxtPath); err == nil {
		result.Language = "Python"
		result.PackageManager = "pip"
		result.BuildCommand = "pip install --no-cache-dir -r requirements.txt"
		result.StartCommand = "python app.py"
		result.ExposedPort = 5000 // Default flask

		if data, err := os.ReadFile(reqTxtPath); err == nil {
			content := strings.ToLower(string(data))
			if strings.Contains(content, "django") {
				result.Framework = "Django"
				result.StartCommand = "python manage.py runserver 0.0.0.0:8000"
				result.ExposedPort = 8000
			} else if strings.Contains(content, "flask") {
				result.Framework = "Flask"
				result.ExposedPort = 5000
			} else if strings.Contains(content, "fastapi") {
				result.Framework = "FastAPI"
				result.ExposedPort = 8000
				result.StartCommand = "uvicorn main:app --host 0.0.0.0 --port 8000"
			}
		}

		if _, err := os.Stat(filepath.Join(projectPath, "main.py")); err == nil {
			if result.Framework != "Django" && result.Framework != "FastAPI" {
				result.StartCommand = "python main.py"
			}
		}
		return result, nil
	}

	// 5. Rust
	cargoPath := filepath.Join(projectPath, "Cargo.toml")
	if _, err := os.Stat(cargoPath); err == nil {
		result.Language = "Rust"
		result.PackageManager = "cargo"
		result.BuildCommand = "cargo build --release"
		result.StartCommand = "./target/release/app"
		result.ExposedPort = 8080
		
		if data, err := os.ReadFile(cargoPath); err == nil {
			content := string(data)
			if strings.Contains(content, "actix-web") {
				result.Framework = "Actix"
			} else if strings.Contains(content, "rocket") {
				result.Framework = "Rocket"
				result.ExposedPort = 8000
			} else if strings.Contains(content, "axum") {
				result.Framework = "Axum"
			}
		}
		return result, nil
	}

	return result, nil
}

func hasDep(deps map[string]string, pkg string) bool {
	if deps == nil {
		return false
	}
	_, exists := deps[pkg]
	return exists
}

func extractPortFromDockerfile(path string, result *Analysis) {
	if data, err := os.ReadFile(path); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(strings.ToUpper(line), "EXPOSE ") {
				// E.g., EXPOSE 8080
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					var port int
					// Simplistic parser, ignores multiple ports or udp
					if _, err := fmt.Sscanf(parts[1], "%d", &port); err == nil {
						result.ExposedPort = port
						break
					}
				}
			}
		}
	}
}
