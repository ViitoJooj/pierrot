// Code samples for /docs/cli (see note in /docs/snippets.ts).
const snipCmds: string = `pierrot init <name>     # create a project
pierrot dev             # development server
pierrot build           # generate the static site
pierrot vscode install  # install the VS Code extension`;

const snipInit: string = `pierrot init myapp`;

const snipDev: string = `cd myapp
pierrot dev`;

const snipBuild: string = `cd myapp
pierrot build`;

const snipVscode: string = `pierrot vscode install`;

const snipSettings: string = `{
    "app": {
        "name": "myapp",
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
    }
}`;
