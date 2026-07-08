# Template syntax

[← back to index](./arquitetura.md)

The template is everything after the `<script>` in a `.pierrot`. It's plain HTML
plus a handful of directives: `${}` interpolation, comments, and the `@for` and
`@if` control blocks.

- Reactivity (`@bind`, `@event`) has its own page: [Reactivity](./reatividade.md).
- Props in components: [Components & props](./componentes.md).

---

## Interpolation `${...}`

Put any JavaScript expression between `${` and `}`. The value is **escaped**
(HTML-safe) and stays reactive: when the state changes, the spot is repainted.

```html
<script>
    let name: string = "world";
    let user = { age: 22 };
</script>

<h1>Hello, ${name}!</h1>
<p>Age: ${user.age}</p>
<p>In 10 years: ${user.age + 10}</p>
```

- `${name}` (plain identifier) becomes a `<span data-bind="name">`.
- `${user.age}` / `${user.age + 10}` (composite expression) becomes a
  `<span data-pierrot-expr>` re-evaluated on every update.

An **undeclared** variable becomes an empty string instead of breaking the page
(a `typeof` guard). Handy in components where a prop may not have been passed.

> Interpolation does **not** run in attributes **outside** a block. Inside
> `@for`/`@if` it works fine — see [props in attributes](./componentes.md#props-in-attributes).

---

## Comments

A line starting with `//` (at the first character, no indentation) is removed from
the output:

```html
// this comment is gone from the final HTML
<p>visible</p>
```

An **indented** `//` is treated as plain text. To comment HTML in the middle of
markup, use a normal HTML comment `<!-- ... -->` (which stays in the output).

---

## `@for` — repetition

```html
<script>
    let items = [
        { label: "no node_modules", color: "#FFB300" },
        { label: "no virtual DOM", color: "#C8553D" },
        { label: "no eval", color: "#1A1410" },
    ];
</script>

<ul>
@for item in items
    <li style="color: ${item.color}">${item.label}</li>
@end
</ul>
```

- Syntax: `@for <var> in <expr>` … `@end`.
- Becomes a `for...of` in JavaScript, re-evaluated on every `__pierrotUpdate()`.
  If the list changes, the block re-renders.
- `<expr>` is any JS iterable (array, result of `.filter()`, etc).

> Each directive (`@for`, `@if`) goes on **its own line**. The block content sits
> between the opening line and the `@end`.

---

## `@if` / `@else` — conditional

```html
<script>
    let status: string = get.Status();
</script>

@if status == "404"
    <h1>404 — not found</h1>
    <p>That route doesn't exist.</p>
@else
    <h1>Error ${status}</h1>
    <p>Something broke.</p>
@endif
```

- Syntax: `@if <expr>` … (`@else`) … `@end` **or** `@endif`.
- `<expr>` is any JavaScript condition.
- `@end` and `@endif` are equivalent.

---

## Nesting

`@for` and `@if` can be nested freely. The closing `@end` counts depth, so each
inner block needs its own `@end`:

```html
@for col in columns
    <div class="col">
        <h3>${col.title}</h3>
@for link in col.links
        <a href="${link.href}">${link.label}</a>
@end
    </div>
@end
```

→ See this pattern in the [`footer` of the official site](../www/src/components/footer/index.pierrot).

---

## Common errors

| Symptom | Cause |
|---------|-------|
| `diretiva "@for ..." sem @end/@endif` | block not closed |
| `"@end" sem @for/@if correspondente` | stray `@end`/`@else` |
| `${expr}` shows up literally in an attribute | expression in an attribute outside a block — move it inside a `@for`/`@if` or use a literal prop |
| a variable silently "disappears" | undeclared variable becomes `""` (`typeof` guard) |

Under `pierrot dev`, these errors appear in the overlay at the corner of the page.
Under `pierrot build`, they abort the build.

Continue to [**Components & props →**](./componentes.md)
