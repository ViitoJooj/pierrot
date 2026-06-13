package readers

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Component é um arquivo .pierrot parseado
type Component struct {
	Styles   []string          // css importados dentro do <script>
	Scripts  []string          // arquivos .ts/.js importados dentro do <script>
	Imports  []Import          // componentes .pierrot importados dentro do <script>
	Meta     map[string]string // chamadas set.X("...") do <script>
	Props    []string          // props: let nome(: tipo); sem valor no <script>
	Script   string            // código do <script> sem imports, set.X e props
	Template string            // HTML restante
}

// Import é um componente importado: import { Name } from "./path.pierrot";
type Import struct {
	Name string
	Path string
}

var (
	scriptRe = regexp.MustCompile(`(?s)<script>(.*?)</script>`)
	cssRe    = regexp.MustCompile(`(?m)^\s*import\s+"(.+?\.css)";?\s*$`)
	tsRe     = regexp.MustCompile(`(?m)^\s*import\s+"(.+?\.(?:ts|js))";?\s*$`)
	compRe   = regexp.MustCompile(`(?m)^\s*import\s+\{\s*(\w+)\s*\}\s+from\s+"(.+?\.pierrot)";?\s*$`)
	metaRe   = regexp.MustCompile(`(?m)^\s*set\.(\w+)\(\s*(?:"(.*?)"|(\w+))\s*\);?\s*$`)
	// let nome: tipo; SEM valor = declaração de prop do componente
	propRe = regexp.MustCompile(`(?m)^\s*let\s+(\w+)\s*(?::[^=;\n]*)?;\s*$`)
)

func ParsePierrot(path string) (*Component, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("lendo %s: %w", path, err)
	}
	return ParseSource(string(data)), nil
}

// ParseSource parseia fonte .pierrot já em memória (arquivo ou trecho de
// @render pierrot)
func ParseSource(src string) *Component {
	c := &Component{Meta: map[string]string{}}

	if m := scriptRe.FindStringSubmatch(src); m != nil {
		c.Script = m[1]
		src = scriptRe.ReplaceAllString(src, "")
	}

	for _, m := range cssRe.FindAllStringSubmatch(c.Script, -1) {
		c.Styles = append(c.Styles, m[1])
	}
	for _, m := range tsRe.FindAllStringSubmatch(c.Script, -1) {
		c.Scripts = append(c.Scripts, m[1])
	}
	for _, m := range compRe.FindAllStringSubmatch(c.Script, -1) {
		c.Imports = append(c.Imports, Import{Name: m[1], Path: m[2]})
	}
	for _, m := range propRe.FindAllStringSubmatch(c.Script, -1) {
		c.Props = append(c.Props, m[1])
	}
	// set.X("texto") guarda a string; set.X(Identificador) guarda o nome
	for _, m := range metaRe.FindAllStringSubmatch(c.Script, -1) {
		if m[3] != "" {
			c.Meta[m[1]] = m[3]
		} else {
			c.Meta[m[1]] = m[2]
		}
	}
	c.Script = cssRe.ReplaceAllString(c.Script, "")
	c.Script = tsRe.ReplaceAllString(c.Script, "")
	c.Script = compRe.ReplaceAllString(c.Script, "")
	c.Script = propRe.ReplaceAllString(c.Script, "")
	c.Script = strings.TrimSpace(metaRe.ReplaceAllString(c.Script, ""))
	c.Template = strings.TrimSpace(src)

	return c
}
