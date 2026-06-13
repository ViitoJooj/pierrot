package workers

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ViitoJooj/pierrot/internal/readers"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/spf13/cobra"
)

// Build renderiza todas as páginas de pages/ como HTML estático no outDir do
// settings, copiando assets. A página de set.Default também vira o index.html
// da raiz
func Build(cmd *cobra.Command, args []string) {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	p, err := readers.LoadProject(root)
	if err != nil {
		log.Fatal(err)
	}
	if p.Dotenv != "" {
		if p.Env, err = readers.LoadDotenv(p.Dotenv); err != nil {
			log.Printf("dotenv: %v", err)
		}
	}
	pagesDir := filepath.Join(p.Src, "pages")
	dist := p.OutDir

	if err := os.RemoveAll(dist); err != nil {
		log.Fatal(err)
	}

	var pages []string
	filepath.WalkDir(pagesDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "index.pierrot" {
			rel, err := filepath.Rel(pagesDir, filepath.Dir(p))
			if err != nil {
				return err
			}
			pages = append(pages, filepath.ToSlash(rel))
		}
		return nil
	})
	if len(pages) == 0 {
		log.Fatalf("nenhuma página encontrada em %s", pagesDir)
	}

	failed := false
	for _, page := range pages {
		comp, err := readers.ParsePierrot(filepath.Join(pagesDir, page, "index.pierrot"))
		if err != nil {
			log.Printf("FALHA /%s: %v", page, err)
			failed = true
			continue
		}
		html, styles, err := renderPage(p, page, comp, false, 200)
		if err != nil {
			log.Printf("FALHA /%s:\n  %v", page, err)
			failed = true
			continue
		}
		out := filepath.Join(dist, page, "index.html")
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(out, []byte(html), 0o644); err != nil {
			log.Fatal(err)
		}
		// globals.css + css da página + css de componentes -> um bundle só
		if len(styles) > 0 {
			css, err := bundleCSS(p.Src, styles)
			if err != nil {
				log.Printf("FALHA /%s: %v", page, err)
				failed = true
				continue
			}
			if err := os.WriteFile(filepath.Join(dist, page, "bundle.css"), css, 0o644); err != nil {
				log.Fatal(err)
			}
		}
		fmt.Printf("ok /%s\n", page)
	}

	// rota "/" aponta para a página default
	def := defaultPage(p)
	if data, err := os.ReadFile(filepath.Join(dist, def, "index.html")); err == nil {
		if err := os.WriteFile(filepath.Join(dist, "index.html"), data, 0o644); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ok / -> /%s\n", def)
	}

	// página de set.Fallback re-renderizada com status 404 vira o 404.html da
	// raiz (convenção dos hosts estáticos), com get.Status() devolvendo "404"
	if fb, ok := fallbackPage(p); ok {
		comp, err := readers.ParsePierrot(filepath.Join(pagesDir, fb, "index.pierrot"))
		if err != nil {
			log.Printf("FALHA 404.html: %v", err)
			failed = true
		} else if html, _, err := renderPage(p, fb, comp, false, 404); err != nil {
			log.Printf("FALHA 404.html:\n  %v", err)
			failed = true
		} else if err := os.WriteFile(filepath.Join(dist, "404.html"), []byte(html), 0o644); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("ok /404.html -> /%s\n", fb)
		}
	}

	// arquivo de set.Robots vira o /robots.txt da raiz
	if path, ok := robotsPath(p); ok {
		if data, err := os.ReadFile(path); err != nil {
			log.Printf("robots.txt: %v", err)
		} else if err := os.WriteFile(filepath.Join(dist, "robots.txt"), data, 0o644); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("ok /robots.txt")
		}
	}

	copyAssets(p.Src, dist)

	if failed {
		log.Fatal("build falhou")
	}
	fmt.Println("build pronto em", dist)
}

// bundleCSS concatena os css na ordem dos hrefs e minifica, reduzindo a
// página a uma request de css
func bundleCSS(root string, hrefs []string) ([]byte, error) {
	var b strings.Builder
	for _, href := range hrefs {
		path := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(href, "/")))
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("css %s: %w", href, err)
		}
		b.Write(data)
		b.WriteString("\n")
	}
	res := api.Transform(b.String(), api.TransformOptions{
		Loader:           api.LoaderCSS,
		MinifyWhitespace: true,
		MinifySyntax:     true,
	})
	if len(res.Errors) > 0 {
		return nil, fmt.Errorf("erro no css: %s", res.Errors[0].Text)
	}
	return res.Code, nil
}

// copyAssets copia para dist/ tudo que a página servida referencia por URL
// (css, imagens, fontes...), preservando a estrutura de pastas. Fontes do
// framework (.pierrot, .ts) já foram compiladas para dentro do HTML
func copyAssets(root, dist string) {
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if p == dist || d.Name() == "node_modules" || strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "settings.pierrot.json" {
			return nil
		}
		switch filepath.Ext(p) {
		case ".pierrot", ".ts", ".exe":
			return nil
		case ".css":
			// css entra nos bundles por página, não precisa do arquivo solto
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return nil
		}
		dst := filepath.Join(dist, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			log.Printf("asset %s: %v", rel, err)
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			log.Printf("asset %s: %v", rel, err)
			return nil
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			log.Printf("asset %s: %v", rel, err)
		}
		return nil
	})
}
