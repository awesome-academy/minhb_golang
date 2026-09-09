package web

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path"
	"strings"
)

//go:embed templates static
var Files embed.FS

const layoutFile = "templates/admin/layout.html"

type Templates map[string]*template.Template

func ParseTemplates() (Templates, error) {
	pages, err := fs.Glob(Files, "templates/admin/*.html")
	if err != nil {
		return nil, err
	}

	templates := Templates{}
	for _, page := range pages {
		if page == layoutFile {
			continue
		}
		parsed, err := template.ParseFS(Files, layoutFile, page)
		if err != nil {
			return nil, fmt.Errorf("parse template %s: %w", page, err)
		}
		name := "admin/" + strings.TrimSuffix(path.Base(page), ".html")
		templates[name] = parsed
	}
	return templates, nil
}

func (t Templates) ExecuteTemplate(w io.Writer, name string, data any) error {
	page, ok := t[name]
	if !ok {
		return fmt.Errorf("template %q not found", name)
	}
	return page.ExecuteTemplate(w, "layout", data)
}
