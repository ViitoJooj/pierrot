# Installation & Hello World in 30s

[← back to index](./arquitetura.md)

## Installation

### Via `go install`

```bash
go install github.com/pierrot/cmd/pierrot@latest
```

The binary lands in `~/.pierrot/bin/pierrot` (`pierrot.exe` on Windows).

### Via release

Download a pre-built binary from the
[releases page](https://github.com/ViitoJooj/pierrot/releases/).

We recommend creating a `.pierrot` folder in your user directory and putting the
binary in `.pierrot/bin`. Then add that folder to your `PATH` so you can call
`pierrot` from anywhere in your terminal.

Verify the install:

```bash
pierrot --help
```

---

## Hello World in 30s

### 1. Create the project

```bash
pierrot init myapp
```

This generates the minimal scaffold:

```txt
myapp/
├── src/
│   ├── assets/                # robots.txt + favicon.ico
│   ├── components/            # (empty, ready for your components)
│   ├── pages/
│   │   ├── errors/            # fallback page (404)
│   │   └── home/              # default page ("/")
│   ├── globals.css
│   └── main.pierrot           # global layout
└── settings.pierrot.json      # configuration
```

→ What each file does: [Project structure](./estrutura-do-projeto.md).

### 2. Run the dev server

```bash
cd myapp
pierrot dev
```

Open <http://localhost:3000>. The server recompiles on every request and reloads
the browser by itself over SSE when you save a file (live reload). Compilation
errors show up in an overlay over the page, without taking the server down.

### 3. Edit the page

`src/pages/home/index.pierrot` ships with a clickable counter:

```html
<script>
    import "./styles.css";
    import "./script.ts";

    let count: number = 0;

    function increment() {
        count++;
    }
</script>

<h1>Clicks: ${count}</h1>
<button @click={increment}>+1</button>
```

- `let count` is the state. It's just a top-level variable in the `<script>`.
- `${count}` is interpolated into the HTML and stays reactive.
- `@click={increment}` wires the click to the function; after it runs, Pierrot
  repaints the bits that depend on `count`.

Save the file — the browser reloads on its own.

### 4. Build the static site

```bash
pierrot build
```

Renders every page under `src/pages/` to static HTML in the `outDir` (set in
`settings.pierrot.json`), with one minified `bundle.css` per page. The result is
pure static output — deploy it to any CDN or static host.

→ Details in [CLI & configuration](./cli.md) and
[deploy](./cli.md#deploy-static-site).

---

## Next steps

- [**Project structure**](./estrutura-do-projeto.md) — understand each file.
- [**Template syntax**](./templates.md) — `${}`, `@for`, `@if`.
- [**Components & props**](./componentes.md) — split the UI into components.
- [**Reactivity**](./reatividade.md) — `@bind`, `@event`, `@render`.

### Full example

Pierrot's own site is built in Pierrot. Use it as a reference for a real project,
with a layout, components, props and `@for` blocks:
[`www/src/`](../www/src/).
