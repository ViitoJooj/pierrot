package workers

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
)

// Props de componente: <Card title="literal" price={expr} />. O componente
// declara as props como `let nome: tipo;` sem valor no <script> (parser.go,
// propRe). A expansão é textual, por instância: literal entra como texto
// direto no HTML do filho; {expr} vira ${expr} e o browser avalia — dentro de
// @for/@if a expressão roda por iteração (ex.: {card.title}), fora de bloco
// vira um span preenchido pelo runtime.
//
// Limites: prop não chega no <script> do componente (o JS é compartilhado
// entre as instâncias) e prop expressão em atributo só funciona dentro de
// bloco @for/@if (fora, o valor viraria um <span> dentro do atributo).

// prop é o valor passado num atributo da tag do componente
type prop struct {
	text    string // valor de atributo com aspas, usado como texto direto
	expr    string // expressão JS equivalente, para uso dentro de ${...}
	literal bool
}

var (
	// nome="literal" ou nome={expr}
	attrRe = regexp.MustCompile(`(\w+)\s*=\s*(?:"([^"]*)"|\{([^{}\n]*)\})`)
	// atributo={expr} no template do filho; [^@\w] de prefixo pula @click={fn}
	attrInterpRe = regexp.MustCompile(`([^@\w])([A-Za-z][\w-]*)\s*=\s*\$?\{([^{}\n]+)\}`)
	// {expr} ou ${expr} no texto do template do filho; o prefixo [^=] pula
	// formas de atributo já expandidas (@bind={...}, @evento={...})
	spanInterpRe = regexp.MustCompile(`(^|[^=])\$?\{([^{}\n]+)\}`)
)

// parseProps lê os atributos da tag de uma instância de componente
func parseProps(attrs string) map[string]prop {
	out := map[string]prop{}
	for _, m := range attrRe.FindAllStringSubmatch(attrs, -1) {
		val := strings.TrimSpace(m[0][strings.Index(m[0], "=")+1:])
		if strings.HasPrefix(val, `"`) {
			out[m[1]] = prop{text: m[2], expr: strconv.Quote(m[2]), literal: true}
		} else {
			out[m[1]] = prop{expr: "(" + strings.TrimSpace(m[3]) + ")"}
		}
	}
	return out
}

// expandEvents troca @evento={fn} ou @evento={fn(args)} pelo atributo
// on<evento> nativo + atualização dos bindings. Aspas nos args (props literais
// já substituídas) são escapadas para o valor do atributo HTML
func expandEvents(s string) string {
	return eventRe.ReplaceAllStringFunc(s, func(m string) string {
		g := eventRe.FindStringSubmatch(m)
		return fmt.Sprintf(`on%s="%s(%s); __pierrotUpdate()"`, g[1], g[2], html.EscapeString(g[3]))
	})
}

// applyProps expande uma instância: cada referência às props declaradas no
// template já renderizado do filho vira o valor do atributo da tag. Prop
// declarada e não passada vira ""
func applyProps(childHTML string, declared []string, vals map[string]prop) string {
	if len(declared) == 0 {
		return childHTML
	}
	props := map[string]prop{}
	for _, name := range declared {
		if v, ok := vals[name]; ok {
			props[name] = v
		} else {
			props[name] = prop{text: "", expr: `""`, literal: true}
		}
	}

	// @evento={fn(args)}: prop nos args vira o valor da instância — é o único
	// jeito de prop chegar no <script> do componente (como argumento). Roda
	// antes dos outros passes para o spanInterpRe não reescrever {fn(prop)}
	out := eventRe.ReplaceAllStringFunc(childHTML, func(m string) string {
		g := eventRe.FindStringSubmatch(m)
		if g[3] == "" {
			return m
		}
		sub, changed := substIdents(g[3], props)
		if !changed {
			return m
		}
		return fmt.Sprintf("@%s={%s(%s)}", g[1], g[2], sub)
	})

	// @render nome(arg): prop no arg vira o valor da instância, igual aos
	// args de @evento — literal entra como string quotada (que o expandRenders
	// do pierrot consegue desquotar)
	out = renderLineRe.ReplaceAllStringFunc(out, func(m string) string {
		g := renderLineRe.FindStringSubmatch(m)
		sub, changed := substIdents(strings.TrimSpace(g[2]), props)
		if !changed {
			return m
		}
		return fmt.Sprintf("@render %s(%s);", g[1], sub)
	})

	// @bind={prop}: o alvo vira a expressão da instância (ex. @bind={(code)}),
	// antes do spanInterpRe reescrever o {prop}
	out = bindRe.ReplaceAllStringFunc(out, func(m string) string {
		g := bindRe.FindStringSubmatch(m)
		sub, changed := substIdents(strings.TrimSpace(g[1]), props)
		if !changed {
			return m
		}
		return "@bind={" + sub + "}"
	})

	// atributo={prop} ganha aspas: literal entra direto, expressão viram
	// atributo="${expr}" (o emitText de bloco escapa o valor)
	out = attrInterpRe.ReplaceAllStringFunc(out, func(m string) string {
		g := attrInterpRe.FindStringSubmatch(m)
		inner := strings.TrimSpace(g[3])
		if v, ok := props[inner]; ok && v.literal {
			return fmt.Sprintf(`%s%s="%s"`, g[1], g[2], html.EscapeString(v.text))
		}
		sub, changed := substIdents(inner, props)
		if !changed {
			return m
		}
		return fmt.Sprintf(`%s%s="${%s}"`, g[1], g[2], sub)
	})

	// {expr} / ${expr} no texto: literal sozinho vira texto direto; o resto
	// vira ${expr} com as props substituídas, avaliado no browser
	out = spanInterpRe.ReplaceAllStringFunc(out, func(m string) string {
		prefix := ""
		if m[0] != '$' && m[0] != '{' {
			prefix = m[:1]
		}
		inner := strings.TrimSpace(m[strings.Index(m, "{")+1 : len(m)-1])
		if v, ok := props[inner]; ok && v.literal {
			return prefix + html.EscapeString(v.text)
		}
		sub, changed := substIdents(inner, props)
		if !changed {
			return m
		}
		return prefix + "${" + sub + "}"
	})
	return out
}

// substIdents troca cada identificador de prop dentro da expressão pelo valor;
// uma passada só, para o valor inserido não ser re-substituído. Identificador
// precedido de "." (campo de objeto, ex. card.title) fica como está
func substIdents(expr string, props map[string]prop) (string, bool) {
	names := make([]string, 0, len(props))
	for n := range props {
		names = append(names, n)
	}
	re := regexp.MustCompile(`(^|[^.\w$])(` + strings.Join(names, "|") + `)\b`)
	changed := false
	expr = re.ReplaceAllStringFunc(expr, func(m string) string {
		g := re.FindStringSubmatch(m)
		changed = true
		return g[1] + props[g[2]].expr
	})
	return expr, changed
}
