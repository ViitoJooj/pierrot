package compiler

import "strings"

// Lex tokeniza o template .pierrot. Gramática igual à das regexes antigas:
// comentário // só na coluna 0 (linha inteira morre); diretivas @for/@if/
// @else/@end ocupam a linha inteira (indentação permitida); ${expr} sem
// chaves nem quebra de linha dentro; tags com atributos podem atravessar
// linhas; < que não abre tag é texto.
func Lex(src string) []Token {
	lx := &lexer{src: src, line: 1, col: 1, textStart: -1}
	for lx.i < len(lx.src) {
		if lx.col == 1 && lx.tryLineToken() {
			continue
		}
		c := lx.src[lx.i]
		if c == '$' && lx.i+1 < len(lx.src) && lx.src[lx.i+1] == '{' && lx.tryInterp() {
			continue
		}
		if c == '<' && lx.tryTag() {
			continue
		}
		if lx.textStart < 0 {
			lx.textStart = lx.i
			lx.textPos = lx.pos()
		}
		lx.advance(1)
	}
	lx.flushText()
	return lx.toks
}

type lexer struct {
	src       string
	i         int
	line, col int
	toks      []Token
	textStart int
	textPos   Pos
}

func (lx *lexer) pos() Pos { return Pos{lx.line, lx.col} }

// advance move n bytes atualizando linha/coluna
func (lx *lexer) advance(n int) {
	for ; n > 0 && lx.i < len(lx.src); n-- {
		if lx.src[lx.i] == '\n' {
			lx.line++
			lx.col = 1
		} else {
			lx.col++
		}
		lx.i++
	}
}

func (lx *lexer) flushText() {
	if lx.textStart < 0 {
		return
	}
	lx.toks = append(lx.toks, Token{Kind: TokText, Pos: lx.textPos, Text: lx.src[lx.textStart:lx.i]})
	lx.textStart = -1
}

func (lx *lexer) emit(t Token) {
	lx.flushText()
	lx.toks = append(lx.toks, t)
}

// tryLineToken testa comentário/diretiva na linha atual inteira (só chamado
// na coluna 1). Consumindo a linha + quebra quando casa.
func (lx *lexer) tryLineToken() bool {
	eol := strings.IndexByte(lx.src[lx.i:], '\n')
	var line string
	var lineLen int
	if eol < 0 {
		line = lx.src[lx.i:]
		lineLen = len(line)
	} else {
		line = lx.src[lx.i : lx.i+eol]
		lineLen = eol + 1 // consome o \n junto
	}

	pos := lx.pos()
	if strings.HasPrefix(line, "//") {
		lx.emit(Token{Kind: TokComment, Pos: pos, Text: line})
		lx.advance(lineLen)
		return true
	}

	trimmed := strings.TrimSpace(line)
	switch {
	case trimmed == "@else":
		lx.emit(Token{Kind: TokDirElse, Pos: pos})
	case trimmed == "@end" || trimmed == "@endif":
		lx.emit(Token{Kind: TokDirEnd, Pos: pos})
	default:
		if v, iter, ok := parseForLine(trimmed); ok {
			lx.emit(Token{Kind: TokDirFor, Pos: pos, ForVar: v, ForIter: iter})
			break
		}
		if cond, ok := parseIfLine(trimmed); ok {
			lx.emit(Token{Kind: TokDirIf, Pos: pos, Expr: cond})
			break
		}
		return false
	}
	lx.advance(lineLen)
	return true
}

// parseForLine casa "@for VAR in EXPR" (VAR = \w+, EXPR não vazio)
func parseForLine(s string) (v, iter string, ok bool) {
	rest, found := cutKeyword(s, "@for")
	if !found {
		return "", "", false
	}
	j := 0
	for j < len(rest) && isWordChar(rest[j]) {
		j++
	}
	if j == 0 || j == len(rest) || !isSpaceByte(rest[j]) {
		return "", "", false
	}
	v = rest[:j]
	iter, found = cutKeyword(strings.TrimSpace(rest[j:]), "in")
	if !found || iter == "" {
		return "", "", false
	}
	return v, iter, true
}

// parseIfLine casa "@if COND"
func parseIfLine(s string) (string, bool) {
	cond, found := cutKeyword(s, "@if")
	if !found || cond == "" {
		return "", false
	}
	return cond, true
}

// cutKeyword corta o prefixo kw seguido de espaço e devolve o resto trimado
func cutKeyword(s, kw string) (string, bool) {
	rest, ok := strings.CutPrefix(s, kw)
	if !ok || rest == "" || !isSpaceByte(rest[0]) {
		return "", false
	}
	return strings.TrimSpace(rest), true
}

