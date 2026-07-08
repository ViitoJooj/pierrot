package readers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Settings espelha o settings.pierrot.json
type Settings struct {
	App struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Entry   string `json:"entry"`
		Port    int    `json:"port"`
	} `json:"app"`
	Dotenv struct {
		Enabled bool   `json:"enabled"`
		Path    string `json:"path"`
	} `json:"dotenv"`
	Build struct {
		OutDir    string `json:"outDir"`
		Minify    *bool  `json:"minify"`
		Sourcemap bool   `json:"sourcemap"`
	} `json:"build"`
}

// Project é o settings resolvido: caminhos absolutos prontos para uso.
// Regra: app.entry é relativo à pasta do settings; todos os outros caminhos
// são relativos à pasta do entry (o src do projeto)
type Project struct {
	Root      string // pasta do settings.pierrot.json
	Src       string // pasta do entry (main.pierrot)
	Entry     string // caminho do main.pierrot
	Port      int
	OutDir    string
	Minify    bool
	Sourcemap bool
	Dotenv    string            // vazio = desabilitado
	Env       map[string]string // variáveis do .env (nil = dotenv desabilitado)
}

// LoadProject lê o settings.pierrot.json de root e resolve os caminhos.
// Sem settings, assume o layout padrão: src/main.pierrot (ou main.pierrot na
// raiz, layout antigo), porta 3000, build em dist/, minify ligado
func LoadProject(root string) (*Project, error) {
	var s Settings
	data, err := os.ReadFile(filepath.Join(root, "settings.pierrot.json"))
	if err == nil {
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("settings.pierrot.json: %w", err)
		}
	}

	entry := s.App.Entry
	if entry == "" {
		entry = "./src/main.pierrot"
		if _, err := os.Stat(filepath.Join(root, "src", "main.pierrot")); err != nil {
			entry = "./main.pierrot"
		}
	}

	p := &Project{
		Root:   root,
		Entry:  resolve(root, entry),
		Port:   3000,
		Minify: true,
	}
	p.Src = filepath.Dir(p.Entry)

	if s.App.Port != 0 {
		p.Port = s.App.Port
	}
	out := s.Build.OutDir
	if out == "" {
		out = "./dist"
	}
	p.OutDir = resolve(p.Src, out)
	if s.Build.Minify != nil {
		p.Minify = *s.Build.Minify
	}
	p.Sourcemap = s.Build.Sourcemap
	if s.Dotenv.Enabled && s.Dotenv.Path != "" {
		p.Dotenv = resolve(p.Src, s.Dotenv.Path)
	}
	return p, nil
}

func resolve(base, rel string) string {
	if filepath.IsAbs(rel) {
		return filepath.Clean(rel)
	}
	return filepath.Clean(filepath.Join(base, filepath.FromSlash(rel)))
}

// LoadDotenv injeta as linhas KEY=VALUE do arquivo no ambiente do processo e
// devolve só as variáveis do arquivo — é esse map que o get.Dotenv consulta,
// para não expor o ambiente inteiro do sistema ao front
func LoadDotenv(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	env := map[string]string{}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key := strings.TrimSpace(k)
		val := strings.Trim(strings.TrimSpace(v), `"`)
		env[key] = val
		os.Setenv(key, val)
	}
	return env, nil
}
