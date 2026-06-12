# Componentes `.pierrot`

Um arquivo `.pierrot` tem duas partes: um bloco `<script>` opcional (o primeiro do
arquivo) e o HTML do template — tudo que sobra fora do `<script>`.

```html
<script>
    import "./style.css";
    import "./script.ts";
    import { Header } from "../../components/Header.pierrot";

    set.Title("Minha página");

    let count: number = 0;

    function increment() {
        count++;
    }
</script>

<Header />
<h1>Cliques: ${count}</h1>
<button @click={increment}>+1</button>
```

O parser (`internal/readers/parser.go`) extrai do `<script>`, linha a linha:

## Imports

| Forma | Efeito |
|---|---|
| `import "./style.css";` | CSS linkado no `<head>` da página (dev: um `<link>` por arquivo; build: tudo vira `bundle.css`) |
| `import "./script.ts";` ou `.js` | Conteúdo do arquivo entra no bundle JS da página |
| `import { Nome } from "./x.pierrot";` | Componente — a tag `<Nome />` no template é substituída pelo HTML dele |

Caminhos são relativos ao arquivo que importa. CSS e scripts duplicados (importados
por mais de um componente) entram uma vez só.

## Metadados — `set.X(...)`

Chamadas `set.X("texto")` ou `set.X(Identificador)` no `<script>` são removidas do JS
e guardadas como metadados. O `set.X` da página **sobrescreve** o do layout.

| Chamada | Efeito no `<head>` |
|---|---|
| `set.Title("...")` | `<title>` |
| `set.Description("...")` | `<meta name="description">` |
| `set.Charset("UTF-8")` | `<meta charset>` (default `UTF-8`) |
| `set.Icon("/assets/favicon.ico")` | `<link rel="icon">` |
| `set.Default(Componente)` | Só no `main.pierrot`: define a página da rota `/` (precisa ser um import de `pages/`) |
| `set.Fallback(Componente)` | Só no `main.pierrot`: página servida quando a rota não existe. No dev server responde com status 404; no build vira o `404.html` da raiz. Dentro dela `get.Status()` devolve `"404"` |
| `set.Robots("./assets/robots.txt")` | Só no `main.pierrot`: arquivo servido em `/robots.txt` (copiado para a raiz no build) |

## Código do `<script>`

Tudo que sobra (depois de remover imports e `set.X`) é TypeScript/JavaScript da página,
transpilado pelo esbuild e injetado num `<script>` no final do `<body>`. Variáveis
`let`/`var` de topo ficam visíveis para os bindings `${...}` do template
(ver [template.md](template.md)).

## Helpers do runtime — `get.X(...)`, `client.X(...)`, `time.X(...)`

Diferente de `set.X` (resolvido no servidor), esses helpers são JS que roda no
browser. O runtime injeta os objetos `get`, `client` e `time` antes dos scripts
da página, então dá para usar em declaração de topo:

```js
// URL /about/team
let section = get.Path(1); // "about"
let page    = get.Path(2); // "team"
let status  = get.Status(); // "200" (ou "404" na página de set.Fallback)
```

| Helper | Faz |
|---|---|
| `get.Path(n)` | N-ésimo segmento do `location.pathname`, 1-based; fora do range devolve `""` |
| `get.Status()` | Status HTTP da página como string (`"200"`, `"404"`...) |
| `client.Redirect(url)` | Navega para a url |
| `time.Sleep(n).sec()` | Espera n unidades antes de seguir. Unidades: `msec()`, `sec()`, `min()`, `hr()`, `day()`, `week()`, `month()`, `year()` |

`time.Sleep` parece bloqueante no código, mas por baixo é uma Promise: o compilador
injeta o `await` e torna `async` as funções do arquivo que usam `time.Sleep`. Ao
acordar, os bindings `${...}` são atualizados — dá para fazer contagem regressiva
mudando uma variável de topo dentro de um loop com `time.Sleep`:

```js
let count = 30;

function returning() {
    for (let i = 0; i < 30; i++) {
        time.Sleep(1).sec();
        count--; // ${count} no template atualiza a cada segundo
    }
    client.Redirect("/");
}

returning();
```

## Usando componentes

A tag é **self-closing**: `<Header />`. Tag capitalizada que sobrar no HTML
sem import correspondente vira erro no overlay
(`tag <X /> no HTML não corresponde a nenhum componente importado`).

A expansão é recursiva: um componente pode importar outros componentes, CSS e scripts —
tudo é coletado para a página final.

## Props

O componente declara props como `let nome: tipo;` **sem valor** no `<script>`
(`let` com valor é estado normal). Quem usa passa pelos atributos da tag:
`nome="literal"` ou `nome={expressão}`.

```html
<!-- components/Card.pierrot -->
<script>
    let title: string;
    let price: number;
</script>

<div class="card">
    <h2>{title}</h2>
    <p>${price.toFixed(2)}</p>
</div>
```

```html
<!-- na página -->
<Card title="Product 1" price={19.99} />

@for card in cards
    <Card title={card.title} price={card.price} />
@end
```

No template do filho a prop entra em `{prop}`, `${prop}`, dentro de expressões
(`${price.toFixed(2)}`) e em atributos (`src={image}` — ganha aspas sozinho).

Como funciona: a expansão é textual, por instância. Prop **literal** entra como
texto direto no HTML. Prop **expressão** vira `${expr}` avaliado no browser —
dentro de `@for`/`@if` por iteração (o `card` do loop fica em escopo), fora de
bloco vira um span do runtime. Prop declarada e não passada vira `""`.

Limites:
- Prop não chega no `<script>` do componente — o JS é compartilhado entre as
  instâncias; prop só existe no template.
- Prop expressão em **atributo** só funciona dentro de bloco `@for`/`@if`
  (fora, o valor viraria um `<span>` dentro do atributo). Prop literal em
  atributo funciona em qualquer lugar.

## Layout global — `main.pierrot` e `<Slot />`

O `main.pierrot` (o `app.entry` do settings) envolve toda página. O HTML da página
entra no lugar do `<Slot />`:

```html
<script>
    import "./globals.css";
    import { Home } from "./pages/home/index.pierrot";

    set.Title("meuapp");
    set.Charset("UTF-8");
    set.Default(Home)
</script>

<Slot />
```

O layout pode ter HTML em volta do `<Slot />` (header/footer fixos do site inteiro),
e o CSS/JS dele entra em todas as páginas.

## Limitações atuais

- Um único bloco `<script>` por arquivo (o primeiro; os demais são descartados).
- Props só chegam no template do componente, não no `<script>` dele (ver seção Props).
- `set.X` só aceita uma string literal ou um identificador, sem expressões.
