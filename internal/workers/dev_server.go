//TODO: melhorar muito esse código!!!
//
// Separar funções, parar de utilizar regex, preciso pensar em uma nova arch,
// para evitar essa bagunça.

package workers

import (
	"fmt"
	"html"
	"io/fs"
	"log"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ViitoJooj/pierrot/internal/readers"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/spf13/cobra"
)

var (
	bindingRe    = regexp.MustCompile(`\$\{(\w+)\}`)                           // ${nome} em texto = binding de nome "nome"
	eventRe      = regexp.MustCompile(`@(\w+)=\{(\w+)(?:\(([^(){}\n]*)\))?\}`) // @evento={fn} ou @evento={fn(args)}
	slotRe       = regexp.MustCompile(`<Slot\s*/>`)                            // tag de slot para renderização de filho único (sem props ou múltiplos slots por enquanto)
	unknownTagRe = regexp.MustCompile(`<([A-Z]\w*)[^>]*/?>`)                   // tags HTML legítimas começam minúsculas, então isso pega tags que parecem componente mas não foram importadas
	dotenvCallRe = regexp.MustCompile(`get\.Dotenv\(\s*"([^"\n]*)"\s*\)`)      // get.Dotenv("NOME") — resolvido no servidor (ver substDotenv)
	dotenvAnyRe  = regexp.MustCompile(`get\.Dotenv\s*\(`)                      // qualquer get.Dotenv( — para avisar erro de argumento não string literal, já que a substituição é textual
)

// substDotenv troca cada get.Dotenv("NOME") do script pelo valor literal do
// .env, antes do bundle. Só as variáveis referenciadas saem no HTML — o resto
// do .env nunca chega no browser. env == nil significa dotenv desabilitado no
// settings. Chamada com argumento que não seja string literal vira erro:
// a substituição é textual, não dá para resolver nome dinâmico
func substDotenv(code, src string, env map[string]string, errs *[]string) string {
	code = dotenvCallRe.ReplaceAllStringFunc(code, func(m string) string {
		name := dotenvCallRe.FindStringSubmatch(m)[1]
		if env == nil {
			*errs = append(*errs, fmt.Sprintf("get.Dotenv(%q) em %s: dotenv desabilitado no settings.pierrot.json", name, src))
			return `""`
		}
		v, ok := env[name]
		if !ok {
			*errs = append(*errs, fmt.Sprintf("get.Dotenv(%q) em %s: variável não existe no .env", name, src))
			return `""`
		}
		return strconv.Quote(v)
	})
	if dotenvAnyRe.MatchString(code) {
		*errs = append(*errs, fmt.Sprintf("get.Dotenv em %s: só aceita string literal entre aspas duplas", src))
	}
	return code
}

