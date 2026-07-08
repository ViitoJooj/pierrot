package compiler

import (
	"html"
	"strconv"
	"strings"
)

// Expansão de componentes como operação de árvore, com a mesma semântica do
// applyProps textual antigo: prop literal entra como texto/atributo direto
// (escapada); prop expressão vira interpolação avaliada no browser; prop
// declarada e não passada vira "". A árvore do filho nunca é mutada — cada
// instância reconstrói os nós no caminho da substituição.

// PropVal é o valor de uma prop numa instância
type PropVal struct {
	Text    string // valor literal, usado como texto direto
	Expr    string // expressão JS equivalente, para uso em interpolação
	Literal bool
}

// ExpandComponent devolve os nós do filho com as props da instância
// aplicadas. Sem props declaradas, devolve a árvore do filho como está.
func ExpandComponent(child *Template, declared []string, inst *ComponentInst) []Node {
	if len(declared) == 0 {
		return child.Children
	}
	props := buildProps(declared, inst.Props)
	return substNodes(child.Children, props)
}

// buildProps monta o mapa de props: passada literal, passada expressão
// (parênteses em volta, igual parseProps antigo) ou ausente ("")
func buildProps(declared []string, given []Attr) map[string]PropVal {
	vals := map[string]PropVal{}
	for _, a := range given {
		switch {
		case a.Quoted:
			vals[a.Name] = PropVal{Text: a.Val, Expr: strconv.Quote(a.Val), Literal: true}
		case a.Expr:
			vals[a.Name] = PropVal{Expr: "(" + strings.TrimSpace(a.Val) + ")"}
		}
	}
	props := map[string]PropVal{}
	for _, name := range declared {
		if v, ok := vals[name]; ok {
			props[name] = v
		} else {
			props[name] = PropVal{Text: "", Expr: `""`, Literal: true}
		}
	}
	return props
}

func substNodes(nodes []Node, props map[string]PropVal) []Node {
	var out []Node
	for _, n := range nodes {
		out = append(out, substNode(n, props)...)
	}
	return out
}

func substNode(n Node, props map[string]PropVal) []Node {
	switch v := n.(type) {
	case *Text:
		return substText(v, props)
	case *Interp:
		return []Node{substInterp(v, props)}
	case *Element:
		el := *v
		el.Attrs = substAttrs(v.Attrs, props)
		el.Events = substEvents(v.Events, props)
		if v.HasBind {
			if sub, changed := SubstIdents(strings.TrimSpace(v.Bind), props); changed {
				el.Bind = sub
			}
		}
		el.Children = substNodes(v.Children, props)
		return []Node{&el}
	case *ForBlock:
		fb := *v
		fb.Children = substNodes(v.Children, props)
		return []Node{&fb}
	case *IfBlock:
		ib := *v
		ib.Then = substNodes(v.Then, props)
		ib.Else = substNodes(v.Else, props)
		return []Node{&ib}
	case *ComponentInst:
		ci := *v
		ci.Props = substAttrs(v.Props, props)
		return []Node{&ci}
	default: // *Slot
		return []Node{n}
	}
}

// substInterp aplica props numa interpolação ${expr}: prop literal sozinha
// vira texto escapado; expressão com prop dentro vira a expressão da
// instância; sem prop referenciada fica como está
func substInterp(in *Interp, props map[string]PropVal) Node {
	trimmed := strings.TrimSpace(in.Expr)
	if v, ok := props[trimmed]; ok && v.Literal {
		return &Text{Raw: html.EscapeString(v.Text), Pos: in.Pos}
	}
	if sub, changed := SubstIdents(trimmed, props); changed {
		return &Interp{Expr: sub, Pos: in.Pos}
	}
	return in
}

// substText aplica props nas formas de chave crua do texto: {expr} — a
// forma ${expr} já virou Interp no lexer. {x} precedido de = fica (guarda
// do spanInterpRe antigo). Pode devolver mistura de Text e Interp.
func substText(t *Text, props map[string]PropVal) []Node {
	raw := t.Raw
	var out []Node
	var lit strings.Builder
	for {
		start, inner := nextBareBrace(raw)
		if start < 0 {
			break
		}
		trimmed := strings.TrimSpace(inner)
		end := start + len(inner) + 2
		if v, ok := props[trimmed]; ok && v.Literal {
			lit.WriteString(raw[:start])
			lit.WriteString(html.EscapeString(v.Text))
			raw = raw[end:]
			continue
		}
		if sub, changed := SubstIdents(trimmed, props); changed {
			lit.WriteString(raw[:start])
			if lit.Len() > 0 {
				out = append(out, &Text{Raw: lit.String(), Pos: t.Pos})
				lit.Reset()
			}
			out = append(out, &Interp{Expr: sub, Pos: t.Pos})
			raw = raw[end:]
			continue
		}
		lit.WriteString(raw[:end])
		raw = raw[end:]
	}
	lit.WriteString(raw)
	if lit.Len() > 0 || len(out) == 0 {
		out = append(out, &Text{Raw: lit.String(), Pos: t.Pos})
	}
	return out
}

