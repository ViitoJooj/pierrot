# Script API

[← back to index](./arquitetura.md)

Inside the `<script>` of any `.pierrot` you get, on top of plain TypeScript, four
framework objects: `set` (metadata, at compile time), `get`, `client` and `time`
(runtime in the browser). Plus access to `.env` variables.

---

## `set` — metadata

`set.X(...)` configures the page. It makes sense in the **layout**
(`main.pierrot`) and in **pages** — a page can override what the layout set. The
calls are resolved at compile time and stripped from the script.

| Call | Effect |
|------|--------|
| `set.Title("...")` | `<title>` |
| `set.Description("...")` | `<meta name="description">` |
| `set.Charset("...")` | `<meta charset>` (default `UTF-8`) |
| `set.Icon("./assets/favicon.ico")` | `<link rel="icon">` |
| `set.Robots("./assets/robots.txt")` | file served at `/robots.txt` |
| `set.Default(Component)` | page for the `/` route |
| `set.Fallback(Component)` | page served for non-existent routes (404) |

```html
<script>
    import { Home } from "./pages/home/index.pierrot";
    import { Errors } from "./pages/errors/index.pierrot";

    set.Title("myapp");
    set.Description("Page description");
    set.Icon("./assets/favicon.ico");
    set.Robots("./assets/robots.txt");
    set.Default(Home);
    set.Fallback(Errors);
</script>
```

> `set.Default` and `set.Fallback` take an **imported component**, and its folder
> (relative to `pages/`) becomes the route. The import must point inside
> `pages/`.

---

## `get` — request info

| Call | Returns |
|------|---------|
| `get.Path(n)` | the n-th segment of the pathname, **1-based**. `/about/team` → `get.Path(1) == "about"`, `get.Path(2) == "team"`. Out of range: `""`. |
| `get.Status()` | HTTP status as a string: `"200"`, `"404"`. |
| `get.Dotenv("NAME")` | value of an `.env` variable (see below). |

```html
<script>
    let status: string = get.Status();
</script>

@if status == "404"
    <h1>Page not found</h1>
@endif
```

---

## `client` — navigation

```html
<script>
    function openGithub() {
        client.Redirect("https://github.com/ViitoJooj/pierrot").newTab();
    }

    function goHome() {
        client.Redirect("/");
    }
</script>
```

- `client.Redirect(url)` — navigates to `url`.
- `client.Redirect(url).newTab()` — opens in a new tab instead of navigating.

---

## `time` — async waiting

`time.Sleep(n)` returns a `Promise` that resolves after `n` units. Pick the unit
by chaining:

```html
<script>
    let count: number = 30;

    async function countdown() {
        for (let i = 0; i < 30; i++) {
            await time.Sleep(1).sec();
            count--;
        }
        client.Redirect("/");
    }
</script>

<p>Returning in ${count}s...</p>
```

Available units: `.msec()`, `.sec()`, `.min()`, `.hr()`, `.day()`, `.week()`,
`.month()`, `.year()`.

> The compiler injects the `await` and makes the function `async` automatically
> when it sees `time.Sleep(...)`. After the code following the `await` runs, a
> `__pierrotUpdate()` fires — so state changed post-sleep is already on screen.

→ Real example (404 countdown):
[`pages/errors`](../www/src/pages/errors/index.pierrot).

---

## Environment variables (`.env`)

Enable it in `settings.pierrot.json`:

```json
{
  "dotenv": { "enabled": true, "path": "./.env" }
}
```

Read with `get.Dotenv("NAME")`:

```html
<script>
    let apiUrl: string = get.Dotenv("API_URL");
</script>
```

How it works, and the limits:

- The substitution happens **on the server**, at compile time:
  `get.Dotenv("API_URL")` becomes the literal value before the bundle. It doesn't
  exist at runtime.
- **Only the referenced variables** end up in the HTML — the rest of the `.env`
  never reaches the browser.
- The argument must be a **double-quoted string literal**. A dynamic name
  (`get.Dotenv(variable)`) is an error — the substitution is textual.
- A variable missing from `.env`, or dotenv disabled, becomes `""` + an error in
  the overlay/build.

> ⚠️ Since the value is embedded in the HTML sent to the browser, do **not** put
> secrets (private API keys, passwords) in variables read by `get.Dotenv`. Treat
> anything that goes through it as public.

→ Full configuration in [CLI & configuration](./cli.md#settingspierrotjson).
