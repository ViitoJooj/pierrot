package compiler

import (
	"strings"
	"testing"
)

func TestParseSimpleTree(t *testing.T) {
	tpl := ParseTemplate("<div class=\"a\">oi ${name}</div>")
	if len(tpl.Errs) != 0 {
		t.Fatalf("errs: %v", tpl.Errs)
	}
	if len(tpl.Children) != 1 {
		t.Fatalf("children: %d", len(tpl.Children))
	}
	el, ok := tpl.Children[0].(*Element)
	if !ok || el.Tag != "div" || !el.Closed {
		t.Fatalf("el: %+v", tpl.Children[0])
	}
	if len(el.Children) != 2 {
		t.Fatalf("el children: %d", len(el.Children))
	}
	if _, ok := el.Children[0].(*Text); !ok {
		t.Fatalf("child 0: %T", el.Children[0])
	}
	if in, ok := el.Children[1].(*Interp); !ok || in.Expr != "name" {
		t.Fatalf("child 1: %+v", el.Children[1])
	}
}

func TestParseBlocks(t *testing.T) {
	src := "@for item in items\n<li>${item}</li>\n@end\n@if n > 1\nmuitos\n@else\num\n@endif\n"
	tpl := ParseTemplate(src)
	if len(tpl.Errs) != 0 {
		t.Fatalf("errs: %v", tpl.Errs)
	}
	var fb *ForBlock
	var ib *IfBlock
	for _, n := range tpl.Children {
		switch v := n.(type) {
		case *ForBlock:
			fb = v
		case *IfBlock:
			ib = v
		}
	}
	if fb == nil || fb.Var != "item" || fb.Iter != "items" {
		t.Fatalf("for: %+v", fb)
	}
	if len(fb.Children) == 0 {
		t.Fatal("for sem filhos")
	}
	if ib == nil || ib.Cond != "n > 1" || !ib.HasElse {
		t.Fatalf("if: %+v", ib)
	}
	if len(ib.Then) == 0 || len(ib.Else) == 0 {
		t.Fatalf("if branches: then=%d else=%d", len(ib.Then), len(ib.Else))
	}
}

func TestParseNestedBlocks(t *testing.T) {
	src := "@for a in xs\n@if a\n${a}\n@end\n@end\n"
	tpl := ParseTemplate(src)
	if len(tpl.Errs) != 0 {
		t.Fatalf("errs: %v", tpl.Errs)
	}
	fb, ok := tpl.Children[0].(*ForBlock)
	if !ok {
		t.Fatalf("child 0: %T", tpl.Children[0])
	}
	found := false
	for _, n := range fb.Children {
		if _, ok := n.(*IfBlock); ok {
			found = true
		}
	}
	if !found {
		t.Fatal("if aninhado não achado dentro do for")
	}
}

func TestParseUnclosedDirective(t *testing.T) {
	tpl := ParseTemplate("@for x in xs\n<li>a</li>\n")
	if len(tpl.Errs) != 1 || !strings.Contains(tpl.Errs[0], "sem @end/@endif") {
		t.Fatalf("errs: %v", tpl.Errs)
	}
	// a linha da diretiva vira texto e os filhos continuam no fluxo (igual hoje)
	if _, ok := tpl.Children[0].(*Text); !ok {
		t.Fatalf("child 0: %T", tpl.Children[0])
	}
}

func TestParseOrphanElseEnd(t *testing.T) {
	tpl := ParseTemplate("@else\n@end\n")
	if len(tpl.Errs) != 2 {
		t.Fatalf("errs: %v", tpl.Errs)
	}
	for _, e := range tpl.Errs {
		if !strings.Contains(e, "sem @for/@if correspondente") {
			t.Fatalf("err: %q", e)
		}
	}
}

func TestParseVoidElement(t *testing.T) {
	tpl := ParseTemplate(`<img src="x.png"><p>a</p>`)
	if len(tpl.Children) != 2 {
		t.Fatalf("children: %d (%+v)", len(tpl.Children), tpl.Children)
	}
	img := tpl.Children[0].(*Element)
	if !img.Void || len(img.Children) != 0 {
		t.Fatalf("img: %+v", img)
	}
}

func TestParseComponentAndSlot(t *testing.T) {
	tpl := ParseTemplate(`<Card title="oi" price={p} /><Slot />`)
	ci, ok := tpl.Children[0].(*ComponentInst)
	if !ok || ci.Name != "Card" || len(ci.Props) != 2 {
		t.Fatalf("comp: %+v", tpl.Children[0])
	}
	if _, ok := tpl.Children[1].(*Slot); !ok {
		t.Fatalf("slot: %T", tpl.Children[1])
	}
}

func TestParseStrayCloseTag(t *testing.T) {
	tpl := ParseTemplate("a</div>b")
	// close sem open vira texto, sem erro (tolerância do pipeline antigo)
	if len(tpl.Errs) != 0 {
		t.Fatalf("errs: %v", tpl.Errs)
	}
	joined := ""
	for _, n := range tpl.Children {
		if txt, ok := n.(*Text); ok {
			joined += txt.Raw
		}
	}
	if joined != "a</div>b" {
		t.Fatalf("texto: %q", joined)
	}
}

func TestParseBlockCrossingElement(t *testing.T) {
	// bloco fechando com elemento aberto dentro: elemento dobra (fica sem
	// close estrutural) e o </div> posterior vira texto — ordem preservada
	src := "@if x\n<div>\n@end\n</div>\n"
	tpl := ParseTemplate(src)
	ib, ok := tpl.Children[0].(*IfBlock)
	if !ok {
		t.Fatalf("child 0: %T", tpl.Children[0])
	}
	el := ib.Then[0].(*Element)
	if el.Closed {
		t.Fatal("div não devia estar Closed")
	}
}

func TestParseUppercaseNotSelfClose(t *testing.T) {
	// componente sem self-close não é instância (igual tagRe antigo): vira Element
	tpl := ParseTemplate("<Card>x</Card>")
	el, ok := tpl.Children[0].(*Element)
	if !ok || el.Tag != "Card" {
		t.Fatalf("child 0: %+v", tpl.Children[0])
	}
}