// nextBareBrace acha o próximo {inner} válido (sem chaves/quebra dentro,
// não vazio, não precedido de = nem de $)
func nextBareBrace(s string) (start int, inner string) {
	for i := 0; i < len(s); i++ {
		if s[i] != '{' {
			continue
		}
		if i > 0 && (s[i-1] == '=' || s[i-1] == '$') {
			continue
		}
		for j := i + 1; j < len(s); j++ {
			switch s[j] {
			case '}':
				if j > i+1 {
					return i, s[i+1 : j]
				}
				i = j
				goto next
			case '{', '\n':
				i = j - 1
				goto next
			}
		}
		return -1, ""
	next:
	}
	return -1, ""
}

// substAttrs aplica props nos valores de atributo: nome={prop} literal
// ganha aspas com o texto; expressão vira nome="${expr}"; valor com aspas
// tem os {x}/${x} internos substituídos
func substAttrs(attrs []Attr, props map[string]PropVal) []Attr {
	out := make([]Attr, len(attrs))
	for i, a := range attrs {
		out[i] = substAttr(a, props)
	}
	return out
}

func substAttr(a Attr, props map[string]PropVal) Attr {
	switch {
	case a.Expr:
		inner := strings.TrimSpace(a.Val)
		if v, ok := props[inner]; ok && v.Literal {
			return Attr{Name: a.Name, Val: html.EscapeString(v.Text), HasVal: true, Quoted: true, Quote: '"', Pos: a.Pos}
		}
		if sub, changed := SubstIdents(inner, props); changed {
			return Attr{Name: a.Name, Val: "${" + sub + "}", HasVal: true, Quoted: true, Quote: '"', Pos: a.Pos}
		}
		return a
	case a.Quoted:
		a.Val = substAttrVal(a.Val, props)
		return a
	default:
		return a
	}
}

// substAttrVal troca {x} e ${x} dentro de um valor de atributo com aspas
func substAttrVal(val string, props map[string]PropVal) string {
	var b strings.Builder
	for {
		start, inner, dollar := nextAnyBrace(val)
		if start < 0 {
			b.WriteString(val)
			return b.String()
		}
		end := start + len(inner) + 2
		if dollar {
			end++
		}
		trimmed := strings.TrimSpace(inner)
		if v, ok := props[trimmed]; ok && v.Literal {
			b.WriteString(val[:start])
			b.WriteString(html.EscapeString(v.Text))
		} else if sub, changed := SubstIdents(trimmed, props); changed {
			b.WriteString(val[:start])
			b.WriteString("${")
			b.WriteString(sub)
			b.WriteString("}")
		} else {
			b.WriteString(val[:end])
		}
		val = val[end:]
	}
}

// nextAnyBrace acha o próximo {inner} ou ${inner} (guarda de = só para a
// forma sem $, igual ao spanInterpRe antigo)
func nextAnyBrace(s string) (start int, inner string, dollar bool) {
	for i := 0; i < len(s); i++ {
		if s[i] != '{' {
			continue
		}
		d := i > 0 && s[i-1] == '$'
		if !d && i > 0 && s[i-1] == '=' {
			continue
		}
		for j := i + 1; j < len(s); j++ {
			switch s[j] {
			case '}':
				if j > i+1 {
					if d {
						return i - 1, s[i+1 : j], true
					}
					return i, s[i+1 : j], false
				}
				i = j
				goto next
			case '{', '\n':
				i = j - 1
				goto next
			}
		}
		return -1, "", false
	next:
	}
	return -1, "", false
}

func substEvents(events []Event, props map[string]PropVal) []Event {
	out := make([]Event, len(events))
	for i, ev := range events {
		out[i] = ev
		if ev.Args == "" {
			continue
		}
		if sub, changed := SubstIdents(ev.Args, props); changed {
			out[i].Args = sub
		}
	}
	return out
}

// SubstIdents troca cada identificador de prop dentro da expressão pelo
// valor da instância, numa passada só (o valor inserido não é re-visitado).
// Identificador precedido de "." (campo, ex. card.title) ou de char de
// palavra/$ fica como está.
func SubstIdents(expr string, props map[string]PropVal) (string, bool) {
	var b strings.Builder
	changed := false
	i := 0
	for i < len(expr) {
		c := expr[i]
		if !isWordChar(c) && c != '$' {
			b.WriteByte(c)
			i++
			continue
		}
		start := i
		for i < len(expr) && (isWordChar(expr[i]) || expr[i] == '$') {
			i++
		}
		word := expr[start:i]
		prevOK := start == 0 || !isPrevIdentChar(expr[start-1])
		if v, ok := props[word]; ok && prevOK {
			b.WriteString(v.Expr)
			changed = true
			continue
		}
		b.WriteString(word)
	}
	return b.String(), changed
}

// isPrevIdentChar espelha o [^.\w$] do substIdents antigo
func isPrevIdentChar(c byte) bool {
	return c == '.' || c == '$' || isWordChar(c)
}

