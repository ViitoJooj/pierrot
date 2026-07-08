package compiler

import (
	"fmt"
	"html"
	"strings"
)

// Emitted é o que o renderPage consome: o HTML com placeholders e as
// funções que o runtime usa para preenchê-los (mesmo contrato do pipeline
// antigo de compileTemplate + expandBinds + expandEvents + replaceExprs)
type Emitted struct {
	HTML      string
	Blocks    []string // corpo JS de cada bloco @for/@if, na ordem dos placeholders
	Exprs     []string // expressão de cada ${expr} composto fora de bloco
	Binds     []string // alvo de cada @bind={var}, na ordem dos data-pierrot-bindval
	BindNames []string // nomes de ${var} fora de bloco, dedup, para o state do runtime
	HasEvents bool     // algum @evento no HTML (o handler chama __pierrotUpdate)
	Errs      []string
}

// NeedsRuntime informa se a página precisa do runtime de bindings
// (__pierrotUpdate etc.): qualquer binding, bloco, expr, bind ou evento
func (e *Emitted) NeedsRuntime() bool {
	return len(e.BindNames) > 0 || len(e.Blocks) > 0 || len(e.Exprs) > 0 ||
		len(e.Binds) > 0 || e.HasEvents
}

// Emit percorre a árvore e gera os artefatos. A numeração de blocos/exprs/
// binds é própria e consistente entre HTML e listas (o pareamento é o que
// importa para o runtime, não os números em si).
func Emit(t *Template) *Emitted {
	em := &emitter{seenBind: map[string]bool{}}
	var b strings.Builder
	sink := &htmlSink{em: em, b: &b}
	for _, n := range t.Children {
		em.node(sink, n)
	}
	return &Emitted{
		HTML: b.String(), Blocks: em.blocks, Exprs: em.exprs,
		Binds: em.binds, BindNames: em.bindNames, HasEvents: em.hasEvents,
		Errs: t.Errs,
	}
}

type emitter struct {
	blocks, exprs, binds, bindNames []string
	seenBind                        map[string]bool
	hasEvents                       bool
}

// sink recebe os pedaços da emissão. Fora de bloco os pedaços viram HTML
// direto (interp vira span); dentro de bloco viram concatenação JS em __h
// (interp vira __pierrotEsc(expr))
type sink interface {
	lit(s string)
	interp(expr string)
}

// ---------- fora de bloco ----------

type htmlSink struct {
	em *emitter
	b  *strings.Builder
}

func (s *htmlSink) lit(str string) { s.b.WriteString(str) }

// interp fora de bloco: ${ident} vira <span data-bind> + registro do nome;
// expressão composta vira <span data-pierrot-expr="N"> (expr crua, sem trim)
func (s *htmlSink) interp(expr string) {
	if isWordName(expr) {
		if !s.em.seenBind[expr] {
			s.em.seenBind[expr] = true
			s.em.bindNames = append(s.em.bindNames, expr)
		}
		fmt.Fprintf(s.b, `<span data-bind="%s"></span>`, expr)
		return
	}
	id := len(s.em.exprs)
	s.em.exprs = append(s.em.exprs, expr)
	fmt.Fprintf(s.b, `<span data-pierrot-expr="%d"></span>`, id)
}

// ---------- dentro de bloco ----------

// jsSink acumula literais e expressões e descarrega statements __h += ...
type jsSink struct {
	js    *strings.Builder
	parts []string // já em sintaxe JS: "litQuoted" ou __pierrotEsc(expr)
}

func (s *jsSink) lit(str string) {
	if str == "" {
		return
	}
	// junta literais consecutivos antes de quotar? quotar por pedaço é
	// equivalente em JS; acumula cru e quota no flush
	if n := len(s.parts); n > 0 && strings.HasPrefix(s.parts[n-1], "\x00") {
		s.parts[n-1] += str
		return
	}
	s.parts = append(s.parts, "\x00"+str) // \x00 marca literal cru
}

func (s *jsSink) interp(expr string) {
	trimmed := strings.TrimSpace(expr)
	// ${var} simples ganha o guard de typeof, igual ao emitText antigo:
	// variável não declarada vira "" em vez de ReferenceError no bloco todo
	if isJSIdent(trimmed) {
		trimmed = fmt.Sprintf(`typeof %s === "undefined" ? "" : %s`, trimmed, trimmed)
	}
	s.parts = append(s.parts, fmt.Sprintf("__pierrotEsc(%s)", trimmed))
}

// flush emite __h += pedaços;
func (s *jsSink) flush() {
	if len(s.parts) == 0 {
		return
	}
	s.js.WriteString("__h += ")
	for i, p := range s.parts {
		if i > 0 {
			s.js.WriteString(" + ")
		}
		if strings.HasPrefix(p, "\x00") {
			fmt.Fprintf(s.js, "%q", p[1:])
		} else {
			s.js.WriteString(p)
		}
	}
	s.js.WriteString(";\n")
	s.parts = s.parts[:0]
}

// isJSIdent espelha o identRe antigo: ^[A-Za-z_$][\w$]*$
func isJSIdent(s string) bool {
	if s == "" {
		return false
	}
	c := s[0]
	if !isAlpha(c) && c != '_' && c != '$' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !isWordChar(s[i]) && s[i] != '$' {
			return false
		}
	}
	return true
}

