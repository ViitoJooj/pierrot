# Sintaxe de template

O HTML do `.pierrot` aceita interpolação, eventos e diretivas de bloco. O `$` é o
marcador de variável: só `${...}` é interpolado — chaves soltas `{...}` no texto ficam
como estão.

## Interpolação — `${...}`

```html
<h1>${title}</h1>
<p>Total: ${price * quantity}</p>
<h2>${product.name}</h2>
```

- `${var}` com nome simples vira `<span data-bind="var">` que o runtime preenche e
  atualiza a cada `__pierrotUpdate`.
- `${expr}` com expressão composta (acesso a campo, conta, chamada) vira
  `<span data-pierrot-expr="N">` com a expressão reavaliada a cada update.
- A expressão é JS normal, avaliada no browser com as variáveis do `<script>` em escopo.
- Restrições: a expressão não pode conter `{`, `}` nem quebra de linha.
- Variável não declarada não quebra a página — o binding rende string vazia.
- Dentro de blocos `@for`/`@if` o valor é **escapado** (`&`, `<`, `>`, `"`), então
  interpolação não injeta HTML.

## Eventos — `@evento={fn}`

```html
<button @click={increment}>+1</button>
<input @input={onType} />
```

Vira o atributo nativo `onclick="increment(); __pierrotUpdate()"` — qualquer nome de
evento DOM funciona (`@click`, `@input`, `@change`...). Depois do handler o runtime
re-renderiza todos os bindings e blocos, então mutar uma variável do `<script>` dentro
do handler já reflete na tela.

Restrição: só aceita o **nome** de uma função (`\w+`), sem argumentos nem arrow function.

## Diretivas de bloco — `@for` / `@if`

Diretivas ocupam uma linha própria e fecham com `@end` (ou `@endif`):

```html
@if products.length > 0
    @for product in products
        <div class="product">
            <h2>${product.name}</h2>
            <p>Preço: ${product.price}</p>
            <button @click={buy}>Comprar</button>
        </div>
    @end
@else
    <p>Nenhum produto.</p>
@endif
```

| Diretiva | Gera |
|---|---|
| `@for item in lista` | `for (const item of (lista)) { ... }` |
| `@if cond` | `if (cond) { ... }` |
| `@else` | `} else {` |
| `@end` / `@endif` | fecha o bloco |

Blocos podem ser aninhados. Cada bloco de topo vira um
`<div data-pierrot-block="N" style="display:contents">` no HTML e uma função JS que
gera o conteúdo no browser — re-executada a cada `__pierrotUpdate`. Os dados vivem só
no browser (são os `let` do `<script>`), por isso o servidor não consegue renderizar
o bloco de antemão.

Bloco sem fechamento ou `@end` órfão vira erro no overlay.

## Comentários

Linha começando com `//` (fora do `<script>`) é removida do HTML final:

```html
// este comentário não aparece na página
<h1>${title}</h1>
```

## Reatividade — o modelo mental

Não há observação automática de variáveis. O fluxo é:

1. A página carrega e `__pierrotUpdate()` roda uma vez, preenchendo spans e blocos.
2. Um `@evento={fn}` dispara: chama `fn()` e depois `__pierrotUpdate()` de novo.
3. O update reconstrói **todos** os bindings, expressões e blocos a partir do estado atual.

Ou seja: mudar variável fora de um `@evento` (ex.: dentro de um `setTimeout`) não
atualiza a tela sozinho — chame `__pierrotUpdate()` manualmente nesse caso.
