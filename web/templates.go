package web

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path"
	"strings"
)

//go:embed templates static
var Files embed.FS

const (
	templatesRoot = "templates"
	adminRoot     = "templates/admin"
	layoutFile    = "templates/admin/layout.html"
)

type Templates map[string]*template.Template

func ParseTemplates() (Templates, error) {
	templates := Templates{}
	err := fs.WalkDir(Files, adminRoot, func(file string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || file == layoutFile || path.Ext(file) != ".html" {
			return nil
		}
		parsed, err := template.ParseFS(Files, layoutFile, file)
		if err != nil {
			return fmt.Errorf("parse template %s: %w", file, err)
		}
		name := strings.TrimSuffix(strings.TrimPrefix(file, templatesRoot+"/"), ".html")
		templates[name] = parsed.Option("missingkey=error")
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(templates) == 0 {
		return nil, errors.New("no admin templates found under " + adminRoot)
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
