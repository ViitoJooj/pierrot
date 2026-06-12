# Arquitetura interna

Mapa do código Go e do pipeline de compilação. Útil para mexer no framework.

## Mapa de arquivos

```
cmd/main.go                        # entrypoint, chama cli.Execute()
internal/cli/
├── root.go                        # comando raiz (cobra)
├── init.go                        # pierrot init  -> workers.CreateProject
├── dev.go                         # pierrot dev   -> workers.DevServer
└── build.go                       # pierrot build -> workers.Build
internal/readers/
├── settings.go                    # Settings/Project, LoadProject, LoadDotenv
└── parser.go                      # ParsePierrot: .pierrot -> Component
internal/workers/
├── create_project.go              # scaffold do init (templates embutidos)
├── dev_server.go                  # servidor dev + renderPage (coração do framework)
├── template.go                    # compileTemplate: @for/@if e ${expr}
└── build.go                       # build estático + bundleCSS + copyAssets
```

## Pipeline de renderização de uma página

`renderPage` (`dev_server.go`) é usado tanto pelo dev server (a cada request) quanto
pelo build (uma vez por página). Etapas, na ordem:

1. **Parse** — `ParsePierrot` (`parser.go`) separa o `<script>` do HTML e extrai por
   regex: imports de CSS (`cssRe`), de scripts (`tsRe`), de componentes (`compRe`) e
   metadados `set.X(...)` (`metaRe`). O que sobra do script é o código da página.

2. **Expansão** — `renderCtx.render` expande recursivamente: parseia o layout
   (`main.pierrot`), depois a página, depois cada componente importado, substituindo
   `<Nome />` pelo HTML do filho. No caminho acumula (sem duplicar) CSS, chunks de
   script (um por arquivo, para o erro apontar a origem) e metadados. A página entra
   no layout no lugar do `<Slot />` (`slotRe`). Substituições usam
   `ReplaceAllLiteralString` porque o HTML pode conter `${var}`, que
   `ReplaceAllString` interpretaria como referência de grupo de captura.

3. **Validação de tags** — tag capitalizada que sobrou no HTML (`unknownTagRe`) =
   componente usado sem import; vira erro.

4. **Template** — `compileTemplate` (`template.go`):
   - remove comentários `//` de linha (`commentLineRe`);
   - encontra blocos `@for`/`@if` (com contagem de aninhamento), traduz cada um para
     o corpo de uma função JS (`compileBlock`/`emitText`) e troca o bloco no HTML por
     `<div data-pierrot-block="N" style="display:contents">`;
   - dentro dos blocos, `${expr}` (`interpRe`) vira concatenação
     `"..." + __pierrotEsc(expr) + "..."` e `@evento={fn}` vira `onclick` nativo.

5. **Bindings fora de bloco** — sobre o HTML restante, em ordem:
   - `${var}` simples (`bindingRe`, só `\w+`) → `<span data-bind="var">`; os nomes
     viram o objeto `state` do runtime;
   - `@evento={fn}` (`eventRe`) → `on<evento>="fn(); __pierrotUpdate()"`;
   - `${expr}` composto que sobrou (`replaceExprs`) → `<span data-pierrot-expr="N">`.

   A ordem importa: `bindingRe` precisa rodar antes de `replaceExprs` para os nomes
   simples virarem `data-bind` (state) em vez de expressão genérica.

6. **TypeScript → JS** — cada chunk passa pelo esbuild (`api.Transform`, loader TS).
   Chunk com erro fica fora do bundle e o erro vai para o overlay (dev) ou derruba o
   build. Minificação nunca renomeia identificadores de topo
   (`MinifyIdentifiers` desligado), senão o `state` dos bindings quebraria.

7. **Montagem** — o HTML final é `<!DOCTYPE html>` + `<head>` (charset, title,
   description, icon, links de CSS) + `<body>` (HTML + overlay + scripts da página +
   runtime + live reload). Dev linka cada CSS; build linka `/<página>/bundle.css`.

## Runtime JS (gerado por `runtimeJS`, sem eval)

O servidor conhece todos os nomes/expressões usados no template, então gera código
estático:

- `preludeJS` — injetado antes dos scripts da página; define os helpers `get`
  (`get.Path(n)` = n-ésimo segmento do pathname).
- `__pierrotEsc(v)` — escapa `& < > "`.
- `__pierrotBlocks` — array de funções, uma por bloco `@for`/`@if`; cada uma retorna
  o HTML do bloco.
- `__pierrotExprs` — array de funções `() => (expr)`, uma por `${expr}` fora de bloco.
- `__pierrotUpdate()` — monta o `state` com os nomes de `${var}` (com guard
  `typeof x === "undefined"` para variável não declarada virar string vazia) e
  preenche `[data-bind]`, `[data-pierrot-block]` e `[data-pierrot-expr]`. Roda uma vez
  no load e depois de cada `@evento`.

## Dev server (`dev_server.go`)

- Handler `/`: URL com extensão → arquivo estático de `src/`; sem extensão → parseia
  e renderiza `pages/<rota>/index.pierrot` na hora. Rota vazia → `defaultPage`
  (resolve `set.Default`).
- Live reload: `watchFiles` tira um snapshot de mtimes (`.pierrot`, `.css`, `.ts`,
  `.js`, `.html`) a cada 300ms; mudou → `notifyReload` avisa os browsers conectados
  via SSE (`/__pierrot/events`) e o `reloadJS` injetado faz `location.reload()`.
- Erros não fatais (componente faltando, TS inválido, bloco sem `@end`...) acumulam
  em `renderCtx.errs` e viram o overlay (`overlayHTML`) — a página continua servindo.

## Build (`build.go`)

- Apaga o `outDir`, varre `pages/**/index.pierrot`, renderiza cada página com
  `dev=false` (erro = fatal, minify conforme settings).
- `bundleCSS`: concatena os CSS na ordem de descoberta e minifica via esbuild — um
  arquivo por página.
- Copia a página default para `outDir/index.html`.
- `copyAssets`: copia tudo que a página referencia por URL (imagens, fontes...),
  pulando fontes do framework (`.pierrot`, `.ts`, `.css` soltos, settings).

## Regexes centrais (referência rápida)

| Nome | Arquivo | Padrão | Pega |
|---|---|---|---|
| `scriptRe` | parser.go | `(?s)<script>(.*?)</script>` | bloco script |
| `cssRe` | parser.go | `import "x.css";` | import de CSS |
| `tsRe` | parser.go | `import "x.ts/js";` | import de script |
| `compRe` | parser.go | `import { N } from "x.pierrot";` | import de componente |
| `metaRe` | parser.go | `set.X("..."` ou `set.X(Ident)` | metadados |
| `bindingRe` | dev_server.go | `\$\{(\w+)\}` | `${var}` simples |
| `eventRe` | dev_server.go | `@(\w+)=\{(\w+)\}` | `@click={fn}` |
| `slotRe` | dev_server.go | `<Slot\s*/>` | slot do layout |
| `unknownTagRe` | dev_server.go | `<([A-Z]\w*)[^>]*/?>` | componente sem import |
| `interpRe` | template.go | `\$\{([^{}\n]+)\}` | `${expr}` (inclusive composto) |
| `forLineRe` / `ifLineRe` / `elseLineRe` / `endLineRe` | template.go | linhas `@for/@if/@else/@end` | diretivas |
| `commentLineRe` | template.go | linha `//...` | comentários |
