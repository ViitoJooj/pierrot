package compiler

// Node é um nó da árvore do template
type Node interface {
	NodePos() Pos
}

// Text é HTML/texto cru, passa direto para o output
type Text struct {
	Raw string
	Pos Pos
}

// Interp é uma interpolação ${expr} (Expr cru, sem trim)
type Interp struct {
	Expr string
	Pos  Pos
}

// Element é uma tag HTML comum. Closed indica se o fechamento </tag> deve
// ser emitido (tags não fechadas no fonte são toleradas e emitidas abertas)
type Element struct {
	Tag       string
	Attrs     []Attr
	Events    []Event
	Bind      string
	HasBind   bool
	SelfClose bool
	Closed    bool
	Void      bool
	Children  []Node
	Pos       Pos
}

// ComponentInst é uma instância de componente: <Nome prop="x" outra={y} />
// ponytail: @evento/@bind em tag de componente são descartados, igual hoje
// ninguém depende deles chegarem no filho
type ComponentInst struct {
	Name  string
	Props []Attr
	Pos   Pos
}

// Slot é a tag <Slot /> do layout, substituída pela página
type Slot struct {
	Pos Pos
}

// ForBlock é @for VAR in ITER ... @end
type ForBlock struct {
	Var, Iter string
	Children  []Node
	Pos       Pos
}

// IfBlock é @if COND ... [@else ...] @end
type IfBlock struct {
	Cond    string
	Then    []Node
	Else    []Node
	HasElse bool
	Pos     Pos
}

func (n *Text) NodePos() Pos          { return n.Pos }
func (n *Interp) NodePos() Pos        { return n.Pos }
func (n *Element) NodePos() Pos       { return n.Pos }
func (n *ComponentInst) NodePos() Pos { return n.Pos }
func (n *Slot) NodePos() Pos          { return n.Pos }
func (n *ForBlock) NodePos() Pos      { return n.Pos }
func (n *IfBlock) NodePos() Pos       { return n.Pos }

// Template é o resultado do parse: a floresta de nós + erros não fatais
// (mesma tolerância do pipeline antigo)
type Template struct {
	Children []Node
	Errs     []string
}