// runtimeJS gera o runtime de bindings sem eval: o servidor conhece os nomes
// usados em ${...} e monta um objeto state com eles. typeof protege contra
// binding de variável não declarada (vira string vazia em vez de quebrar tudo).
// blocks e exprs vêm de compileTemplate: cada bloco @for/@if e cada {expr}
// vira uma função reavaliada a cada __pierrotUpdate, então os placeholders
// re-renderizam quando um @evento muda o estado
// compilerJS é o compilador de template portado para o browser, incluído só
// quando a página usa @render pierrot com expressão. Compila o trecho inteiro
// como uma função de render (blocos viram controle de fluxo JS) e monta um
// documento completo para o iframe sandboxed — o <script> do exemplo roda
// isolado lá dentro, sem tocar no estado da página
const compilerJS = `function __pierrotCompile(src) {
	var esc = function (s) { return String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;"); };
	var m = String(src).match(/<script>([\s\S]*?)<\/script>/i);
	var script = m ? m[1] : "";
	var tpl = m ? String(src).replace(m[0], "") : String(src);
	script = script.replace(/^\s*import\s+.*$/gm, "");
	// anotação de tipo em let/const/var some; o resto do TS não é suportado
	script = script.replace(/^(\s*(?:let|const|var)\s+\w+)\s*:\s*[^=;\n]+(?=[=;])/gm, "$1");
	var events = function (l) {
		return l.replace(/@(\w+)=\{(\w+)(?:\(([^(){}\n]*)\))?\}/g, function (_, ev, fn, args) {
			return "on" + ev + '="' + fn + "(" + esc(args || "") + '); __update()"';
		});
	};
	var body = 'var __h = "";\n';
	var lines = tpl.split(/\r?\n/);
	var interp = /\$\{([^{}\n]+)\}/g;
	for (var i = 0; i < lines.length; i++) {
		var l = lines[i], mm;
		if (/^\/\//.test(l)) continue;
		if ((mm = l.match(/^\s*@for\s+(\w+)\s+in\s+(.+?)\s*$/))) { body += "for (const " + mm[1] + " of (" + mm[2] + ")) {\n"; continue; }
		if ((mm = l.match(/^\s*@if\s+(.+?)\s*$/))) { body += "if (" + mm[1] + ") {\n"; continue; }
		if (/^\s*@else\s*$/.test(l)) { body += "} else {\n"; continue; }
		if (/^\s*@(?:end|endif)\s*$/.test(l)) { body += "}\n"; continue; }
		if (!l.trim()) continue;
		l = events(l);
		var part = "__h += ", last = 0, mt;
		interp.lastIndex = 0;
		while ((mt = interp.exec(l))) {
			part += JSON.stringify(l.slice(last, mt.index)) + " + __esc(" + mt[1] + ") + ";
			last = mt.index + mt[0].length;
		}
		part += JSON.stringify(l.slice(last) + "\n") + ";\n";
		body += part;
	}
	body += "return __h;";
	var js = 'function __esc(v){return String(v).replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;").replace(/"/g,"&quot;")}\n'
		+ script
		+ "\nfunction __render() {\n" + body + "\n}\n"
		+ "function __update(){try{document.getElementById('app').innerHTML=__render()}catch(e){document.getElementById('app').innerHTML=\"<pre style='color:#e5484d;white-space:pre-wrap'>\"+__esc(e)+\"</pre>\"}}\n__update();";
	// "</script" dentro das strings geradas fecharia o <script> do documento;
	// só essa sequência — "</" solto aparece em regex legítima (ex. /</g)
	js = js.replace(/<\/script/gi, "<\\/script");
	return '<!DOCTYPE html><html><head><meta charset="UTF-8"><style>body{font-family:system-ui;margin:14px;color:#1a1410}h1,h2{font-family:Georgia,serif;font-weight:400;margin:0 0 10px}button{font-family:ui-monospace,Consolas,monospace;font-size:13px;padding:5px 16px;border:none;border-radius:8px;background:#1a1410;color:#f5f0e4;cursor:pointer}</style></head><body><div id="app"></div><script>' + js + '<\/script></body></html>';
}
`

