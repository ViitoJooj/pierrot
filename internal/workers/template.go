package workers

import (
	"fmt"
	"regexp"
	"strings"
)

// Diretivas de template: @for/@if viram funções JS que rendem o bloco no
// navegador, porque os dados (let do <script>) só existem lá. Cada bloco vira
// um placeholder <div data-pierrot-block="N"> preenchido (e re-renderizado a
// cada __pierrotUpdate) pelo runtime. Interpolação ${expr} fora de bloco vira
// <span data-pierrot-expr="N"> atualizado do mesmo jeito.
var (
	commentLineRe = regexp.MustCompile(`(?m)^[ \t]*//.*\r?\n?`)
	forLineRe     = regexp.MustCompile(`^\s*@for\s+(\w+)\s+in\s+(.+?)\s*$`)
	ifLineRe      = regexp.MustCompile(`^\s*@if\s+(.+?)\s*$`)
	elseLineRe    = regexp.MustCompile(`^\s*@else\s*$`)
	endLineRe     = regexp.MustCompile(`^\s*@(?:end|endif)\s*$`)
	interpRe      = regexp.MustCompile(`\$\{([^{}\n]+)\}`)
	identRe       = regexp.MustCompile(`^[A-Za-z_$][\w$]*$`)
)

// compiledTemplate é o resultado de compileTemplate: o HTML com placeholders
// e o JS que o runtime usa para preenchê-los
type compiledTemplate struct {
	html   string
	blocks []string // corpo da função JS de cada bloco @for/@if, na ordem dos placeholders
	exprs  []string // expressão JS de cada ${expr} fora de bloco, na ordem dos spans
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
	// @click={fn} dentro de bloco vira o mesmo onclick gerado fora dele
	line = eventRe.ReplaceAllString(line, `on$1="$2(); __pierrotUpdate()"`)
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