// ---------- nós ----------

func (em *emitter) node(s sink, n Node) {
	switch v := n.(type) {
	case *Text:
		s.lit(v.Raw)
	case *Interp:
		s.interp(v.Expr)
	case *Element:
		em.element(s, v)
	case *ForBlock, *IfBlock:
		em.block(s, n)
	case *ComponentInst:
		// não expandido: fica no HTML, o unknown-tag check acha depois
		s.lit(reconstructComponent(v))
	case *Slot:
		s.lit("<Slot />")
	}
}

// element emite a tag com atributos (interp em valor de atributo passa pelo
// sink), eventos e bind expandidos, e os filhos
func (em *emitter) element(s sink, el *Element) {
	s.lit("<" + el.Tag)
	for _, a := range el.Attrs {
		em.attr(s, a)
	}
	for _, ev := range el.Events {
		em.hasEvents = true
		s.lit(fmt.Sprintf(` on%s="%s(%s); __pierrotUpdate()"`, ev.Name, ev.Fn, html.EscapeString(ev.Args)))
	}
	if el.HasBind {
		target := strings.TrimSpace(el.Bind)
		id := len(em.binds)
		em.binds = append(em.binds, target)
		s.lit(fmt.Sprintf(` oninput="%s = this.value; __pierrotUpdate()" data-pierrot-bindval="%d"`, html.EscapeString(target), id))
	}
	if el.SelfClose {
		s.lit(" />")
		return
	}
	s.lit(">")
	for _, c := range el.Children {
		em.node(s, c)
	}
	if el.Closed {
		s.lit("</" + el.Tag + ">")
	}
}

// attr re-emite um atributo comum; valor entre aspas pode conter ${...},
// que vai para o sink como interp (mesma semântica do pipeline antigo, que
// rodava os passes de ${} sobre o corpo inteiro, atributos inclusos)
func (em *emitter) attr(s sink, a Attr) {
	switch {
	case !a.HasVal:
		s.lit(" " + a.Name)
	case a.Expr:
		s.lit(" " + a.Name + "={" + a.Val + "}")
	case a.Quoted:
		q := string(a.Quote)
		s.lit(" " + a.Name + "=" + q)
		em.interpolatable(s, a.Val)
		s.lit(q)
	default:
		s.lit(" " + a.Name + "=" + a.Val)
	}
}

// interpolatable manda texto cru para o sink separando os ${expr} válidos
// (mesma regra do lexer: não vazio, sem chaves nem quebra de linha)
func (em *emitter) interpolatable(s sink, str string) {
	for {
		start := strings.Index(str, "${")
		if start < 0 {
			s.lit(str)
			return
		}
		end := -1
		for j := start + 2; j < len(str); j++ {
			if str[j] == '}' {
				end = j
				break
			}
			if str[j] == '{' || str[j] == '\n' {
				break
			}
		}
		if end < 0 || end == start+2 {
			s.lit(str[:start+2])
			str = str[start+2:]
			continue
		}
		s.lit(str[:start])
		s.interp(str[start+2 : end])
		str = str[end+1:]
	}
}

// block emite o placeholder e compila o corpo do bloco para JS
func (em *emitter) block(s sink, n Node) {
	if js, ok := s.(*jsSink); ok {
		// bloco aninhado vira controle de fluxo inline no mesmo corpo
		em.blockFlow(js, n)
		return
	}
	id := len(em.blocks)
	em.blocks = append(em.blocks, "") // reserva a posição na ordem dos placeholders
	var body strings.Builder
	js := &jsSink{js: &body}
	body.WriteString("var __h = \"\";\n")
	em.blockFlow(js, n)
	body.WriteString("return __h;")
	em.blocks[id] = body.String()
	s.lit(fmt.Sprintf("<div data-pierrot-block=\"%d\" style=\"display:contents\"></div>\n", id))
}

// blockFlow emite o for/if como JS, com os filhos concatenando em __h
func (em *emitter) blockFlow(js *jsSink, n Node) {
	js.flush()
	switch v := n.(type) {
	case *ForBlock:
		fmt.Fprintf(js.js, "for (const %s of (%s)) {\n", v.Var, v.Iter)
		for _, c := range v.Children {
			em.node(js, c)
		}
		js.flush()
		js.js.WriteString("}\n")
	case *IfBlock:
		fmt.Fprintf(js.js, "if (%s) {\n", v.Cond)
		for _, c := range v.Then {
			em.node(js, c)
		}
		js.flush()
		if v.HasElse {
			js.js.WriteString("} else {\n")
			for _, c := range v.Else {
				em.node(js, c)
			}
			js.flush()
		}
		js.js.WriteString("}\n")
	}
}

// reconstructComponent devolve a tag de um componente não expandido
func reconstructComponent(c *ComponentInst) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<%s", c.Name)
	for _, a := range c.Props {
		switch {
		case !a.HasVal:
			fmt.Fprintf(&b, " %s", a.Name)
		case a.Expr:
			fmt.Fprintf(&b, " %s={%s}", a.Name, a.Val)
		case a.Quoted:
			fmt.Fprintf(&b, " %s=%c%s%c", a.Name, a.Quote, a.Val, a.Quote)
		default:
			fmt.Fprintf(&b, " %s=%s", a.Name, a.Val)
		}
	}
	b.WriteString(" />")
	return b.String()
}
