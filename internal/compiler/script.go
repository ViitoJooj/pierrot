package compiler

import "strings"

// ScriptImport é um componente importado no <script>:
// import { Name } from "./path.pierrot";
type ScriptImport struct {
	Name, Path string
}

// ScriptParts é o resultado de ParseScript: as declarações especiais da
// seção <script> separadas do código restante
type ScriptParts struct {
	Styles  []string
	Scripts []string
	Imports []ScriptImport
	Meta    map[string]string
	Props   []string
	Script  string
}

// SplitScript separa a seção <script> do template. Igual ao regex antigo:
// o conteúdo do primeiro bloco vira o script, todos os blocos <script>
// somem do template.
func SplitScript(src string) (script, template string) {
	first := true
	for {
		start := strings.Index(src, "<script>")
		if start < 0 {
			return script, src
		}
		end := strings.Index(src[start:], "</script>")
		if end < 0 {
			return script, src
		}
		if first {
			script = src[start+len("<script>") : start+end]
			first = false
		}
		src = src[:start] + src[start+end+len("</script>"):]
	}
}

// ParseScript classifica as linhas do script: imports de css/ts/componente,
// set.X(...) e props `let nome;` saem para os campos; o resto fica em Script
func ParseScript(src string) ScriptParts {
	p := ScriptParts{Meta: map[string]string{}}
	var keep []string
	for line := range strings.SplitSeq(src, "\n") {
		t := strings.TrimSpace(line)
		if path, ok := parseFileImport(t); ok {
			switch {
			case strings.HasSuffix(path, ".css"):
				p.Styles = append(p.Styles, path)
				continue
			case strings.HasSuffix(path, ".ts"), strings.HasSuffix(path, ".js"):
				p.Scripts = append(p.Scripts, path)
				continue
			}
			// import de outra coisa não é declaração pierrot: fica no código
		}
		if imp, ok := parseComponentImport(t); ok {
			p.Imports = append(p.Imports, imp)
			continue
		}
		if key, val, ok := parseSetCall(t); ok {
			p.Meta[key] = val
			continue
		}
		if name, ok := parsePropDecl(t); ok {
			p.Props = append(p.Props, name)
			continue
		}
		keep = append(keep, line)
	}
	p.Script = strings.TrimSpace(strings.Join(keep, "\n"))
	return p
}

// parseFileImport casa `import "caminho";` (; opcional) e devolve o caminho
func parseFileImport(t string) (string, bool) {
	rest, ok := cutKeyword(t, "import")
	if !ok || len(rest) < 2 || rest[0] != '"' {
		return "", false
	}
	end := strings.IndexByte(rest[1:], '"')
	if end < 0 {
		return "", false
	}
	return rest[1 : 1+end], trailerOK(rest[end+2:])
}

// parseComponentImport casa `import { Nome } from "x.pierrot";` (; opcional)
func parseComponentImport(t string) (ScriptImport, bool) {
	rest, ok := cutKeyword(t, "import")
	if !ok || len(rest) == 0 || rest[0] != '{' {
		return ScriptImport{}, false
	}
	close := strings.IndexByte(rest, '}')
	if close < 0 {
		return ScriptImport{}, false
	}
	name := strings.TrimSpace(rest[1:close])
	if !isWordName(name) {
		return ScriptImport{}, false
	}
	rest, ok = cutKeyword(strings.TrimSpace(rest[close+1:]), "from")
	if !ok || len(rest) < 2 || rest[0] != '"' {
		return ScriptImport{}, false
	}
	end := strings.IndexByte(rest[1:], '"')
	if end < 0 {
		return ScriptImport{}, false
	}
	path := rest[1 : 1+end]
	if !strings.HasSuffix(path, ".pierrot") || !trailerOK(rest[end+2:]) {
		return ScriptImport{}, false
	}
	return ScriptImport{Name: name, Path: path}, true
}

// parseSetCall casa `set.Chave("texto");` ou `set.Chave(Ident);` (; opcional)
func parseSetCall(t string) (key, val string, ok bool) {
	rest, found := strings.CutPrefix(t, "set.")
	if !found {
		return "", "", false
	}
	j := 0
	for j < len(rest) && isWordChar(rest[j]) {
		j++
	}
	if j == 0 || j >= len(rest) || rest[j] != '(' {
		return "", "", false
	}
	key = rest[:j]
	inner := strings.TrimSpace(rest[j+1:])
	if len(inner) > 0 && inner[0] == '"' {
		end := strings.IndexByte(inner[1:], '"')
		if end < 0 {
			return "", "", false
		}
		val = inner[1 : 1+end]
		inner = strings.TrimSpace(inner[end+2:])
	} else {
		k := 0
		for k < len(inner) && isWordChar(inner[k]) {
			k++
		}
		if k == 0 {
			return "", "", false
		}
		val = inner[:k]
		inner = strings.TrimSpace(inner[k:])
	}
	if len(inner) == 0 || inner[0] != ')' || !trailerOK(inner[1:]) {
		return "", "", false
	}
	return key, val, true
}

// parsePropDecl casa `let nome;` ou `let nome: tipo;` sem valor (`=` no tipo
// desclassifica, igual ao propRe antigo)
func parsePropDecl(t string) (string, bool) {
	rest, ok := cutKeyword(t, "let")
	if !ok {
		return "", false
	}
	j := 0
	for j < len(rest) && isWordChar(rest[j]) {
		j++
	}
	if j == 0 {
		return "", false
	}
	name := rest[:j]
	tail := strings.TrimSpace(rest[j:])
	if tail == ";" {
		return name, true
	}
	if len(tail) > 0 && tail[0] == ':' {
		typ := tail[1:]
		end := strings.IndexByte(typ, ';')
		if end < 0 || strings.ContainsAny(typ[:end], "=") || !trailerOK(typ[end+1:]) {
			return "", false
		}
		return name, true
	}
	return "", false
}

// trailerOK aceita o fim de linha das declarações: nada ou `;` (com espaços)
func trailerOK(s string) bool {
	s = strings.TrimSpace(s)
	return s == "" || s == ";"
}
