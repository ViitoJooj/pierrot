package compiler

import (
	"strings"
	"testing"
)

func emit(t *testing.T, src string) *Emitted {
	t.Helper()
	tpl := ParseTemplate(src)
	if len(tpl.Errs) != 0 {
		t.Fatalf("parse errs: %v", tpl.Errs)
	}
	return Emit(tpl)
}

func TestEmitBindingSpan(t *testing.T) {
	e := emit(t, "<h1>${name}</h1>")
	if e.HTML != `<h1><span data-bind="name"></span></h1>` {
		t.Fatalf("html: %q", e.HTML)
	}
	if len(e.BindNames) != 1 || e.BindNames[0] != "name" {
		t.Fatalf("bindNames: %v", e.BindNames)
	}
}

func TestEmitBindNamesDedup(t *testing.T) {
	e := emit(t, "${a} ${b} ${a}")
	if strings.Join(e.BindNames, ",") != "a,b" {
		t.Fatalf("bindNames: %v", e.BindNames)
	}
}

func TestEmitExprSpan(t *testing.T) {
	e := emit(t, "<p>${obj.x}</p>")
	if e.HTML != `<p><span data-pierrot-expr="0"></span></p>` {
		t.Fatalf("html: %q", e.HTML)
	}
	if len(e.Exprs) != 1 || e.Exprs[0] != "obj.x" {
		t.Fatalf("exprs: %v", e.Exprs)
	}
}

func TestEmitEvent(t *testing.T) {
	e := emit(t, "<button @click={inc}>+</button>")
	if e.HTML != `<button onclick="inc(); __pierrotUpdate()">+</button>` {
		t.Fatalf("html: %q", e.HTML)
	}
}

func TestEmitEventArgsEscaped(t *testing.T) {
	e := emit(t, `<button @click={add(2, "x")}>+</button>`)
	want := `<button onclick="add(2, &#34;x&#34;); __pierrotUpdate()">+</button>`
	if e.HTML != want {
		t.Fatalf("html: %q", e.HTML)
	}
}

func TestEmitBind(t *testing.T) {
	e := emit(t, `<input @bind={ code } />`)
	want := `<input oninput="code = this.value; __pierrotUpdate()" data-pierrot-bindval="0" />`
	if e.HTML != want {
		t.Fatalf("html: %q", e.HTML)
	}
	if len(e.Binds) != 1 || e.Binds[0] != "code" {
		t.Fatalf("binds: %v", e.Binds)
	}
}

func TestEmitForBlock(t *testing.T) {
	e := emit(t, "@for item in items\n<li>${item}</li>\n@end\n")
	if !strings.Contains(e.HTML, `<div data-pierrot-block="0" style="display:contents"></div>`) {
		t.Fatalf("html: %q", e.HTML)
	}
	if len(e.Blocks) != 1 {
		t.Fatalf("blocks: %d", len(e.Blocks))
	}
	js := e.Blocks[0]
	for _, want := range []string{
		`var __h = "";`,
		`for (const item of (items)) {`,
		`__pierrotEsc(typeof item === "undefined" ? "" : item)`,
		`return __h;`,
	} {
		if !strings.Contains(js, want) {
			t.Fatalf("block js sem %q:\n%s", want, js)
		}
	}
}

func TestEmitIfElseBlock(t *testing.T) {
	e := emit(t, "@if n > 1\nmuitos\n@else\num\n@endif\n")
	js := e.Blocks[0]
	for _, want := range []string{"if (n > 1) {", "} else {", `"muitos\n"`, `"um\n"`} {
		if !strings.Contains(js, want) {
			t.Fatalf("block js sem %q:\n%s", want, js)
		}
	}
}

func TestEmitNestedBlockInline(t *testing.T) {
	// bloco aninhado vira controle de fluxo inline, não placeholder próprio
	e := emit(t, "@for a in xs\n@if a\n${a}\n@end\n@end\n")
	if len(e.Blocks) != 1 {
		t.Fatalf("blocks: %d", len(e.Blocks))
	}
	if !strings.Contains(e.Blocks[0], "if (a) {") {
		t.Fatalf("js: %s", e.Blocks[0])
	}
}

func TestEmitEventInsideBlock(t *testing.T) {
	e := emit(t, "@for p in ps\n<button @click={del(p)}>x</button>\n@end\n")
	if !strings.Contains(e.Blocks[0], `onclick=`) || !strings.Contains(e.Blocks[0], "__pierrotUpdate()") {
		t.Fatalf("js: %s", e.Blocks[0])
	}
}

func TestEmitInterpInAttrInsideBlock(t *testing.T) {
	e := emit(t, "@for p in ps\n<a href=\"${p.url}\">l</a>\n@end\n")
	if !strings.Contains(e.Blocks[0], `__pierrotEsc(p.url)`) {
		t.Fatalf("js: %s", e.Blocks[0])
	}
}

func TestEmitAttrInterpOutsideBlock(t *testing.T) {
	// limitação preservada do pipeline antigo: ${} em atributo fora de bloco
	// vira span dentro do valor
	e := emit(t, `<a href="${url}">l</a>`)
	if !strings.Contains(e.HTML, `href="<span data-bind="url"></span>"`) {
		t.Fatalf("html: %q", e.HTML)
	}
	if len(e.BindNames) != 1 || e.BindNames[0] != "url" {
		t.Fatalf("bindNames: %v", e.BindNames)
	}
}

func TestEmitTwoBlocksNumbering(t *testing.T) {
	e := emit(t, "@if a\nx\n@end\n@if b\ny\n@end\n")
	if !strings.Contains(e.HTML, `data-pierrot-block="0"`) || !strings.Contains(e.HTML, `data-pierrot-block="1"`) {
		t.Fatalf("html: %q", e.HTML)
	}
	if len(e.Blocks) != 2 {
		t.Fatalf("blocks: %d", len(e.Blocks))
	}
}

func TestEmitPassthrough(t *testing.T) {
	// HTML comum sai como entrou (atributos re-emitidos)
	e := emit(t, "<div class=\"a\" id='b' data-x={y} hidden>t</div>\n<br />")
	for _, want := range []string{`class="a"`, `id='b'`, `data-x={y}`, `hidden`, `<br />`} {
		if !strings.Contains(e.HTML, want) {
			t.Fatalf("html sem %q: %q", want, e.HTML)
		}
	}
}

func TestNeedsRuntime(t *testing.T) {
	cases := []struct {
		src  string
		want bool
	}{
		{"<h1>Hello</h1>", false},
		{"<h1>${name}</h1>", true},
		{"@if a\nx\n@end\n", true},
		{"<p>${obj.x}</p>", true},
		{"<input @bind={c} />", true},
		{"<button @click={inc}>+</button>", true},
	}
	for _, c := range cases {
		if got := emit(t, c.src).NeedsRuntime(); got != c.want {
			t.Errorf("NeedsRuntime(%q) = %v, want %v", c.src, got, c.want)
		}
	}
}

func TestEmitComponentLeftoverStays(t *testing.T) {
	// componente não expandido fica no HTML (o unknown-tag check acha depois)
	e := emit(t, `<Card title="x" />`)
	if !strings.Contains(e.HTML, "<Card") {
		t.Fatalf("html: %q", e.HTML)
	}
}
