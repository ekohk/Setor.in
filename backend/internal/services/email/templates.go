package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

// Loaded once at startup; ParseFS errors crash the process so we know early.
var templates *template.Template

func init() {
	t, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		panic(fmt.Sprintf("email: parse templates: %v", err))
	}
	templates = t
}

// render executes the named template against data. The base layout is included
// transitively because each event template ends with `{{template "base" .}}`.
func render(name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("execute template %q: %w", name, err)
	}
	return buf.String(), nil
}
