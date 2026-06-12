# Estrutura de projeto e configuração

## Layout de pastas

```
meuapp/
├── settings.pierrot.json     # configuração (opcional, tem defaults)
└── src/
    ├── main.pierrot          # layout global: <Slot /> + set.Default + set.Title...
    ├── globals.css
    ├── assets/
    ├── components/           # componentes reutilizáveis (.pierrot)
    └── pages/
        ├── home/
        │   ├── index.pierrot
        │   ├── style.css
        │   └── script.ts
        └── products/
            └── index.pierrot
```

## Rotas

A rota é o caminho da pasta dentro de `pages/`:

| Arquivo | URL |
|---|---|
| `src/pages/home/index.pierrot` | `/home` |
| `src/pages/products/index.pierrot` | `/products` |
| `src/pages/blog/post/index.pierrot` | `/blog/post` |

A rota `/` é resolvida por `set.Default(X)` no `main.pierrot`: `X` precisa ser um
componente importado de dentro de `pages/`, e a pasta dele vira a página inicial.
Sem `set.Default`, o fallback é `home`.

## `settings.pierrot.json`

Todos os campos são opcionais — sem o arquivo, o projeto funciona com os defaults.

```json
{
    "app": {
        "name": "meuapp",
        "version": "1.0.0",
        "entry": "./src/main.pierrot",
        "port": 3000
    },
    "dotenv": {
        "enabled": false,
        "path": "../.env"
    },
    "build": {
        "outDir": "../build",
        "minify": true,
        "sourcemap": false
    }
}
```

| Campo | Default | Efeito |
|---|---|---|
| `app.entry` | `./src/main.pierrot` (fallback: `./main.pierrot` na raiz) | Caminho do layout global |
| `app.port` | `3000` | Porta do `pierrot dev` |
| `dotenv.enabled` + `dotenv.path` | desligado | Carrega `KEY=VALUE` do arquivo para o ambiente do processo antes de dev/build |
| `build.outDir` | `./dist` | Pasta de saída do `pierrot build` (**apagada a cada build**) |
| `build.minify` | `true` | Minifica JS (whitespace + sintaxe; identificadores não são renomeados) e CSS |
| `build.sourcemap` | `false` | Sourcemap inline no JS do build |

`app.name` e `app.version` são informativos — usados só pelo scaffold, não afetam dev/build.

**Regra de resolução de caminhos** (`internal/readers/settings.go`): `app.entry` é
relativo à pasta do settings; **todos os outros caminhos** (`outDir`, `dotenv.path`)
são relativos à pasta do entry (o `src/` do projeto). Por isso o scaffold usa
`"outDir": "../build"` — sai de `src/` e cai na raiz do projeto.

## Dotenv

Formato simples, uma variável por linha:

```
# comentário
API_URL=https://api.exemplo.com
TOKEN="abc123"
```

Linhas vazias e `#` são ignoradas; aspas duplas em volta do valor são removidas.
As variáveis vão para o ambiente do **processo do pierrot** (Go), não para o JS da página.
