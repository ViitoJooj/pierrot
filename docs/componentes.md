# Components & props

[← back to index](./arquitetura.md)

A component is a reusable `.pierrot`. Same anatomy as a page: an optional
`<script>` and a template. The difference is that it declares **props** and is
used as a tag (`<Name />`) inside another `.pierrot`.

---

## Defining a component

`src/components/cli-div/index.pierrot`:

```html
<script>
    import "./styles.css";

    // props: let WITH NO value
    let command: string;
    let animation: boolean = false;

    async function Click(cmd: string) {
        navigator.clipboard.writeText(cmd);
        animation = true;
        await time.Sleep(800).msec();
        animation = false;
    }
</script>

<div class="cli">
    <p class="command"><span class="prompt">$</span> {command}</p>
    <button @click={Click(command)}>copy</button>

    @if animation == true
        <div class="stars">✦ ✦ ✦</div>
    @endif
</div>
```

→ Real component: [`cli-div`](../www/src/components/cli-div/index.pierrot).

---

## Using a component

Import it inside the `<script>` of whoever uses it, then drop the tag in the
template. **Components start with an uppercase letter**; the name in
`import { Name }` is the tag name.

```html
<script>
    import { CliComponent } from "../../components/cli-div/index.pierrot";
</script>

<CliComponent command="go install github.com/ViitoJooj/pierrot/cmd/pierrot@latest" />
```

Rules:

- The tag must be **self-closing**: `<Name ... />`. There's no separate closing
  tag.
- Components can import other components — expansion is recursive.
- If an uppercase tag is left over in the final HTML (with no matching `import`),
  Pierrot reports `tag <Name /> doesn't match any imported component`.

---

## `<Slot />` — where the page goes

The **layout** (`main.pierrot`) uses `<Slot />` to mark where the current route's
page should be injected:

```html
<Header />
<Slot />     <!-- the rendered page goes here -->
<Footer />
```

`<Slot />` is exclusive to the layout (the layout ↔ page relationship). Regular
components receive data via **props**, not via slot.

---

## Props

A prop is declared in the component as `let name: type;` **with no value**.
Whoever uses the component passes the value as a tag attribute.

```html
<!-- btn-secondary component -->
<script>
    let children: string;
    let url: string;

    function goTo(url: string) {
        client.Redirect(url).newTab();
    }
</script>

<button class="secondary-btn" @click={goTo(url)}>{children}</button>
```

```html
<!-- usage -->
<ButtonSecondary url="https://github.com/ViitoJooj/pierrot" children="github ↗" />
```

→ Real component: [`btn-secondary`](../www/src/components/buttons/btn-secondary/index.pierrot).

### Literal vs. expression

| In the usage | Meaning | How it arrives |
|--------------|---------|----------------|
| `title="Hello"` | **literal** | inserted as plain text into the component's HTML |
| `price={value}` | **expression** | becomes `${value}`, evaluated in the browser |

```html
<Card title="Sale" price={item.price} active={item.inStock} />
```

In the component's template, each reference to the prop (`{title}`, `${price}`) is
replaced by the instance's value. A prop declared but **not passed** becomes `""`.

### Props in attributes

A prop as an **expression** inside an attribute only works **inside a block**
`@for`/`@if`:

```html
<!-- ✅ inside @for: ${expr} is evaluated per iteration -->
@for c in cards
    <div style="color: ${c.color}">${c.title}</div>
@end
```

Outside a block, an expression in an attribute would turn into a `<span>` inside
the attribute (broken). Use a **literal** prop there, or move it inside a block.

### Props in the component's `<script>`

A component's `<script>` is **shared** across all instances — that's why a prop
**doesn't** become a variable inside it. The only way a prop reaches the
component's logic is as an **argument** to an `@event`:

```html
<!-- the `command` prop reaches the function as an argument -->
<button @click={Click(command)}>copy</button>
```

> **Prop limits, summarized:**
> 1. A prop doesn't become a variable in `<script>` (pass it as an event argument).
> 2. An expression prop in an attribute only works inside `@for`/`@if`.
> 3. A prop not passed = `""`.

---

## Composition and CSS

- Each component imports its own `styles.css`. CSS is collected **once** (no
  duplicates) even if the component appears multiple times.
- Under `pierrot dev`, each CSS becomes a `<link>`. Under `pierrot build`, they
  all become a single minified `bundle.css` per page.

→ See the official site's full component tree in
[`www/src/components/`](../www/src/components/).

Continue to [**Reactivity →**](./reatividade.md)
