package compiler

import "fmt"

// voidElements são tags HTML sem fechamento
var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "link": true, "meta": true,
	"source": true, "track": true, "wbr": true,
}

// frame é um nó aberto na pilha do parser. Exatamente um de el/fb/ib é
// não-nulo. inElse marca em qual ramo do if os filhos caem.
type frame struct {
	el     *Element
	fb     *ForBlock
	ib     *IfBlock
	inElse bool
	kids   []Node
	else_  []Node
}

func (f *frame) add(n Node) {
	if f.ib != nil && f.inElse {
		f.else_ = append(f.else_, n)
		return
	}
	f.kids = append(f.kids, n)
}

// ParseTemplate monta a árvore do template a partir dos tokens do lexer.
// Tolerância igual ao pipeline antigo: tag sem fechamento fica aberta,
// close órfão vira texto, diretiva sem fechamento gera erro e a linha vira
// texto com os filhos soltos no fluxo.
func ParseTemplate(src string) *Template {
	p := &tplParser{}
	for _, tok := range Lex(src) {
		p.feed(tok)
	}
	p.finish()
	return &Template{Children: p.root, Errs: p.errs}
}

type tplParser struct {
	root  []Node
	stack []*frame
	errs  []string
}

func (p *tplParser) add(n Node) {
	if len(p.stack) == 0 {
		p.root = append(p.root, n)
		return
	}
	p.stack[len(p.stack)-1].add(n)
}

func (p *tplParser) push(f *frame) { p.stack = append(p.stack, f) }

func (p *tplParser) pop() *frame {
	f := p.stack[len(p.stack)-1]
	p.stack = p.stack[:len(p.stack)-1]
	return f
}

func (p *tplParser) feed(tok Token) {
	switch tok.Kind {
	case TokText:
		p.add(&Text{Raw: tok.Text, Pos: tok.Pos})
	case TokComment:
		// morre aqui, igual ao commentLineRe
	case TokInterp:
		p.add(&Interp{Expr: tok.Expr, Pos: tok.Pos})
	case TokDirFor:
		p.push(&frame{fb: &ForBlock{Var: tok.ForVar, Iter: tok.ForIter, Pos: tok.Pos}})
	case TokDirIf:
		p.push(&frame{ib: &IfBlock{Cond: tok.Expr, Pos: tok.Pos}})
	case TokDirElse:
		p.dirElse()
	case TokDirEnd:
		p.dirEnd()
	case TokTagOpen:
		p.tagOpen(tok)
	case TokTagClose:
		p.tagClose(tok)
	}
}

func (p *tplParser) tagOpen(tok Token) {
	tag := tok.Tag
	if tag.Name == "Slot" && tag.SelfClose && len(tag.Attrs) == 0 {
		p.add(&Slot{Pos: tok.Pos})
		return
	}
	if tag.Name[0] >= 'A' && tag.Name[0] <= 'Z' && tag.SelfClose {
		p.add(&ComponentInst{Name: tag.Name, Props: tag.Attrs, Pos: tok.Pos})
		return
	}
	el := &Element{
		Tag: tag.Name, Attrs: tag.Attrs, Events: tag.Events,
		Bind: tag.Bind, HasBind: tag.HasBind,
		SelfClose: tag.SelfClose, Void: voidElements[tag.Name], Pos: tok.Pos,
	}
	if el.SelfClose || el.Void {
		p.add(el)
		return
	}
	p.push(&frame{el: el})
}

// tagClose fecha o elemento se ele for o elemento aberto mais próximo sem
// cruzar fronteira de diretiva; senão o </tag> vira texto (tolerância)
func (p *tplParser) tagClose(tok Token) {
	for i := len(p.stack) - 1; i >= 0; i-- {
		f := p.stack[i]
		if f.el == nil {
			break // não cruza @for/@if
		}
		if f.el.Tag == tok.Tag.Name {
			// dobra os elementos abertos acima do alvo (ficam sem close)
			for len(p.stack)-1 > i {
				p.foldTop()
			}
			f = p.pop()
			f.el.Children = f.kids
			f.el.Closed = true
			p.add(f.el)
			return
		}
	}
	p.add(&Text{Raw: "</" + tok.Tag.Name + ">", Pos: tok.Pos})
}

// foldTop fecha implicitamente o frame do topo sem marcar Closed
func (p *tplParser) foldTop() {
	f := p.pop()
	switch {
	case f.el != nil:
		f.el.Children = f.kids
		p.add(f.el)
	case f.fb != nil:
		f.fb.Children = f.kids
		p.add(f.fb)
	case f.ib != nil:
		f.ib.Then, f.ib.Else, f.ib.HasElse = f.kids, f.else_, f.inElse
		p.add(f.ib)
	}
}

// dirElse liga o ramo else do if aberto mais próximo, dobrando elementos
// abertos no meio; órfão vira erro e o token morre (igual hoje)
func (p *tplParser) dirElse() {
	for i := len(p.stack) - 1; i >= 0; i-- {
		f := p.stack[i]
		if f.el != nil {
			continue
		}
		if f.ib != nil && !f.inElse {
			for len(p.stack)-1 > i {
				p.foldTop()
			}
			f.inElse = true
			return
		}
		break
	}
	p.errs = append(p.errs, `"@else" sem @for/@if correspondente`)
}

// dirEnd fecha o bloco de diretiva aberto mais próximo, dobrando elementos
// abertos no meio; órfão vira erro e o token morre
func (p *tplParser) dirEnd() {
	for i := len(p.stack) - 1; i >= 0; i-- {
		f := p.stack[i]
		if f.el != nil {
			continue
		}
		for len(p.stack)-1 > i {
			p.foldTop()
		}
		f = p.pop()
		if f.fb != nil {
			f.fb.Children = f.kids
			p.add(f.fb)
		} else {
			f.ib.Then, f.ib.Else, f.ib.HasElse = f.kids, f.else_, f.inElse
			p.add(f.ib)
		}
		return
	}
	p.errs = append(p.errs, `"@end" sem @for/@if correspondente`)
}

// finish trata o fim do fonte: elemento aberto dobra em silêncio; diretiva
// aberta vira erro, a linha reconstruída vira texto e os filhos continuam
// no fluxo (igual ao compileTemplate antigo)
func (p *tplParser) finish() {
	for len(p.stack) > 0 {
		f := p.stack[len(p.stack)-1]
		if f.el != nil {
			p.foldTop()
			continue
		}
		f = p.pop()
		var line string
		var kids []Node
		if f.fb != nil {
			line = fmt.Sprintf("@for %s in %s", f.fb.Var, f.fb.Iter)
			kids = f.kids
		} else {
			line = "@if " + f.ib.Cond
			kids = append(f.kids, f.else_...)
		}
		p.errs = append(p.errs, fmt.Sprintf("diretiva %q sem @end/@endif", line))
		p.add(&Text{Raw: line + "\n", Pos: Pos{}})
		for _, n := range kids {
			p.add(n)
		}
	}
}
