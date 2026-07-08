package compiler

import "strings"

// Scanners de código JS/TS que substituem as regexes do dev server:
// get.Dotenv("X") resolvido em build e time.Sleep com await automático.

// ReplaceDotenvCalls troca cada get.Dotenv("NOME") literal por resolve(NOME)
// e informa se sobrou alguma chamada get.Dotenv( não literal (erro textual
// no caller, já que nome dinâmico não dá para resolver em build)
func ReplaceDotenvCalls(code string, resolve func(name string) string) (string, bool) {
	const call = "get.Dotenv"
	var b strings.Builder
	for {
		idx := strings.Index(code, call)
		if idx < 0 {
			break
		}
		b.WriteString(code[:idx])
		rest := code[idx+len(call):]
		if name, after, ok := scanDotenvArgs(rest); ok {
			b.WriteString(resolve(name))
			code = after
			continue
		}
		b.WriteString(call)
		code = rest
	}
	b.WriteString(code)
	out := b.String()
	return out, hasDotenvCall(out)
}

// scanDotenvArgs casa `("NOME")` (espaços permitidos dentro dos parênteses,
// nome sem aspas nem quebra de linha) logo após get.Dotenv
func scanDotenvArgs(s string) (name, after string, ok bool) {
	if s == "" || s[0] != '(' {
		return "", "", false
	}
	j := 1
	for j < len(s) && isSpaceByte(s[j]) {
		j++
	}
	if j >= len(s) || s[j] != '"' {
		return "", "", false
	}
	end := j + 1
	for end < len(s) && s[end] != '"' && s[end] != '\n' {
		end++
	}
	if end >= len(s) || s[end] != '"' {
		return "", "", false
	}
	name = s[j+1 : end]
	j = end + 1
	for j < len(s) && isSpaceByte(s[j]) {
		j++
	}
	if j >= len(s) || s[j] != ')' {
		return "", "", false
	}
	return name, s[j+1:], true
}

// hasDotenvCall detecta get.Dotenv seguido de ( (com espaços), a forma que
// sobra quando o argumento não é string literal
func hasDotenvCall(code string) bool {
	for {
		idx := strings.Index(code, "get.Dotenv")
		if idx < 0 {
			return false
		}
		j := idx + len("get.Dotenv")
		for j < len(code) && isSpaceByte(code[j]) {
			j++
		}
		if j < len(code) && code[j] == '(' {
			return true
		}
		code = code[idx+len("get.Dotenv"):]
	}
}

// MakeSleepAsync deixa time.Sleep com cara de bloqueante: a chamada ganha
// await e as funções do chunk viram async para o await valer. Sem time.Sleep
// o chunk passa intocado.
func MakeSleepAsync(code string) string {
	if !strings.Contains(code, "time.Sleep(") {
		return code
	}
	code = replaceWordPrefixed(code, "time.Sleep(", "await", "await time.Sleep(")
	return replaceWordPrefixed(code, "function", "async", "async function")
}

// replaceWordPrefixed troca cada ocorrência de target (com fronteira de
// palavra antes e, para nomes, depois) por repl, absorvendo o prefixo
// opcional prefix + espaços para não duplicar
func replaceWordPrefixed(code, target, prefix, repl string) string {
	var b strings.Builder
	for {
		idx := indexWord(code, target)
		if idx < 0 {
			break
		}
		start := idx
		// absorve "prefix\s+" imediatamente antes, se houver
		k := idx
		for k > 0 && isSpaceByte(code[k-1]) {
			k--
		}
		if k < idx && k >= len(prefix) && code[k-len(prefix):k] == prefix && wordBoundaryBefore(code, k-len(prefix)) {
			start = k - len(prefix)
		}
		b.WriteString(code[:start])
		b.WriteString(repl)
		code = code[idx+len(target):]
	}
	b.WriteString(code)
	return b.String()
}

// indexWord acha target com fronteira de palavra antes (e depois, quando o
// target termina em char de palavra)
func indexWord(code, target string) int {
	from := 0
	for {
		idx := strings.Index(code[from:], target)
		if idx < 0 {
			return -1
		}
		idx += from
		if wordBoundaryBefore(code, idx) && wordBoundaryAfter(code, idx+len(target)) {
			return idx
		}
		from = idx + 1
	}
}

func wordBoundaryBefore(code string, i int) bool {
	return i == 0 || !isWordChar(code[i-1])
}

func wordBoundaryAfter(code string, i int) bool {
	return i >= len(code) || !isWordChar(code[i-1]) || !isWordChar(code[i])
}
