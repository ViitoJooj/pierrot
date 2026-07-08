package compiler

import (
	"strings"
	"testing"
)

func card(t *testing.T, childSrc string, declared []string, instSrc string) []Node {
	t.Helper()
	child := ParseTemplate(childSrc)
	inst := ParseTemplate(instSrc).Children[0].(*ComponentInst)
	return ExpandComponent(child, declared, inst)
}

// emitNodes renderiza os nós expandidos para conferir o resultado final
func emitNodes(nodes []Node) *Emitted {
	return Emit(&Template{Children: nodes})
}

func TestSubstIdents(t *testing.T) {
	props := map[string]PropVal{
		"title": {Text: "t", Expr: `"t"`, Literal: true},
		"id":    {Expr: "(p.id)"},
	}
	cases := []struct {
		in, want string
		changed  bool
	}{
		{"buy(id)", "buy((p.id))", true},
		{"title", `"t"`, true},
		{"card.title", "card.title", false}, // precedido de . fica
		{"titleX", "titleX", false},         // boundary de palavra
		{"id + id", "(p.id) + (p.id)", true},
		{"$title", "$title", false}, // $ é char de ident: $title != title
	}
	for _, c := range cases {
		got, changed := SubstIdents(c.in, props)
		if got != c.want || changed != c.changed {
			t.Errorf("SubstIdents(%q) = %q,%v; want %q,%v", c.in, got, changed, c.want, c.changed)
		}
	}
}

func TestExpandLiteralPropInText(t *testing.T) {
	nodes := card(t, "<p>{title}</p>", []string{"title"}, `<Card title="Oi & tal" />`)
	e := emitNodes(nodes)
	if e.HTML != "<p>Oi &amp; tal</p>" {
		t.Fatalf("html: %q", e.HTML)
	}
}

func TestExpandExprPropInText(t *testing.T) {
	nodes := card(t, "<p>{title}</p>", []string{"title"}, `<Card title={x.y} />`)
	e := emitNodes(nodes)
	// vira interp composta -> span de expr com a expressão da instância
	if !strings.Contains(e.HTML, `data-pierrot-expr="0"`) || e.Exprs[0] != "(x.y)" {
		t.Fatalf("html: %q exprs: %v", e.HTML, e.Exprs)
	}
}

func TestExpandMissingPropBecomesEmpty(t *testing.T) {
	nodes := card(t, "<p>{title}</p>", []string{"title"}, `<Card />`)
	e := emitNodes(nodes)
	if e.HTML != "<p></p>" {
		t.Fatalf("html: %q", e.HTML)
	}
}

func TestExpandInterpProp(t *testing.T) {
	// ${title} (interp de verdade) também recebe a prop
	nodes := card(t, "<p>${title}</p>", []string{"title"}, `<Card title="Oi" />`)
	e := emitNodes(nodes)
	if e.HTML != "<p>Oi</p>" {
		t.Fatalf("html: %q", e.HTML)
	}
}

func TestExpandAttrExprLiteral(t *testing.T) {
	nodes := card(t, `<img src={image} />`, []string{"image"}, `<Card image="a.png" />`)
	e := emitNodes(nodes)
	if e.HTML != `<img src="a.png" />` {
		t.Fatalf("html: %q", e.HTML)
	}
}

func TestExpandAttrExprNonLiteral(t *testing.T) {
	nodes := card(t, `<img src={image} />`, []string{"image"}, `<Card image={img} />`)
	e := emitNodes(nodes)
	// atributo vira "${(img)}", que fora de bloco vira span (limitação preservada)
	if !strings.Contains(e.HTML, `src="`) || len(e.Exprs) != 1 || e.Exprs[0] != "(img)" {
		t.Fatalf("html: %q exprs: %v", e.HTML, e.Exprs)
	}
}

func TestExpandAttrExprUntouched(t *testing.T) {
	// attr {expr} sem prop referenciada fica como está
	nodes := card(t, `<div data-x={other}>a</div>`, []string{"title"}, `<Card title="t" />`)
	e := emitNodes(nodes)
	if !strings.Contains(e.HTML, "data-x={other}") {
		t.Fatalf("html: %q", e.HTML)
	}
}

func TestExpandEventArgs(t *testing.T) {
	nodes := card(t, "<button @click={buy(id)}>c</button>", []string{"id"}, `<Card id={p.id} />`)
	e := emitNodes(nodes)
	if !strings.Contains(e.HTML, `onclick="buy((p.id)); __pierrotUpdate()"`) {
		t.Fatalf("html: %q", e.HTML)
	}
}

func TestExpandBindTarget(t *testing.T) {
	nodes := card(t, "<input @bind={code} />", []string{"code"}, `<Card code={c} />`)
	e := emitNodes(nodes)
	if len(e.Binds) != 1 || e.Binds[0] != "(c)" {
		t.Fatalf("binds: %v html: %q", e.Binds, e.HTML)
	}
}

func TestExpandInsideBlock(t *testing.T) {
	nodes := card(t, "@for x in xs\n<p>{title}</p>\n@end\n", []string{"title"}, `<Card title="T" />`)
	e := emitNodes(nodes)
	if len(e.Blocks) != 1 || !strings.Contains(e.Blocks[0], "T") {
		t.Fatalf("blocks: %v", e.Blocks)
	}
}

func TestExpandEqGuard(t *testing.T) {
	// {x} precedido de = em texto cru não é interpolação de prop (guarda do
	// spanInterpRe antigo)
	nodes := card(t, "<p>a={title} b</p>", []string{"title"}, `<Card title="T" />`)
	e := emitNodes(nodes)
	if !strings.Contains(e.HTML, "a={title}") {
		t.Fatalf("html: %q", e.HTML)
	}
}

func TestExpandNoDeclared(t *testing.T) {
	nodes := card(t, "<p>{x}</p>", nil, `<Card x="1" />`)
	e := emitNodes(nodes)
	if e.HTML != "<p>{x}</p>" {
		t.Fatalf("html: %q", e.HTML)
	}
}

func TestExpandQuotedAttrWithBareProp(t *testing.T) {
	// {prop} dentro de valor de atributo com aspas também expande
	nodes := card(t, `<div class="card {kind}">a</div>`, []string{"kind"}, `<Card kind="big" />`)
	e := emitNodes(nodes)
	if !strings.Contains(e.HTML, `class="card big"`) {
		t.Fatalf("html: %q", e.HTML)
	}
}

func TestReplaceSlot(t *testing.T) {
	layout := ParseTemplate("<main><Slot /></main>")
	page := ParseTemplate("<p>x</p>")
	out := SpliceSlots(layout.Children, page.Children)
	e := Emit(&Template{Children: out})
	if e.HTML != "<main><p>x</p></main>" {
		t.Fatalf("html: %q", e.HTML)
	}
}
