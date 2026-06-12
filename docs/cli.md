# CLI

Binário: `pierrot` (compilado de `cmd/main.go`, comandos registrados em `internal/cli/`).
Todos os comandos rodam a partir do diretório atual — execute-os na raiz do projeto
(onde está o `settings.pierrot.json`).

## `pierrot init <nome>`

Cria um projeto novo na pasta `<nome>` com a estrutura padrão:

```
<nome>/
├── settings.pierrot.json
└── src/
    ├── main.pierrot          # layout global + set.Default
    ├── globals.css
    ├── assets/               # servido como /assets, copiado no build
    ├── components/
    └── pages/
        └── home/
            ├── index.pierrot
            └── style.css
```

Código: `internal/workers/create_project.go`.

## `pierrot dev`

Servidor de desenvolvimento em `http://localhost:<porta>` (porta do settings, padrão 3000).

- **Compila a cada request** — nada de cache; salvar o arquivo e recarregar já mostra a mudança.
- **Live reload** — o browser mantém uma conexão SSE em `/__pierrot/events`; o servidor
  compara mtimes dos arquivos (`.pierrot`, `.css`, `.ts`, `.js`, `.html`) a cada 300ms
  e manda recarregar quando algo muda. `node_modules` e pastas `.x` são ignoradas.
- **Overlay de erros** — erro de template, import quebrado ou erro de TypeScript não
  derruba a página: aparece um popup vermelho no canto inferior listando os erros.
  Corrigir o arquivo recarrega sozinho.
- **Arquivos estáticos** — qualquer URL com extensão (`/globals.css`, `/assets/x.png`)
  é servida direto de `src/`.
- **Rota sem página** — responde 404 com o overlay + live reload.

Código: `internal/workers/dev_server.go`.

## `pierrot build`

Gera o site estático no `outDir` do settings.

1. **Apaga o `outDir` inteiro** antes de começar.
2. Acha toda página (`pages/**/index.pierrot`) e renderiza cada uma em
   `<outDir>/<rota>/index.html`.
3. Concatena e minifica os CSS de cada página (globals + página + componentes) em um
   único `<outDir>/<rota>/bundle.css`.
4. Copia a página de `set.Default(...)` também para `<outDir>/index.html` (rota `/`).
5. Copia assets (imagens, fontes, ícones...) preservando a estrutura de pastas.
   `.pierrot`, `.ts`, `.css` soltos e `settings.pierrot.json` ficam de fora — já foram
   compilados para dentro do HTML/bundle.

Diferenças em relação ao dev: **qualquer erro é fatal** (a página falha e o build
termina com erro), JS/CSS são minificados conforme `build.minify`, e sourcemap inline
é gerado se `build.sourcemap: true`.

Código: `internal/workers/build.go`.
