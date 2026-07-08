# Reactivity

[← back to index](./arquitetura.md)

Pierrot has no virtual DOM and no `eval`. State is just the top-level variables of
the `<script>`. Every interaction calls `__pierrotUpdate()` at the end, and the
runtime repaints only the bits that depend on the state:

- `${var}` spans get the new value;
- `@for`/`@if` blocks re-run;
- inputs with `@bind` get the value (except the focused one).

This page covers `@event` and `@bind`. Interpolation `${}` and blocks
are in [Template syntax](./templates.md).

---

## `@event={fn}` — events

Any DOM event becomes an `@<event>` directive:

```html
<script>
    let count: number = 0;
    function increment() { count++; }
</script>

<button @click={increment}>+1</button>
<p>${count}</p>
```

`@click={increment}` becomes `onclick="increment(); __pierrotUpdate()"`. After the
handler runs, the `__pierrotUpdate()` repaints `${count}`.

Works with any event: `@input`, `@submit`, `@mouseover`, `@scroll`, etc.

### With arguments

```html
<button @click={Click(command)}>copy</button>
```

- `@click={fn}` → `fn()`
- `@click={fn(args)}` → `fn(args)`

In a component, passing a **prop** as an argument is the way to get it into the
logic (see [props in the script](./componentes.md#props-in-the-components-script)).

---

## `@bind={var}` — two-way binding

Links an `<input>` or `<textarea>` to a variable both ways:

```html
<script>
    let name: string = "";
</script>

<input @bind={name} placeholder="your name" />
<p>Hello, ${name}!</p>
```

- Typing in the input updates `name` and fires `__pierrotUpdate()`.
- Changing `name` in code updates the input's value.
- The **focused** element is not overwritten while you type.

→ Real example (live code editor):
[`code-example`](../www/src/components/code-example/index.pierrot).

---

> Raw HTML, Markdown rendering, and similar cases are not part of the language
> core. Pierrot's compiler stays deliberately small; resolve those needs with
> a library or an external API instead.

## How it all connects

```html
<script>
    let query: string = "";
    let results = ["alpha", "beta", "gamma"];

    function filtered() {
        return results.filter(r => r.includes(query));
    }
</script>

<input @bind={query} placeholder="filter..." />

<ul>
@for r in filtered()
    <li>${r}</li>
@end
</ul>
```

Typing in the input → `query` changes → `__pierrotUpdate()` → the `@for` block
re-evaluates `filtered()` and re-renders the list. No virtual DOM, no state
management framework.

→ Helpers available in the `<script>` (`get`, `client`, `time`):
[Script API](./script-api.md).
