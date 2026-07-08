package readers

import (
	"fmt"
	"os"
	"strings"

	"github.com/ViitoJooj/pierrot/internal/compiler"
)

// RuntimeUsage define quais partes do runtime Pierrot a página realmente precisa.
type RuntimeUsage struct {
	Get      bool
	Client   bool
	Time     bool
	Reactive bool
	Markdown bool
	Render   bool
}

// Component é um arquivo .pierrot parseado
type Component struct {
	Styles   []string
	Scripts  []string
	Imports  []Import
	Meta     map[string]string
	Props    []string
	Script   string
	Template string

	// Novo: controla injeção seletiva do runtime
	Runtime RuntimeUsage
}

// Import é um componente importado
type Import struct {
	Name string
	Path string
}

// ParsePierrot lê um arquivo .pierrot
func ParsePierrot(path string) (*Component, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("lendo %s: %w", path, err)
	}

	return ParseSource(string(data)), nil
}

func ParseSource(src string) *Component {
	script, template := compiler.SplitScript(src)
	parts := compiler.ParseScript(script)

	c := &Component{
		Styles:   parts.Styles,
		Scripts:  parts.Scripts,
		Meta:     parts.Meta,
		Props:    parts.Props,
		Script:   parts.Script,
		Template: strings.TrimSpace(template),
	}

	for _, imp := range parts.Imports {
		c.Imports = append(c.Imports, Import{
			Name: imp.Name,
			Path: imp.Path,
		})
	}

	c.Runtime = detectRuntime(c)

	return c
}

func detectRuntime(c *Component) RuntimeUsage {
	var r RuntimeUsage

	source := c.Script + "\n" + c.Template

	if strings.Contains(source, "get.") {
		r.Get = true
	}

	if strings.Contains(source, "client.") {
		r.Client = true
	}

	if strings.Contains(source, "time.") {
		r.Time = true
	}

	if strings.Contains(source, "@render") {
		r.Render = true
	}

	if strings.Contains(source, "markdown") {
		r.Markdown = true
	}

	if strings.Contains(source, "${") ||
		strings.Contains(source, "@if") ||
		strings.Contains(source, "@for") {
		r.Reactive = true
	}

	return r
}
