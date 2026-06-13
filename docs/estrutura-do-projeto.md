# Project structure

[← back to index](./arquitetura.md)

## Folder layout

```txt
myapp/
├── src/
│   ├── assets/                     # static files
│   │   ├── robots.txt              # crawler instructions
│   │   └── favicon.ico             # site icon
│   ├── components/                 # reusable components
│   │   └── header/
│   │       ├── index.pierrot       # component template + script
│   │       ├── script.ts           # logic (optional)
│   │       └── styles.css          # styles (optional)
│   ├── pages/                      # application routes
│   │   ├── errors/                 # fallback page (404)
│   │   │   └── index.pierrot
│   │   └── home/                   # "/" route (default page)
│   │       └── index.pierrot
│   ├── globals.css                 # global styles and variables
│   └── main.pierrot                # layout / entry point
└── settings.pierrot.json           # project configuration
```

The paths aren't magic: the entry point is `app.entry` from
`settings.pierrot.json` (default `./src/main.pierrot`), and its folder becomes the
project's `src` root. Everything else is relative to that folder.

→ See [configuration](./cli.md#settingspierrotjson) to change these paths.

---

## Routes = folders in `pages/`

Each `pages/<route>/index.pierrot` becomes a route. The folder name is the path:

| Folder | Route |
|--------|-------|
| `pages/home/index.pierrot` | depends on `set.Default` (usually `/`) |
| `pages/about/index.pierrot` | `/about` |
| `pages/blog/post/index.pierrot` | `/blog/post` |

The `/` route points to the page marked with `set.Default(...)` in the layout.
Non-existent routes fall back to the `set.Fallback(...)` page with status 404.

→ How to set default and fallback: [Script API › `set`](./script-api.md#set--metadata).

---

## Anatomy of a `.pierrot` file

Every `.pierrot` — page, component or layout — has the same shape: an optional
`<script>` followed by the template.

```html
<script>
    /* 1. imports */
    import "./styles.css";                                  // page/component CSS
    import "./script.ts";                                   // TS/JS bundled in
    import { Card } from "../../components/card/index.pierrot"; // component

    /* 2. metadata (only takes effect in the layout/page) */
    set.Title("My page");

    /* 3. props — let WITH NO value (components only) */
    let title: string;

    /* 4. state and logic */
    let count: number = 0;
    function increment() { count++; }
</script>

<!-- 5. template: HTML + directives -->
<h1>${title}</h1>
<button @click={increment}>${count}</button>
```

What the parser does with each part (`internal/readers/parser.go`):

| In `<script>` | Becomes | Detail |
|---------------|---------|--------|
| `import "x.css";` | stylesheet | added to `<head>` (dev) or to the bundle (build) |
| `import "x.ts";` / `"x.js";` | script chunk | bundled into the page |
| `import { Name } from "x.pierrot";` | component | usable as `<Name />` in the template |
| `set.X("...")` / `set.X(Ident)` | metadata | title, icon, default, fallback... |
| `let name: type;` (no value) | **prop** | declares a component prop |
| any other code | logic | transformed TS → JS and sent to the browser |

Everything **outside** the `<script>` is the **template**.

> The `import`, `set.X` and prop-declaration lines are **stripped** from the
> script that ships to the browser — they're instructions for the compiler, not
> runtime code.

### The three roles of a `.pierrot`

- **Layout** (`main.pierrot`): the entry. Has `<Slot />` (where the page is
  injected) and the global `set.X`. May wrap the page with components
  (`<Header />`, `<Footer />`).
- **Page** (`pages/*/index.pierrot`): the content of a route. May override the
  layout's metadata.
- **Component** (`components/*/index.pierrot`): a reusable piece, with props.

→ Go deeper in [Components & props](./componentes.md).

### Real layout example

```html
<script>
    import "./globals.css";
    import { Home } from "./pages/home/index.pierrot";
    import { Errors } from "./pages/errors/index.pierrot";
    import { Header } from "./components/header/index.pierrot";
    import { Footer } from "./components/footer/index.pierrot";

    set.Title("myapp");
    set.Description("Page description");
    set.Icon("./assets/favicon.ico");
    set.Robots("./assets/robots.txt");
    set.Default(Home);
    set.Fallback(Errors);
</script>

<Header />
<Slot />
<Footer />
```

→ See the [`main.pierrot` from the official site](../www/src/main.pierrot).
