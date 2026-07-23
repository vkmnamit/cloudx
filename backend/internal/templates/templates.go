package templates

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/vkmnamit/cloudx/internal/analyzer"
)

//go:embed *.tmpl
var fs embed.FS

// TemplateData holds the data passed to the Dockerfile templates.
type TemplateData struct {
	Language       string
	Framework      string
	PackageManager string
	BuildCommand   string
	StartCommand   string
	ExposedPort    int
	ModuleName     string
}

// RenderDockerfile generates a Dockerfile string from the embedded templates based on the language.
func RenderDockerfile(analysis *analyzer.Analysis) (string, error) {
	if analysis == nil {
		return "", fmt.Errorf("analysis cannot be nil")
	}

	data := TemplateData{
		Language:       analysis.Language,
		Framework:      analysis.Framework,
		PackageManager: analysis.PackageManager,
		BuildCommand:   analysis.BuildCommand,
		StartCommand:   analysis.StartCommand,
		ExposedPort:    analysis.ExposedPort,
		ModuleName:     analysis.ModuleName,
	}

	var tmplName string
	switch analysis.Language {
	case "Node.js":
		tmplName = "node.tmpl"
	case "Go":
		tmplName = "go.tmpl"
	case "Python":
		tmplName = "python.tmpl"
	case "Rust":
		tmplName = "rust.tmpl"
	default:
		return "", fmt.Errorf("no template found for language: %s", analysis.Language)
	}

	tmpl, err := template.ParseFS(fs, tmplName)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", tmplName, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", tmplName, err)
	}

	return buf.String(), nil
}
