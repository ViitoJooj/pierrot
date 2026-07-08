package compiler

import (
	"testing"
)

// kinds resume a sequência de tipos para asserts curtos
func kinds(toks []Token) []TokKind {
	out := make([]TokKind, len(toks))
	for i, t := range toks {
		out[i] = t.Kind
	}
	return out
}

func assertKinds(t *testing.T, toks []Token, want ...TokKind) {
	t.Helper()
	got := kinds(toks)
	if len(got) != len(want) {
		t.Fatalf("tokens: got %d %v, want %d %v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: got kind %d, want %d (all: %v)", i, got[i], want[i], got)
		}
	}
}

func TestLexTextAndTags(t *testing.T) {
	toks := Lex("<p>oi</p>")
	assertKinds(t, toks, TokTagOpen, TokText, TokTagClose)
	if toks[0].Tag.Name != "p" || toks[0].Tag.SelfClose {
		t.Fatalf("tag open: %+v", toks[0].Tag)
	}
	if toks[1].Text != "oi" {
		t.Fatalf("text: %q", toks[1].Text)
	}
	if toks[2].Tag.Name != "p" {
		t.Fatalf("tag close: %+v", toks[2].Tag)
	}
}

func TestLexInterp(t *testing.T) {
	toks := Lex("a ${name} b ${ obj.x } c")
	assertKinds(t, toks, TokText, TokInterp, TokText, TokInterp, TokText)
	if toks[1].Expr != "name" {
		t.Fatalf("interp 1: %q", toks[1].Expr)
	}
	// conteúdo cru, sem trim (o trim é decisão do emitter, como hoje)
	if toks[3].Expr != " obj.x " {
		t.Fatalf("interp 2: %q", toks[3].Expr)
	}
}

func TestLexInterpInvalid(t *testing.T) {
	// ${} vazio, ${ com { dentro e ${ sem fechar na linha são texto
	toks := Lex("${} ${a{b} ${x\n}")
	for _, tok := range toks {
		if tok.Kind == TokInterp {
			t.Fatalf("não devia ter interp: %+v", tok)
		}
	}
}

func TestLexDirectives(t *testing.T) {
	src := "@for item in items\n  <li>${item}</li>\n@end\n  @if count > 1\nmuitos\n@else\num\n@endif\n"
	toks := Lex(src)
	assertKinds(t, toks,
		TokDirFor, TokText, TokTagOpen, TokInterp, TokTagClose, TokText,
		TokDirEnd, TokDirIf, TokText, TokDirElse, TokText, TokDirEnd)
	if toks[0].ForVar != "item" || toks[0].ForIter != "items" {
		t.Fatalf("for: %+v", toks[0])
	}
	if toks[7].Expr != "count > 1" {
		t.Fatalf("if: %q", toks[7].Expr)
	}
}

func TestLexDirectiveNotWholeLine(t *testing.T) {
	// @for no meio de texto não é diretiva
	toks := Lex("texto @for x in y\n")
	assertKinds(t, toks, TokText)
}

func TestLexComment(t *testing.T) {
	// // na coluna 0 morre (linha inteira); indentado é texto
	toks := Lex("// some\n  // fica\ntexto\n")
	assertKinds(t, toks, TokComment, TokText)
	if toks[1].Text != "  // fica\ntexto\n" {
		t.Fatalf("text: %q", toks[1].Text)
	}
}

func TestLexTagAttrs(t *testing.T) {
	toks := Lex(`<a href="/x" data-y={expr} @click={go} @input={set(a, 2)} @bind={code} disabled />`)
	assertKinds(t, toks, TokTagOpen)
	tag := toks[0].Tag
	if !tag.SelfClose || tag.Name != "a" {
		t.Fatalf("tag: %+v", tag)
	}
	if len(tag.Attrs) != 3 {
		t.Fatalf("attrs: %+v", tag.Attrs)
	}
	if tag.Attrs[0].Name != "href" || tag.Attrs[0].Val != "/x" || !tag.Attrs[0].Quoted {
		t.Fatalf("attr href: %+v", tag.Attrs[0])
	}
	if tag.Attrs[1].Name != "data-y" || tag.Attrs[1].Val != "expr" || !tag.Attrs[1].Expr {
		t.Fatalf("attr data-y: %+v", tag.Attrs[1])
	}
	if tag.Attrs[2].Name != "disabled" || tag.Attrs[2].HasVal {
		t.Fatalf("attr disabled: %+v", tag.Attrs[2])
	}
	if len(tag.Events) != 2 {
		t.Fatalf("events: %+v", tag.Events)
	}
	if tag.Events[0].Name != "click" || tag.Events[0].Fn != "go" || tag.Events[0].Args != "" {
		t.Fatalf("event click: %+v", tag.Events[0])
	}
	if tag.Events[1].Name != "input" || tag.Events[1].Fn != "set" || tag.Events[1].Args != "a, 2" {
		t.Fatalf("event input: %+v", tag.Events[1])
	}
	if !tag.HasBind || tag.Bind != "code" {
		t.Fatalf("bind: %+v", tag)
	}
}

func TestLexEventInvalidStaysAttr(t *testing.T) {
	// @click={a.b} não casa a gramática de evento (fn precisa ser \w+): vira attr cru
	toks := Lex(`<b @click={a.b}>`)
	tag := toks[0].Tag
	if len(tag.Events) != 0 || len(tag.Attrs) != 1 || tag.Attrs[0].Name != "@click" {
		t.Fatalf("tag: %+v", tag)
	}
}

func TestLexMultilineTag(t *testing.T) {
	toks := Lex("<Card\n  title=\"oi\"\n  price={p}\n/>")
	assertKinds(t, toks, TokTagOpen)
	tag := toks[0].Tag
	if tag.Name != "Card" || !tag.SelfClose || len(tag.Attrs) != 2 {
		t.Fatalf("tag: %+v", tag)
	}
}

func TestLexLoneLtIsText(t *testing.T) {
	toks := Lex("a < b e 2 <3\n")
	assertKinds(t, toks, TokText)
}

func TestLexInterpInsideAttrValue(t *testing.T) {
	// ${} dentro de valor de atributo fica no valor cru
	toks := Lex(`<a href="${url}">x</a>`)
	tag := toks[0].Tag
	if tag.Attrs[0].Val != "${url}" {
		t.Fatalf("attr: %+v", tag.Attrs[0])
	}
}

func TestLexPositions(t *testing.T) {
	toks := Lex("ab\n<p>${x}")
	// texto em 1:1, tag em 2:1, interp em 2:4
	if toks[0].Pos != (Pos{1, 1}) {
		t.Fatalf("text pos: %+v", toks[0].Pos)
	}
	if toks[1].Pos != (Pos{2, 1}) {
		t.Fatalf("tag pos: %+v", toks[1].Pos)
	}
	if toks[2].Pos != (Pos{2, 4}) {
		t.Fatalf("interp pos: %+v", toks[2].Pos)
	}
}

func TestLexCRLFDirective(t *testing.T) {
	toks := Lex("@if a\r\nx\r\n@end\r\n")
	assertKinds(t, toks, TokDirIf, TokText, TokDirEnd)
	if toks[0].Expr != "a" {
		t.Fatalf("if cond: %q", toks[0].Expr)
	}
}