func isSpaceByte(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

func isWordChar(c byte) bool {
	return c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isAlpha(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// tryInterp casa ${expr} com expr não vazia, sem {, } nem \n dentro
func (lx *lexer) tryInterp() bool {
	rest := lx.src[lx.i+2:]
	for j := 0; j < len(rest); j++ {
		switch rest[j] {
		case '}':
			if j == 0 {
				return false
			}
			lx.emit(Token{Kind: TokInterp, Pos: lx.pos(), Expr: rest[:j]})
			lx.advance(2 + j + 1)
			return true
		case '{', '\n':
			return false
		}
	}
	return false
}

// tryTag casa <name ...>, <name ... /> ou </name>. Não casando (sem nome
// válido ou sem > até o fim do arquivo), devolve false e o < vira texto.
func (lx *lexer) tryTag() bool {
	pos := lx.pos()
	j := lx.i + 1
	closing := false
	if j < len(lx.src) && lx.src[j] == '/' {
		closing = true
		j++
	}
	if j >= len(lx.src) || !isAlpha(lx.src[j]) {
		return false
	}
	nameStart := j
	for j < len(lx.src) && (isWordChar(lx.src[j]) || lx.src[j] == '-') {
		j++
	}
	tag := &Tag{Name: lx.src[nameStart:j], Pos: pos}

	if closing {
		for j < len(lx.src) && isSpaceByte(lx.src[j]) {
			j++
		}
		if j >= len(lx.src) || lx.src[j] != '>' {
			return false
		}
		lx.emit(Token{Kind: TokTagClose, Pos: pos, Tag: tag})
		lx.advance(j + 1 - lx.i)
		return true
	}

	end, ok := lx.scanAttrs(tag, j)
	if !ok {
		return false
	}
	lx.emit(Token{Kind: TokTagOpen, Pos: pos, Tag: tag})
	lx.advance(end - lx.i)
	return true
}

// scanAttrs lê os atributos a partir de j até > ou />, classificando
// @bind={...} e @evento={fn(args)}. Devolve a posição logo após o > .
func (lx *lexer) scanAttrs(tag *Tag, j int) (int, bool) {
	src := lx.src
	for {
		for j < len(src) && isSpaceByte(src[j]) {
			j++
		}
		if j >= len(src) {
			return 0, false
		}
		switch src[j] {
		case '>':
			return j + 1, true
		case '/':
			if j+1 < len(src) && src[j+1] == '>' {
				tag.SelfClose = true
				return j + 2, true
			}
			j++ // / solto dentro da tag: ignora, igual um parser tolerante
			continue
		}

		nameStart := j
		for j < len(src) && !isSpaceByte(src[j]) && src[j] != '=' && src[j] != '>' && src[j] != '/' {
			j++
		}
		if j == nameStart {
			return 0, false
		}
		name := src[nameStart:j]

		if j >= len(src) || src[j] != '=' {
			tag.Attrs = append(tag.Attrs, Attr{Name: name})
			continue
		}
		j++ // =
		if j >= len(src) {
			return 0, false
		}
		switch src[j] {
		case '"', '\'':
			quote := src[j]
			k := strings.IndexByte(src[j+1:], quote)
			if k < 0 {
				return 0, false
			}
			tag.Attrs = append(tag.Attrs, Attr{Name: name, Val: src[j+1 : j+1+k], HasVal: true, Quoted: true, Quote: quote})
			j += k + 2
		case '{':
			k := j + 1
			for k < len(src) && src[k] != '}' && src[k] != '{' && src[k] != '\n' {
				k++
			}
			if k >= len(src) || src[k] != '}' {
				return 0, false
			}
			val := src[j+1 : k]
			j = k + 1
			lx.classifyBraced(tag, name, val)
		default:
			k := j
			for k < len(src) && !isSpaceByte(src[k]) && src[k] != '>' {
				k++
			}
			tag.Attrs = append(tag.Attrs, Attr{Name: name, Val: src[j:k], HasVal: true})
			j = k
		}
	}
}

// classifyBraced decide o destino de name={val}: @bind vira Bind, @evento
// com corpo fn ou fn(args) vira Event, o resto (inclusive @x que não casa a
// gramática de evento) fica em Attrs como expressão
func (lx *lexer) classifyBraced(tag *Tag, name, val string) {
	if name == "@bind" && val != "" {
		tag.Bind = val
		tag.HasBind = true
		return
	}
	if strings.HasPrefix(name, "@") && len(name) > 1 && isWordName(name[1:]) {
		if fn, args, ok := parseEventBody(val); ok {
			tag.Events = append(tag.Events, Event{Name: name[1:], Fn: fn, Args: args})
			return
		}
	}
	tag.Attrs = append(tag.Attrs, Attr{Name: name, Val: val, HasVal: true, Expr: true})
}

// parseEventBody casa "fn" ou "fn(args)" (fn = \w+, args sem parênteses)
func parseEventBody(s string) (fn, args string, ok bool) {
	j := 0
	for j < len(s) && isWordChar(s[j]) {
		j++
	}
	if j == 0 {
		return "", "", false
	}
	fn = s[:j]
	if j == len(s) {
		return fn, "", true
	}
	if s[j] != '(' || s[len(s)-1] != ')' {
		return "", "", false
	}
	args = s[j+1 : len(s)-1]
	if strings.ContainsAny(args, "()") {
		return "", "", false
	}
	return fn, args, true
}

func isWordName(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isWordChar(s[i]) {
			return false
		}
	}
	return s != ""
}
