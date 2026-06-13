package workers

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/ViitoJooj/pierrot/internal/readers"
)

// Diretivas de template: @for/@if viram funções JS que rendem o bloco no
// navegador, porque os dados (let do <script>) só existem lá. Cada bloco vira
// um placeholder <div data-pierrot-block="N"> preenchido (e re-renderizado a
// cada __pierrotUpdate) pelo runtime. Interpolação ${expr} fora de bloco vira
// <span data-pierrot-expr="N"> atualizado do mesmo jeito.
var (
	// comentário só com // no primeiro caractere da linha; indentado vira texto
	commentLineRe = regexp.MustCompile(`(?m)^//.*\r?\n?`)
	forLineRe     = regexp.MustCompile(`^\s*@for\s+(\w+)\s+in\s+(.+?)\s*$`)
	ifLineRe      = regexp.MustCompile(`^\s*@if\s+(.+?)\s*$`)
	elseLineRe    = regexp.MustCompile(`^\s*@else\s*$`)
	endLineRe     = regexp.MustCompile(`^\s*@(?:end|endif)\s*$`)
	interpRe      = regexp.MustCompile(`\$\{([^{}\n]+)\}`)
	identRe       = regexp.MustCompile(`^[A-Za-z_$][\w$]*$`)
	// @render nome(arg); em linha própria. O arg é ganancioso até o último ")"
	// da linha, para string literal com parênteses dentro das aspas funcionar
	renderLineRe = regexp.MustCompile(`(?m)^[ \t]*@render\s+(\w+)\s*\((.*)\)\s*;?[ \t\r]*$`)
	// @bind={var}: two-way binding de input/textarea com uma variável do <script>
	bindRe = regexp.MustCompile(`@bind=\{([^{}\n]+)\}`)
)

// renderEntry é um @render html/markdown avaliado pelo runtime no browser
type renderEntry struct {
	kind string // "html" ou "markdown"
	expr string
}

// expandRenders processa as linhas @render antes de compileTemplate.
// pierrot com string literal é resolvido em compilação: o trecho é parseado
// como um mini-componente — o <script> dele entra no bundle da página e o
// template é inlined no corpo. pierrot com expressão vira um iframe de preview
// re-compilado no browser quando o valor muda (debounce de 300ms). html e
// markdown viram placeholders <div data-pierrot-render="N"> preenchidos pelo
// runtime a cada __pierrotUpdate
func (rc *renderCtx) expandRenders(body string) (string, []renderEntry) {
	var entries []renderEntry
	body = renderLineRe.ReplaceAllStringFunc(body, func(m string) string {
		g := renderLineRe.FindStringSubmatch(m)
		kind, arg := g[1], strings.TrimSpace(g[2])
		switch kind {
		case "pierrot":
			if src, ok := staticString(arg); ok {
				frag := readers.ParseSource(src)
				if frag.Script != "" {
					rc.scripts = append(rc.scripts, chunk{name: "@render pierrot", code: frag.Script})
				}
				return frag.Template
			}
			fallthrough
		case "html", "markdown":
			id := len(entries)
			// mesmo guard de typeof dos bindings: variável não declarada vira ""
			if identRe.MatchString(arg) {
				arg = fmt.Sprintf(`typeof %s === "undefined" ? "" : %s`, arg, arg)
			}
			entries = append(entries, renderEntry{kind: kind, expr: arg})
			return fmt.Sprintf(`<div data-pierrot-render="%d" style="display:contents"></div>`, id)
		default:
			rc.errs = append(rc.errs, fmt.Sprintf("@render %s: renderizador desconhecido (use html, markdown ou pierrot)", kind))
			return ""
		}
	})
	return body, entries
}

// staticString reconhece um argumento que é string literal JS (a substituição
// de prop literal gera um deles, possivelmente entre parênteses)
func staticString(arg string) (string, bool) {
	arg = trimBalancedParens(arg)
	if len(arg) >= 2 {
		switch arg[0] {
		case '\'', '"', '`':
			if arg[len(arg)-1] == arg[0] {
				return jsUnquote(arg), true
			}
		}
	}
	return "", false
}

// expandBinds troca @bind={var} pelo oninput que escreve no alvo + um marcador
// que o runtime usa para preencher o valor do elemento a partir da variável
// (pulando o elemento focado, para não brigar com o usuário digitando)
func (ct *compiledTemplate) expandBinds(s string) string {
	return bindRe.ReplaceAllStringFunc(s, func(m string) string {
		g := bindRe.FindStringSubmatch(m)
		target := strings.TrimSpace(g[1])
		id := len(ct.binds)
		ct.binds = append(ct.binds, target)
		return fmt.Sprintf(`oninput="%s = this.value; __pierrotUpdate()" data-pierrot-bindval="%d"`, html.EscapeString(target), id)
	})
}

// trimBalancedParens tira um par externo de parênteses, se de fato envolver o
// argumento todo (ex. "(code)" sim, "(a)+(b)" não)
func trimBalancedParens(s string) string {
	for strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		depth := 0
		for i, r := range s {
			switch r {
			case '(':
				depth++
			case ')':
				depth--
			}
			if depth == 0 && i < len(s)-1 {
				return s
			}
		}
		s = strings.TrimSpace(s[1 : len(s)-1])
	}
	return s
}

