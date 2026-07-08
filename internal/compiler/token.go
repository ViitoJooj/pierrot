// Package compiler contém o lexer, parser, AST e emitter da linguagem
// .pierrot — a substituição das regexes por parsing de verdade.
package compiler

// Pos é a posição de um token/nó no fonte, 1-based
type Pos struct {
	Line, Col int
}

type TokKind int

const (
	TokText TokKind = iota
	TokTagOpen
	TokTagClose
	TokInterp
	TokDirFor
	TokDirIf
	TokDirElse
	TokDirEnd
	TokComment
)

// Attr é um atributo comum de tag: bare (HasVal=false), nome="valor"
// (Quoted), nome={expr} (Expr) ou nome=valor sem aspas
type Attr struct {
	Name   string
	Val    string
	HasVal bool
	Quoted bool
	Quote  byte // ' ou " quando Quoted
	Expr   bool
	Pos    Pos
}

// Event é @evento={fn} ou @evento={fn(args)} num atributo de tag
type Event struct {
	Name, Fn, Args string
	Pos            Pos
}

// Tag é uma tag aberta/fechada com os atributos já classificados:
// @bind={alvo} vira Bind, @evento={fn(...)} vira Event, o resto fica em Attrs
type Tag struct {
	Name      string
	Attrs     []Attr
	Events    []Event
	Bind      string
	HasBind   bool
	SelfClose bool
	Pos       Pos
}

// Token é a unidade emitida pelo lexer. Campos usados por Kind:
// TokText/TokComment: Text; TokTagOpen/TokTagClose: Tag;
// TokInterp: Expr (conteúdo cru de ${...}); TokDirFor: ForVar/ForIter;
// TokDirIf: Expr (condição)
type Token struct {
	Kind    TokKind
	Pos     Pos
	Text    string
	Tag     *Tag
	Expr    string
	ForVar  string
	ForIter string
}