func runtimeJS(names, blocks, exprs, binds []string, renders []renderEntry) string {
	var b strings.Builder
	for _, r := range renders {
		if r.kind == "pierrot" {
			b.WriteString(compilerJS)
			break
		}
	}
	b.WriteString(`function __pierrotEsc(v) {
	return String(v).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
}
function __pierrotMd(src) {
	var esc = function (s) { return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;"); };
	var inline = function (s) {
		return esc(s)
			.replace(/\x60([^\x60]+)\x60/g, "<code>$1</code>")
			.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>")
			.replace(/\*([^*]+)\*/g, "<em>$1</em>")
			.replace(/\[([^\]]+)\]\(([^()\s]+)\)/g, '<a href="$2">$1</a>');
	};
	var lines = String(src).split(/\r?\n/), out = [], i = 0;
	while (i < lines.length) {
		var l = lines[i];
		if (/^\x60\x60\x60/.test(l)) {
			var buf = [];
			for (i++; i < lines.length && !/^\x60\x60\x60/.test(lines[i]); i++) buf.push(lines[i]);
			i++;
			out.push("<pre><code>" + esc(buf.join("\n")) + "</code></pre>");
			continue;
		}
		var h = l.match(/^(#{1,6})\s+(.*)$/);
		if (h) { out.push("<h" + h[1].length + ">" + inline(h[2]) + "</h" + h[1].length + ">"); i++; continue; }
		if (/^\s*[-*]\s+/.test(l)) {
			var items = [];
			for (; i < lines.length && /^\s*[-*]\s+/.test(lines[i]); i++) items.push("<li>" + inline(lines[i].replace(/^\s*[-*]\s+/, "")) + "</li>");
			out.push("<ul>" + items.join("") + "</ul>");
			continue;
		}
		if (/^\s*$/.test(l)) { i++; continue; }
		var para = [];
		for (; i < lines.length && !/^\s*$/.test(lines[i]) && !/^(#{1,6}\s|\x60\x60\x60|\s*[-*]\s)/.test(lines[i]); i++) para.push(lines[i]);
		out.push("<p>" + para.map(inline).join("<br>") + "</p>");
	}
	return out.join("\n");
}
var __pierrotBlocks = [
`)
	for _, code := range blocks {
		fmt.Fprintf(&b, "function () {\n%s\n},\n", code)
	}
	b.WriteString("];\nvar __pierrotExprs = [\n")
	for _, e := range exprs {
		fmt.Fprintf(&b, "function () { return (%s); },\n", e)
	}
	b.WriteString("];\nvar __pierrotRenders = [\n")
	for _, r := range renders {
		fmt.Fprintf(&b, "{ kind: %q, fn: function () { return (%s); } },\n", r.kind, r.expr)
	}
	b.WriteString("];\nvar __pierrotBindVals = [\n")
	for _, e := range binds {
		fmt.Fprintf(&b, "function () { return (%s); },\n", e)
	}
	b.WriteString("];\nfunction __pierrotUpdate() {\n\tconst state = {\n")
	for _, n := range names {
		fmt.Fprintf(&b, "\t\t%q: typeof %s === \"undefined\" ? \"\" : %s,\n", n, n, n)
	}
	b.WriteString(`	};
	document.querySelectorAll("[data-bind]").forEach(function (el) {
		el.textContent = state[el.getAttribute("data-bind")];
	});
	__pierrotBlocks.forEach(function (fn, i) {
		document.querySelectorAll('[data-pierrot-block="' + i + '"]').forEach(function (el) {
			try { el.innerHTML = fn(); } catch (e) { console.error("pierrot: bloco " + i + ":", e); }
		});
	});
	__pierrotExprs.forEach(function (fn, i) {
		document.querySelectorAll('[data-pierrot-expr="' + i + '"]').forEach(function (el) {
			try { el.textContent = fn(); } catch (e) { console.error("pierrot: {expr} " + i + ":", e); }
		});
	});
	__pierrotBindVals.forEach(function (fn, i) {
		document.querySelectorAll('[data-pierrot-bindval="' + i + '"]').forEach(function (el) {
			if (el === document.activeElement) return;
			try { var v = fn(); el.value = v == null ? "" : v; } catch (e) {}
		});
	});
	__pierrotRenders.forEach(function (r, i) {
		document.querySelectorAll('[data-pierrot-render="' + i + '"]').forEach(function (el) {
			try {
				if (r.kind === "pierrot") {
					// preview num iframe, recriado só quando o valor muda;
					// debounce de 300ms segura recompilação durante digitação
					var fr = el.firstChild;
					if (!fr) {
						fr = document.createElement("iframe");
						fr.className = "pierrot-preview";
						fr.setAttribute("sandbox", "allow-scripts");
						el.appendChild(fr);
					}
					var v = String(r.fn());
					if (v === el.__pierrotSrc) return;
					var first = el.__pierrotSrc === undefined;
					el.__pierrotSrc = v;
					clearTimeout(el.__pierrotT);
					el.__pierrotT = setTimeout(function () { fr.srcdoc = __pierrotCompile(v); }, first ? 0 : 300);
					return;
				}
				var v = String(r.fn());
				el.innerHTML = r.kind === "markdown" ? __pierrotMd(v) : v;
			} catch (e) { console.error("pierrot: @render " + i + ":", e); }
		});
	});
}
__pierrotUpdate();
`)
	return b.String()
}