// jsUnquote decodifica uma string literal JS ('...', "..." ou `...`):
// remove as aspas e resolve os escapes comuns (\n, \t, \r, \\, \', \", \/)
func jsUnquote(s string) string {
	if len(s) < 2 {
		return s
	}
	s = s[1 : len(s)-1]
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// compiledTemplate é o resultado de compileTemplate: o HTML com placeholders
// e o JS que o runtime usa para preenchê-los
type compiledTemplate struct {
	html   string
	blocks []string // corpo da função JS de cada bloco @for/@if, na ordem dos placeholders
	exprs  []string // expressão JS de cada ${expr} fora de bloco, na ordem dos spans
	binds  []string // alvo JS de cada @bind={var}, na ordem dos data-pierrot-bindval
	errs   []string
}

// compileTemplate remove comentários // de linha e extrai os blocos de
// diretiva do corpo já montado (layout + página + componentes expandidos)
func compileTemplate(body string) *compiledTemplate {
	ct := &compiledTemplate{}
	body = commentLineRe.ReplaceAllString(body, "")

	lines := strings.Split(body, "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if forLineRe.MatchString(line) || ifLineRe.MatchString(line) {
			// acha o fechamento do bloco, contando os aninhados
			depth := 1
			j := i + 1
			for ; j < len(lines); j++ {
				switch {
				case forLineRe.MatchString(lines[j]) || ifLineRe.MatchString(lines[j]):
					depth++
				case endLineRe.MatchString(lines[j]):
					depth--
				}
				if depth == 0 {
					break
				}
			}
			if depth != 0 {
				ct.errs = append(ct.errs, fmt.Sprintf("diretiva %q sem @end/@endif", strings.TrimSpace(line)))
				out = append(out, line)
				continue
			}
			id := len(ct.blocks)
			ct.blocks = append(ct.blocks, ct.compileBlock(lines[i:j+1]))
			out = append(out, fmt.Sprintf(`<div data-pierrot-block="%d" style="display:contents"></div>`, id))
			i = j
			continue
		}
		if elseLineRe.MatchString(line) || endLineRe.MatchString(line) {
			ct.errs = append(ct.errs, fmt.Sprintf("%q sem @for/@if correspondente", strings.TrimSpace(line)))
			continue
		}
		out = append(out, line)
	}
	ct.html = strings.Join(out, "\n")
	return ct
}

// compileBlock traduz as linhas de um bloco (da diretiva de abertura até o
// fechamento) para o corpo de uma função JS que devolve o HTML do bloco
func (ct *compiledTemplate) compileBlock(lines []string) string {
	var b strings.Builder
	b.WriteString("var __h = \"\";\n")
	for _, line := range lines {
		switch {
		case forLineRe.MatchString(line):
			m := forLineRe.FindStringSubmatch(line)
			fmt.Fprintf(&b, "for (const %s of (%s)) {\n", m[1], m[2])
		case ifLineRe.MatchString(line):
			m := ifLineRe.FindStringSubmatch(line)
			fmt.Fprintf(&b, "if (%s) {\n", m[1])
		case elseLineRe.MatchString(line):
			b.WriteString("} else {\n")
		case endLineRe.MatchString(line):
			b.WriteString("}\n")
		default:
			if strings.TrimSpace(line) == "" {
				continue
			}
			ct.emitText(line, &b)
		}
	}
	b.WriteString("return __h;")
	return b.String()
}

// emitText vira uma linha de HTML do bloco em concatenação JS, trocando cada
// ${expr} pela expressão escapada. %q do Go gera um literal de string válido
// em JS para o trecho fixo
func (ct *compiledTemplate) emitText(line string, b *strings.Builder) {
	// @bind antes de @evento: o eventRe casaria @bind={var} como evento
	line = ct.expandBinds(line)
	// @click={fn} dentro de bloco vira o mesmo onclick gerado fora dele
	line = expandEvents(line)
	b.WriteString("__h += ")
	last := 0
	for _, loc := range interpRe.FindAllStringSubmatchIndex(line, -1) {
		expr := strings.TrimSpace(line[loc[2]:loc[3]])
		// ${var} simples ganha o mesmo guard de typeof do state do runtime:
		// variável não declarada vira "" em vez de ReferenceError no bloco todo
		if identRe.MatchString(expr) {
			expr = fmt.Sprintf(`typeof %s === "undefined" ? "" : %s`, expr, expr)
		}
		fmt.Fprintf(b, "%q + __pierrotEsc(%s) + ", line[last:loc[0]], expr)
		last = loc[1]
	}
	fmt.Fprintf(b, "%q;\n", line[last:]+"\n")
}

// replaceExprs troca cada ${expr} fora de bloco por um span que o runtime
// preenche. Precisa rodar depois de bindingRe, que trata os ${var} simples;
// aqui ficam só as expressões compostas (ex.: ${obj.campo})
func (ct *compiledTemplate) replaceExprs(body string) string {
	return interpRe.ReplaceAllStringFunc(body, func(m string) string {
		id := len(ct.exprs)
		// m é ${expr}: corta o "${" e o "}"
		ct.exprs = append(ct.exprs, m[2:len(m)-1])
		return fmt.Sprintf(`<span data-pierrot-expr="%d"></span>`, id)
	})
}
