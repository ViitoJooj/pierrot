package files

// Template do arquivo de configuração JSON criado pelo comando "pierrot init <project-name>".
// Ele é gerado com o nome do projeto e tem as configurações básicas para rodar o servidor de
// desenvolvimento e buildar o projeto. O usuário pode customizar esse arquivo depois da criação,
// mas ele serve como um ponto de partida funcional.
//
// Possiveis alterações:
// - Adicionar mais configurações, como opções de renderização, plugins, etc.
// - Talvez no futuro ele possa ser expandido para suportar múltiplos ambientes (dev, prod, staging) com configurações específicas.
const ConfJson = `{
    "app": {
        "name": "%s",
        "version": "1.0.0",
        "entry": "./src/main.pierrot",
        "port": 3000
    },
    "dotenv": {
        "enabled": false,
        "path": "./.env"
    },
    "build": {
        "outDir": "../build",
        "minify": true,
        "sourcemap": false
    },
    "public": {
        "path": "../public",
        "assets": "./assets"
    }
}`