// preludeJS define os helpers `get`, `client` e `time` disponíveis no
// <script> das páginas. Roda antes dos scripts da página, então pode ser
// usado em declaração de topo.
// get.Path(n): n-ésimo segmento do pathname, 1-based ("/about/team": 1 = "about",
// 2 = "team"); fora do range devolve ""
// get.Status(): status HTTP da página como string ("200", "404"...), vindo do
// __pierrotStatus injetado pelo renderPage
// get.Dotenv("NOME"): resolvido no servidor (substDotenv) — não existe em
// runtime; o bundle já sai com o valor literal
// client.Redirect(url): navega para a url; .newTab() abre em nova aba em vez
// de navegar (a navegação é agendada num setTimeout(0) para o .newTab poder
// cancelá-la na mesma chamada síncrona)
// time.Sleep(n).sec(): Promise que resolve após n segundos (msec/sec/min/hr/
// day/week/month/year); depois que o código após o await roda, dispara
// __pierrotUpdate para os bindings refletirem o estado novo. O compilador
// injeta o await (ver makeSleepAsync)
const preludeJS = `var get = {
	Path: function (n) {
		var p = location.pathname.split("/").filter(Boolean);
		return n >= 1 && n <= p.length ? p[n - 1] : "";
	},
	Status: function () {
		return String(typeof __pierrotStatus === "undefined" ? 200 : __pierrotStatus);
	},
};
var client = {
	Redirect: function (url) {
		var t = setTimeout(function () { location.assign(url); }, 0);
		return {
			newTab: function () {
				clearTimeout(t);
				window.open(url, "_blank", "noopener");
			},
		};
	},
};
var time = {
	Sleep: function (n) {
		var wait = function (ms) {
			return new Promise(function (done) {
				setTimeout(function () {
					done();
					// microtask depois da continuação do await: o estado mudado
					// após o Sleep já está aplicado quando o update roda
					if (typeof __pierrotUpdate === "function") Promise.resolve().then(__pierrotUpdate);
				}, n * ms);
			});
		};
		return {
			msec: function () { return wait(1); },
			sec: function () { return wait(1000); },
			min: function () { return wait(60 * 1000); },
			hr: function () { return wait(60 * 60 * 1000); },
			day: function () { return wait(24 * 60 * 60 * 1000); },
			week: function () { return wait(7 * 24 * 60 * 60 * 1000); },
			month: function () { return wait(30 * 24 * 60 * 60 * 1000); },
			year: function () { return wait(365 * 24 * 60 * 60 * 1000); },
		};
	},
};
`

// live reload: escuta o servidor e recarrega quando algum arquivo muda
const reloadJS = `
new EventSource("/__pierrot/events").onmessage = function () { location.reload(); };
`

func DevServer(cmd *cobra.Command, args []string) {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	p, err := readers.LoadProject(root)
	if err != nil {
		log.Fatal(err)
	}
	if p.Dotenv != "" {
		if p.Env, err = readers.LoadDotenv(p.Dotenv); err != nil {
			log.Printf("dotenv: %v", err)
		}
	}

	go watchFiles(p.Root)
	http.HandleFunc("/__pierrot/events", reloadEvents)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// sem isso o browser aplica cache heurístico no Last-Modified do
		// ServeFile e segura css/js velho mesmo após o live reload
		w.Header().Set("Cache-Control", "no-store")
		page := strings.Trim(r.URL.Path, "/")
		if page == "" {
			page = defaultPage(p)
		}

		// arquivos com extensão (css, imagens...) são servidos direto do src;
		// /robots.txt vem do arquivo apontado por set.Robots
		if filepath.Ext(page) != "" {
			if page == "robots.txt" {
				if path, ok := robotsPath(p); ok {
					http.ServeFile(w, r, path)
					return
				}
			}
			http.ServeFile(w, r, filepath.Join(p.Src, filepath.Clean("/"+page)))
			return
		}

		comp, err := readers.ParsePierrot(filepath.Join(p.Src, "pages", page, "index.pierrot"))
		if err != nil {
			log.Printf("404 %s: %v", r.URL.Path, err)
			// rota inexistente cai na página de set.Fallback, com get.Status() = "404"
			if fb, ok := fallbackPage(p); ok {
				if fbComp, fbErr := readers.ParsePierrot(filepath.Join(p.Src, "pages", fb, "index.pierrot")); fbErr == nil {
					if html, _, rErr := renderPage(p, fb, fbComp, true, http.StatusNotFound); rErr == nil {
						w.Header().Set("Content-Type", "text/html; charset=utf-8")
						w.WriteHeader(http.StatusNotFound)
						fmt.Fprint(w, html)
						return
					}
				}
			}
			serveError(w, http.StatusNotFound, fmt.Sprintf("página /%s não encontrada: %v", page, err))
			return
		}

		html, _, err := renderPage(p, page, comp, true, http.StatusOK)
		if err != nil {
			serveError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, html)
	})

	fmt.Printf("pierrot dev server: http://localhost:%d\n", p.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", p.Port), nil))
}

