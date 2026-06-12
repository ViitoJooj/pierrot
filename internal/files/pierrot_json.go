package files

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
