# CLI & configuration

[← back to index](./arquitetura.md)

## Commands

```bash
pierrot init <name>     # create a project
pierrot dev             # development server
pierrot build           # generate the static site
pierrot vscode install  # install the VS Code extension
```

---

### `pierrot init <name>`

Creates a `<name>/` folder with the minimal scaffold: `settings.pierrot.json`,
`src/main.pierrot` (layout), a `home` page (default) and an `errors` page
(fallback), plus `globals.css` and `assets/` (`robots.txt` + `favicon.ico`).

```bash
pierrot init myapp
```

→ Generated layout: [Project structure](./estrutura-do-projeto.md).

---

### `pierrot dev`

Starts the development server (default <http://localhost:3000>). Must be run from
inside the project (where `settings.pierrot.json` lives).

```bash
cd myapp
pierrot dev
```

- **Compiles on every request** — no cache, no HMR to go wrong.
- **Live reload over SSE**: the server watches files (`.pierrot`, `.css`, `.ts`,
  `.js`, `.html`) and reloads the browser when something changes.
- **Error overlay**: template or TS transform errors show up in a popup over the
  page, without taking the server down. Fix it, and it reloads on its own.
- Files with an extension (CSS, images) are served straight from `src/`;
  `/robots.txt` comes from the file pointed to by `set.Robots`.

---

### `pierrot build`

Renders every page under `src/pages/` to static HTML in the `outDir`.

```bash
cd myapp
pierrot build
```

What gets generated:

| Output | Source |
|--------|--------|
| `<outDir>/<route>/index.html` | each `pages/<route>/index.pierrot` |
| `<outDir>/<route>/bundle.css` | `globals.css` + page CSS + component CSS, minified |
| `<outDir>/index.html` | copy of the `set.Default` page (route `/`) |
| `<outDir>/404.html` | the `set.Fallback` page rendered with status 404 |
| `<outDir>/robots.txt` | the `set.Robots` file |
| other assets | copied from `src/` preserving the structure |

- CSS and JS are **minified** (configurable).
- **Any error aborts the build** (unlike dev, which shows it in the overlay).

---

### `pierrot vscode install`

Installs the bundled VS Code extension (`.pierrot` syntax highlighting, snippets
and icons) using the `code` CLI.

```bash
pierrot vscode install
```

Requires the `code` command on your PATH. If it's not there: open VS Code, run
**"Shell Command: Install 'code' command in PATH"** and try again.

---

## `settings.pierrot.json`

Lives at the project root. Controls name, entry, port, dotenv and the build.

```json
{
    "app": {
        "name": "myapp",
        "version": "1.0.0",
        "entry": "./src/main.pierrot",
        "port": 3000
    },
    "dotenv": {
        "enabled": false,
        "path": "./.env"
    },
    "build": {
        "outDir": "../build",
        "minify": true,
        "sourcemap": false
    }
}
```

### Fields

| Field | Default | Description |
|-------|---------|-------------|
| `app.name` | — | project name |
| `app.version` | — | version (metadata) |
| `app.entry` | `./src/main.pierrot` | layout / entry point. **Relative to the settings folder.** Its folder becomes the `src`. |
| `app.port` | `3000` | `pierrot dev` port |
| `dotenv.enabled` | `false` | enables `get.Dotenv(...)` |
| `dotenv.path` | — | `.env` path, relative to `src` |
| `build.outDir` | `./dist` | build output folder, relative to `src` |
| `build.minify` | `true` | minify CSS and JS |
| `build.sourcemap` | `false` | inline sourcemap in the build's JS |

> **Path resolution:** `app.entry` is relative to the `settings.pierrot.json`
> folder. All other paths (`outDir`, `dotenv.path`) are relative to the entry's
> folder — i.e. to `src`. That's why the scaffold's default `outDir` is
> `../build`: from `src/`, that points to the project root.

With no `settings.pierrot.json`, Pierrot assumes the default layout
(`src/main.pierrot`, port 3000, build in `dist/`, minify on).

→ Environment variables: [Script API › `.env`](./script-api.md#environment-variables-env).

---

## Deploy (static site)

`pierrot build` produces pure static HTML — no server, no backend runtime. Upload
the contents of `outDir` to any static host:

- **GitHub Pages / Cloudflare Pages / Netlify / Vercel** — point the publish
  directory to your `outDir`.
- **Any CDN or bucket** (S3, etc).

The `404.html`-at-the-root convention is recognized by most static hosts, so
non-existent routes fall back to your `set.Fallback` page automatically.

→ See a real build in [`www/build/`](../www/build/) (generated from
[`www/src/`](../www/src/)).
