# Pierrot Documentation

> Pierrot is a **single-file component** framework compiled in **Go**.
> The CLI understands the `.pierrot` language, compiles everything to static HTML
> and ships the browser a runtime of a few kilobytes — no virtual DOM, no
> `eval`, no `node_modules`.

This is the full documentation. If you just want to see something running, start
with the [Hello World in 30s](./getting-started.md).

---

## Table of contents

### Getting started
- [**Installation & Hello World in 30s**](./getting-started.md) — install the binary, create and run your first project.
- [**Project structure**](./estrutura-do-projeto.md) — what each folder and file is, and the anatomy of a `.pierrot`.

### The `.pierrot` language
- [**Template syntax**](./templates.md) — `${}` interpolation, comments, `@for`, `@if/@else`.
- [**Components & props**](./componentes.md) — single-file components, `import`, `<Slot />`, passing props.
- [**Reactivity**](./reatividade.md) — `${var}`, `@bind`, `@event`, and `@render html/markdown/pierrot`.
- [**Script API**](./script-api.md) — the `get`, `client`, `time`, `set` helpers and `.env`.

### Tooling
- [**CLI & configuration**](./cli.md) — `init`, `dev`, `build`, `vscode` and `settings.pierrot.json`.

### Reference
- [**Architecture**](#architecture) (this page) — the compilation pipeline, step by step.

---

## What Pierrot is

You write the UI in `.pierrot` files. Each file is a single component that bundles
three things:

```html
<script>
    // imports, metadata (set.X), props (let name: type;) and TypeScript logic
    let count: number = 0;

    function increment() {
        count++;
    }
</script>

<!-- template: HTML + directives (@for, @if, @event, ${}) -->
<h1>Clicks: ${count}</h1>
<button @click={increment}>+1</button>
```

The server (`pierrot dev`) or the build (`pierrot build`) reads this file, resolves
imports and components, turns the TypeScript into JavaScript with
[esbuild](https://esbuild.github.io/) and assembles a complete HTML document. The
browser gets ready-made HTML + a small runtime that re-evaluates the dynamic bits
when an event changes the state.

There's no JS build step in your project, no `package.json`, no virtual DOM. State
is just the top-level variables of the `<script>`; every `@event` calls
`__pierrotUpdate()` at the end to repaint the affected spans and blocks.

---

## Architecture

The whole compiler lives in `internal/`. It's a short pipeline of pure functions
that runs **per request** under `pierrot dev` and **once per page** under
`pierrot build`. No giant AST, no worker pool, no cache.

| # | Step | File | What it does |
|---|------|------|--------------|
| 1 | **parse** | `internal/readers/parser.go` | Splits the `<script>` from the template. Extracts CSS/TS/component imports, `set.X` metadata and props (`let name: type;` with no value). |
| 2 | **expansion** | `internal/workers/dev_server.go` (`render`) | Recursive: layout → page → components. `<Slot />` receives the page; each `<Name />` receives the imported component's HTML, with props applied per instance. |
| 3 | **template** | `internal/workers/template.go` | `@for`/`@if` blocks become JS functions. `@render` becomes a placeholder. `//` comments are dropped. |
| 4 | **bindings** | `internal/workers/dev_server.go` (`renderPage`) | Simple `${var}` becomes `<span data-bind>`. `@bind` becomes `oninput`. `@event` becomes `on<event>` + `__pierrotUpdate()`. Composite `${expr}` becomes a re-evaluated span. |
| 5 | **transform** | esbuild (`build.go` / `renderPage`) | Each `<script>` and each `import "*.ts"` becomes a TS → JS chunk. `get.Dotenv("X")` is replaced by the literal value before transform. |
| 6 | **assembly** | `dev_server.go` (`renderPage`) | Final HTML: `<head>` with `title`/`meta`/CSS `links`, `<body>` with the markup, and the generated runtime (`preludeJS` + scripts + `runtimeJS`). |

### The browser runtime

The runtime uses no `eval`. The server already knows the names used in `${...}`
and generates a `state` object from them. On every `__pierrotUpdate()`:

- `data-bind` spans get the variable's value;
- `data-pierrot-block` blocks (`@for`/`@if`) re-run their JS function and
  re-render;
- `data-pierrot-expr` spans (composite expressions) are re-evaluated;
- inputs with `@bind` get the variable's value (skipping the focused element);
- `@render` placeholders are refilled (html/markdown/preview iframe).

`time.Sleep(...)` fires a `__pierrotUpdate()` after the code following the `await`
runs, so state changed after the sleep is already on screen.

### dev vs build

|  | `pierrot dev` | `pierrot build` |
|--|---------------|-----------------|
| When it compiles | every request | once per page |
| CSS | one `<link>` per file | one minified `bundle.css` per page |
| Errors | overlay in the browser + live reload | abort the build (non-zero exit) |
| Live reload | yes (SSE on `/__pierrot/events`) | — |
| Output | in memory | `outDir/` (static HTML + assets) |

→ Details for each command in [CLI & configuration](./cli.md).

---

## Quick conventions

- **Routes** = folders in `src/pages/`. `pages/about/index.pierrot` → route `/about`.
- The **default page** (`/`) and **fallback** (404) are set in the layout with
  `set.Default(...)` and `set.Fallback(...)`.
- **Components** start with an **uppercase letter** (`<Header />`); lowercase
  tags are plain HTML.
- **Props** are declared in the component as `let name: type;` **with no value**.
- **Assets** (images, fonts, `robots.txt`, `favicon.ico`) live in `src/assets/`
  and are copied on build.

Continue to [**Hello World in 30s →**](./getting-started.md)