// ---------- live reload ----------

var reloadClients = struct {
	sync.Mutex
	m map[chan struct{}]struct{}
}{m: map[chan struct{}]struct{}{}}

// reloadEvents mantém uma conexão SSE aberta com o browser e avisa quando recarregar
func reloadEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming não suportado", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	ch := make(chan struct{}, 1)
	reloadClients.Lock()
	reloadClients.m[ch] = struct{}{}
	reloadClients.Unlock()
	defer func() {
		reloadClients.Lock()
		delete(reloadClients.m, ch)
		reloadClients.Unlock()
	}()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			fmt.Fprint(w, "data: reload\n\n")
			flusher.Flush()
		}
	}
}

func notifyReload() {
	reloadClients.Lock()
	defer reloadClients.Unlock()
	for ch := range reloadClients.m {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// watchFiles compara mtimes a cada 300ms e dispara reload quando algo muda
func watchFiles(root string) {
	prev := snapshot(root)
	for {
		time.Sleep(300 * time.Millisecond)
		cur := snapshot(root)
		if !maps.Equal(prev, cur) {
			prev = cur
			notifyReload()
		}
	}
}

func snapshot(root string) map[string]int64 {
	s := map[string]int64{}
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if name := d.Name(); name == "node_modules" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(p) {
		case ".pierrot", ".css", ".ts", ".js", ".html":
			if info, err := d.Info(); err == nil {
				s[p] = info.ModTime().UnixNano()
			}
		}
		return nil
	})
	return s
}

// ---------- overlay de erros ----------

// overlayHTML monta o popup que aparece no browser listando os erros
func overlayHTML(errs []string) string {
	if len(errs) == 0 {
		return ""
	}
	var items strings.Builder
	for _, e := range errs {
		fmt.Fprintf(&items, "<li>%s</li>", html.EscapeString(e))
	}
	return fmt.Sprintf(`<div id="pierrot-overlay" style="position:fixed;left:16px;right:16px;bottom:16px;z-index:99999;background:#1b1b1f;color:#f4f4f5;border:1px solid #e5484d;border-left:6px solid #e5484d;border-radius:8px;box-shadow:0 8px 30px rgba(0,0,0,.5);font-family:ui-monospace,Consolas,monospace;font-size:13px;max-height:45vh;overflow:auto;">
<div style="display:flex;align-items:center;justify-content:space-between;padding:10px 14px;border-bottom:1px solid #333;">
<strong style="color:#e5484d;">pierrot — %d erro(s)</strong>
<button onclick="document.getElementById('pierrot-overlay').remove()" style="background:none;border:none;color:#999;cursor:pointer;font-size:16px;">&times;</button>
</div>
<ul style="margin:0;padding:10px 14px 12px 30px;line-height:1.7;">%s</ul>
</div>`, len(errs), items.String())
}

