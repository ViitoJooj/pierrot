package compiler

import (
	"reflect"
	"testing"
)

func TestSplitScript(t *testing.T) {
	script, tpl := SplitScript("<h1>a</h1>\n<script>\nlet x = 1;\n</script>\n<p>b</p>")
	if script != "\nlet x = 1;\n" {
		t.Fatalf("script: %q", script)
	}
	if tpl != "<h1>a</h1>\n\n<p>b</p>" {
		t.Fatalf("template: %q", tpl)
	}
}

func TestSplitScriptNone(t *testing.T) {
	script, tpl := SplitScript("<h1>a</h1>")
	if script != "" || tpl != "<h1>a</h1>" {
		t.Fatalf("got %q, %q", script, tpl)
	}
}

func TestSplitScriptMultiple(t *testing.T) {
	// comportamento do regex antigo: primeiro <script> vira o Script, todos
	// os blocos <script> somem do template
	script, tpl := SplitScript("<script>a</script>x<script>b</script>y")
	if script != "a" {
		t.Fatalf("script: %q", script)
	}
	if tpl != "xy" {
		t.Fatalf("template: %q", tpl)
	}
}

// golden: espelha o comportamento das regexes antigas de readers/parser.go
func TestParseScriptGolden(t *testing.T) {
	src := `
import "./style.css";
import "./extra.css"
import "./code.ts";
import "./legacy.js";
import { Card } from "./card.pierrot";
  import { Nav } from "../nav.pierrot"

set.Title("Minha Página");
set.Default(Home)
set.Empty("");

let count;
let name: string;
  let price: number ;
let assigned = 1;

function inc() { count++; }
`
	p := ParseScript(src)
	if !reflect.DeepEqual(p.Styles, []string{"./style.css", "./extra.css"}) {
		t.Fatalf("styles: %v", p.Styles)
	}
	if !reflect.DeepEqual(p.Scripts, []string{"./code.ts", "./legacy.js"}) {
		t.Fatalf("scripts: %v", p.Scripts)
	}
	if !reflect.DeepEqual(p.Imports, []ScriptImport{{"Card", "./card.pierrot"}, {"Nav", "../nav.pierrot"}}) {
		t.Fatalf("imports: %v", p.Imports)
	}
	want := map[string]string{"Title": "Minha Página", "Default": "Home", "Empty": ""}
	if !reflect.DeepEqual(p.Meta, want) {
		t.Fatalf("meta: %v", p.Meta)
	}
	if !reflect.DeepEqual(p.Props, []string{"count", "name", "price"}) {
		t.Fatalf("props: %v", p.Props)
	}
	// let com = fica no script (não é prop); função fica
	if !contains(p.Script, "let assigned = 1;") || !contains(p.Script, "function inc()") {
		t.Fatalf("script: %q", p.Script)
	}
	// linhas classificadas somem do script
	for _, gone := range []string{"import", "set.Title", "let count;", "let name:"} {
		if contains(p.Script, gone) {
			t.Fatalf("script ainda tem %q: %q", gone, p.Script)
		}
	}
}

func TestParseScriptEdge(t *testing.T) {
	p := ParseScript("let x2_ ;\nset.K(v)\nimport \"a.css.bak\";\n")
	// let com espaço antes do ; casa (`\s*` antes do `;`)? Não: propRe é
	// `let\s+(\w+)\s*(?::[^=;\n]*)?;` — \s* entre nome e ':'/';' … "let x2_ ;"
	// casa (\s* depois do nome, sem tipo). Golden do comportamento antigo.
	if !reflect.DeepEqual(p.Props, []string{"x2_"}) {
		t.Fatalf("props: %v", p.Props)
	}
	if p.Meta["K"] != "v" {
		t.Fatalf("meta: %v", p.Meta)
	}
	// .css.bak não é css: linha fica no script
	if len(p.Styles) != 0 || !contains(p.Script, "a.css.bak") {
		t.Fatalf("styles: %v script: %q", p.Styles, p.Script)
	}
}

func contains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && index(s, sub) >= 0
}

func index(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
