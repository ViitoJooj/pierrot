package compiler

import "testing"

func resolveMap(m map[string]string) func(string) string {
	return func(name string) string {
		if v, ok := m[name]; ok {
			return `"` + v + `"`
		}
		return `""`
	}
}

func TestReplaceDotenvCalls(t *testing.T) {
	out, nonLit := ReplaceDotenvCalls(`const k = get.Dotenv("KEY");`, resolveMap(map[string]string{"KEY": "v"}))
	if out != `const k = "v";` || nonLit {
		t.Fatalf("out: %q nonLit: %v", out, nonLit)
	}
}

func TestReplaceDotenvSpaces(t *testing.T) {
	out, _ := ReplaceDotenvCalls(`get.Dotenv( "A" ) + get.Dotenv("B")`, resolveMap(map[string]string{"A": "1", "B": "2"}))
	if out != `"1" + "2"` {
		t.Fatalf("out: %q", out)
	}
}

func TestReplaceDotenvNonLiteral(t *testing.T) {
	src := `get.Dotenv(name)`
	out, nonLit := ReplaceDotenvCalls(src, resolveMap(nil))
	if out != src || !nonLit {
		t.Fatalf("out: %q nonLit: %v", out, nonLit)
	}
}

func TestReplaceDotenvNonLiteralSpaced(t *testing.T) {
	_, nonLit := ReplaceDotenvCalls(`get.Dotenv (x)`, resolveMap(nil))
	if !nonLit {
		t.Fatal("nonLit devia ser true")
	}
}

func TestMakeSleepAsync(t *testing.T) {
	out := MakeSleepAsync("function foo() { time.Sleep(2).sec(); }")
	want := "async function foo() { await time.Sleep(2).sec(); }"
	if out != want {
		t.Fatalf("out: %q", out)
	}
}

func TestMakeSleepAsyncNoDouble(t *testing.T) {
	out := MakeSleepAsync("async function f() { await time.Sleep(1).sec(); }")
	if out != "async function f() { await time.Sleep(1).sec(); }" {
		t.Fatalf("out: %q", out)
	}
}

func TestMakeSleepAsyncBoundaries(t *testing.T) {
	// myfunction/functional não são function; xtime.Sleep não é time.Sleep
	src := "const myfunction = 1; functional(); xtime.Sleep(1); time.Sleep(1).sec();"
	out := MakeSleepAsync(src)
	want := "const myfunction = 1; functional(); xtime.Sleep(1); await time.Sleep(1).sec();"
	if out != want {
		t.Fatalf("out: %q", out)
	}
}

func TestMakeSleepAsyncUntouchedWithoutSleep(t *testing.T) {
	src := "function foo() {}"
	if out := MakeSleepAsync(src); out != src {
		t.Fatalf("out: %q", out)
	}
}