// serveError responde com uma página mínima contendo só o overlay + live reload,
// assim corrigir o arquivo recarrega a página sozinho
func serveError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>pierrot — erro</title></head>
<body style="background:#111;">
%s
<script>%s</script>
</body>
</html>`, overlayHTML([]string{msg}), reloadJS)
}

// ---------- render ----------

// metaPage resolve set.<key>(X) do entry (main.pierrot): X precisa ser um
// import de página, e o diretório dele (relativo a pages/) vira a rota
func metaPage(p *readers.Project, key string) (string, bool) {
	layout, err := readers.ParsePierrot(p.Entry)
	if err != nil {
		return "", false
	}
	name, ok := layout.Meta[key]
	if !ok {
		return "", false
	}
	for _, imp := range layout.Imports {
		if imp.Name != name {
			continue
		}
		rel, err := filepath.Rel(filepath.Join(p.Src, "pages"), filepath.Dir(filepath.Join(p.Src, imp.Path)))
		if err != nil || strings.HasPrefix(rel, "..") {
			log.Printf("set.%s(%s): import %q não está em pages/", key, name, imp.Path)
			return "", false
		}
		return filepath.ToSlash(rel), true
	}
	log.Printf("set.%s(%s): nenhum import com esse nome no main.pierrot", key, name)
	return "", false
}

// defaultPage resolve set.Default(X): a página da rota "/"
func defaultPage(p *readers.Project) string {
	if page, ok := metaPage(p, "Default"); ok {
		return page
	}
	return "home"
}

// fallbackPage resolve set.Fallback(X): a página servida quando a rota não existe
func fallbackPage(p *readers.Project) (string, bool) {
	return metaPage(p, "Fallback")
}

// robotsPath resolve set.Robots("./caminho") do entry para o arquivo dentro do
// src, servido em /robots.txt
func robotsPath(p *readers.Project) (string, bool) {
	layout, err := readers.ParsePierrot(p.Entry)
	if err != nil {
		return "", false
	}
	rel, ok := layout.Meta["Robots"]
	if !ok {
		return "", false
	}
	return filepath.Join(p.Src, filepath.FromSlash(strings.TrimPrefix(rel, "./"))), true
}

var (
	sleepCallRe = regexp.MustCompile(`(?:await\s+)?\btime\.Sleep\(`)
	syncFuncRe  = regexp.MustCompile(`(?:async\s+)?\bfunction\b`)
)

// makeSleepAsync deixa time.Sleep com cara de bloqueante: a chamada ganha
// await e as funções do chunk viram async para o await valer. Sem time.Sleep
// o chunk passa intocado
func makeSleepAsync(code string) string {
	if !strings.Contains(code, "time.Sleep(") {
		return code
	}
	code = sleepCallRe.ReplaceAllString(code, "await time.Sleep(")
	return syncFuncRe.ReplaceAllString(code, "async function")
}

// chunk é um pedaço de script com origem rastreável para mensagens de erro
type chunk struct {
	name string // arquivo de origem, relativo ao root
	code string
}

// renderCtx acumula estilos, scripts, meta e erros enquanto o layout, a página
// e os componentes importados são expandidos
type renderCtx struct {
	root    string
	styles  []string // hrefs em ordem de descoberta, sem duplicar
	seen    map[string]bool
	scripts []chunk
	meta    map[string]string
	errs    []string // erros não fatais, mostrados no overlay
}

// rel devolve o caminho de um arquivo relativo ao root, para mensagens
func (rc *renderCtx) rel(path string) string {
	if r, err := filepath.Rel(rc.root, path); err == nil {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(path)
}

// href converte um import de css relativo ao arquivo em uma URL servida pelo dev server
func (rc *renderCtx) href(dir, rel string) string {
	abs := filepath.Join(dir, rel)
	r, err := filepath.Rel(rc.root, abs)
	if err != nil {
		return "/" + strings.TrimPrefix(rel, "./")
	}
	return "/" + filepath.ToSlash(r)
}

// render expande um componente: coleta css/scripts/meta e substitui as tags
// dos componentes importados pelo HTML deles (recursivo). src é o arquivo de
// origem, usado nas mensagens de erro
func (rc *renderCtx) render(c *readers.Component, dir, src string) string {
	for _, s := range c.Styles {
		href := rc.href(dir, s)
		if !rc.seen[href] {
			rc.seen[href] = true
			rc.styles = append(rc.styles, href)
		}
	}
	// import "./x.ts" inclui o conteúdo do arquivo no bundle da página
	for _, s := range c.Scripts {
		path := filepath.Join(dir, s)
		if rc.seen[path] {
			continue
		}
		rc.seen[path] = true
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("script %s: %v", s, err)
			rc.errs = append(rc.errs, fmt.Sprintf("script %q importado em %s: arquivo não encontrado", s, src))
			continue
		}
		rc.scripts = append(rc.scripts, chunk{name: rc.rel(path), code: string(data)})
	}
	maps.Copy(rc.meta, c.Meta)
	if c.Script != "" {
		rc.scripts = append(rc.scripts, chunk{name: src, code: c.Script})
	}

	tpl := c.Template
	for _, imp := range c.Imports {
		tagRe := regexp.MustCompile(`<` + regexp.QuoteMeta(imp.Name) + `\b([^>]*?)/>`)
		// import sem tag no template (ex.: página usada só em set.Default) é ignorado
		if !tagRe.MatchString(tpl) {
			continue
		}
		childPath := filepath.Join(dir, imp.Path)
		child, err := readers.ParsePierrot(childPath)
		if err != nil {
			log.Printf("componente <%s />: %v", imp.Name, err)
			rc.errs = append(rc.errs, fmt.Sprintf("componente <%s />: %v", imp.Name, err))
			tpl = tagRe.ReplaceAllLiteralString(tpl, fmt.Sprintf("<!-- componente %s não encontrado -->", imp.Name))
			continue
		}
		// css/scripts coletados uma vez; props aplicadas por instância da tag
		childHTML := rc.render(child, filepath.Dir(childPath), rc.rel(childPath))
		// StringFunc devolve o HTML literal (sem expandir $ de grupo), e
		// childHTML pode conter ${var}
		tpl = tagRe.ReplaceAllStringFunc(tpl, func(tag string) string {
			attrs := tagRe.FindStringSubmatch(tag)[1]
			return applyProps(childHTML, child.Props, parseProps(attrs))
		})
	}
	return tpl
}

// renderPage monta o HTML final: main.pierrot (layout, se existir) envolve a
// página via <Slot />; set.X("...") da página sobrescreve o do layout.
// dev=true injeta overlay de erros + live reload e linka cada css separado;
// dev=false (build) minifica conforme settings, linka um único
// /<página>/bundle.css e qualquer erro vira fatal. status é o HTTP status da
// página (200 para páginas normais, 404 para a de set.Fallback), exposto ao
// script via get.Status(). Retorna também os hrefs de css na ordem, para o
// build gerar o bundle
func renderPage(p *readers.Project, page string, comp *readers.Component, dev bool, status int) (string, []string, error) {
	rc := &renderCtx{
		root: p.Src,
		seen: map[string]bool{},
		meta: map[string]string{},
	}

	var layoutHTML string
	layout, layoutErr := readers.ParsePierrot(p.Entry)
	if layoutErr == nil {
		layoutHTML = rc.render(layout, p.Src, filepath.Base(p.Entry))
	}

	pageHTML := rc.render(comp, filepath.Join(p.Src, "pages", page), "pages/"+page+"/index.pierrot")

	body := pageHTML
	if layoutErr == nil {
		// LiteralString: pageHTML pode conter ${var}, que ReplaceAllString
		// interpretaria como referência de grupo de captura e apagaria
		body = slotRe.ReplaceAllLiteralString(layoutHTML, pageHTML)
	}

	// tag capitalizada que sobrou no HTML = componente usado sem import
	seenTags := map[string]bool{}
	for _, m := range unknownTagRe.FindAllStringSubmatch(body, -1) {
		if !seenTags[m[1]] {
			seenTags[m[1]] = true
			rc.errs = append(rc.errs, fmt.Sprintf("tag <%s /> no HTML não corresponde a nenhum componente importado", m[1]))
		}
	}

	// linhas @render: pierrot é compilado e inlined; html/markdown viram
	// placeholders preenchidos pelo runtime
	body, renders := rc.expandRenders(body)

	// comentários //, blocos @for/@if e interpolações ${expr} do template
	ct := compileTemplate(body)
	rc.errs = append(rc.errs, ct.errs...)
	body = ct.html

	// nomes dos bindings ${var}, na ordem e sem repetir, para gerar o state do runtime
	var bindNames []string
	seenBind := map[string]bool{}
	for _, m := range bindingRe.FindAllStringSubmatch(body, -1) {
		if !seenBind[m[1]] {
			seenBind[m[1]] = true
			bindNames = append(bindNames, m[1])
		}
	}

	// @bind={var} vira oninput + marcador de valor; antes de expandEvents,
	// que casaria @bind como @evento
	body = ct.expandBinds(body)
	// ${var} vira um span que o runtime preenche
	body = bindingRe.ReplaceAllString(body, `<span data-bind="$1"></span>`)
	// @click={fn} ou @click={fn(args)} vira onclick nativo + atualização do DOM
	body = expandEvents(body)
	// ${expr} composto que sobrou vira span preenchido pelo runtime (depois de ${var} e @evento)
	body = ct.replaceExprs(body)

	// TypeScript -> JavaScript, um chunk por arquivo para o erro apontar a origem;
	// chunk com erro fica de fora e o erro vai pro overlay
	minify := !dev && p.Minify
	sourcemap := api.SourceMapNone
	if !dev && p.Sourcemap {
		sourcemap = api.SourceMapInline
	}
	var js strings.Builder
	for _, ch := range rc.scripts {
		// get.Dotenv("X") vira o valor literal antes do transform
		code := substDotenv(ch.code, ch.name, p.Env, &rc.errs)
		res := api.Transform(makeSleepAsync(code), api.TransformOptions{
			Loader: api.LoaderTS,
			// sem MinifyIdentifiers: renomear variável de topo quebraria o
			// state gerado para os bindings ${var}
			MinifyWhitespace: minify,
			MinifySyntax:     minify,
			Sourcemap:        sourcemap,
			Sourcefile:       ch.name,
		})
		for _, e := range res.Errors {
			loc := ""
			if e.Location != nil {
				loc = fmt.Sprintf(", linha %d", e.Location.Line)
			}
			rc.errs = append(rc.errs, fmt.Sprintf("erro de script em %s%s: %s", ch.name, loc, e.Text))
		}
		if len(res.Errors) == 0 {
			js.Write(res.Code)
			js.WriteString("\n")
		}
	}

	// no build qualquer erro derruba a página
	if !dev && len(rc.errs) > 0 {
		return "", nil, fmt.Errorf("%s", strings.Join(rc.errs, "\n  "))
	}

	charset := rc.meta["Charset"]
	if charset == "" {
		charset = "UTF-8"
	}

	var head strings.Builder
	fmt.Fprintf(&head, "<meta charset=%q>\n", charset)
	if t, ok := rc.meta["Title"]; ok {
		fmt.Fprintf(&head, "<title>%s</title>\n", t)
	}
	if d, ok := rc.meta["Description"]; ok {
		fmt.Fprintf(&head, "<meta name=\"description\" content=%q>\n", d)
	}
	if i, ok := rc.meta["Icon"]; ok {
		fmt.Fprintf(&head, "<link rel=\"icon\" href=%q>\n", i)
	}
	if dev {
		for _, href := range rc.styles {
			fmt.Fprintf(&head, "<link rel=\"stylesheet\" href=%q>\n", href)
		}
	} else if len(rc.styles) > 0 {
		// build: um único css por página, gerado pelo worker de build
		fmt.Fprintf(&head, "<link rel=\"stylesheet\" href=%q>\n", "/"+page+"/bundle.css")
	}

	overlay, reload := "", ""
	if dev {
		overlay = overlayHTML(rc.errs)
		reload = reloadJS
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
%s</head>
<body>
%s
%s
<script>
var __pierrotStatus = %d;
%s
%s
%s
%s
</script>
</body>
</html>`, head.String(), body, overlay, status, preludeJS, js.String(), runtimeJS(bindNames, ct.blocks, ct.exprs, ct.binds, renders), reload), rc.styles, nil
}
