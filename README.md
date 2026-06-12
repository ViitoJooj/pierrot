# Pierrot — documentação

Pierrot é um framework web com componentes de arquivo único (`.pierrot`). O CLI em Go
compila os componentes para HTML + JS no servidor; o browser recebe HTML pronto com um
runtime mínimo (sem virtual DOM, sem eval, sem dependência de npm).

## Índice

| Documento | Conteúdo |
|---|---|
| [cli.md](./docs/cli.md) | Comandos `pierrot init`, `pierrot dev`, `pierrot build` |
| [projeto.md](./docs/projeto.md) | Estrutura de pastas, `settings.pierrot.json`, rotas, dotenv |
| [componentes.md](./docs/componentes.md) | Anatomia de um `.pierrot`: `<script>`, imports, `set.X()`, layout e `<Slot />` |
| [template.md](./docs/template.md) | Sintaxe de template: `${var}`, `@for`, `@if`, `@click={fn}`, comentários |
| [arquitetura.md](./docs/arquitetura.md) | Como o compilador funciona por dentro (arquivos Go, pipeline, runtime JS) |

## Visão geral em 30 segundos

```
pierrot init meuapp
cd meuapp
pierrot dev        # http://localhost:3000, recompila a cada request, live reload
pierrot build      # site estático no outDir do settings
```

Cada página é um `src/pages/<rota>/index.pierrot`. O arquivo tem um bloco `<script>`
(imports, metadados e código TS/JS) e o HTML do template embaixo. O `src/main.pierrot`
é o layout global: envolve toda página no lugar do `<Slot />`.

```html
<script>
    import "./style.css";

    let title: string = "Hello, pierrot!";
</script>

<h1>${title}</h1>
```
