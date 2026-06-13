// i18n.ts — every UI-chrome translation in one place.
//
// The Pierrot compiler bundles .ts/.js into the page <script> (it does not
// resolve `import ... from "*.json"`), so the translation data lives here as a
// plain JSON-shaped object literal. Components never hard-code `pt ? a : b`
// anymore: they read strings through `uiText()` and prefix internal links with
// `localeBasePath()`.
//
// Locale is derived from the URL: anything under /pt-br is Portuguese, the rest
// is English. See the per-language docs pages under src/pages/(pt-br/)docs.

// `get` is a runtime helper injected by the compiler (preludeJS), not a real
// import — declare it so the editor's TypeScript server stops flagging it.
// `declare` is type-only and erased by esbuild, so nothing reaches the bundle.
declare const get: {
    Path(segment: number): string;
    Status(): string;
    Dotenv(name: string): string;
};

const uiTranslations = {
    en: {
        header: {
            switchLanguageLabel: "PT-BR",
            navigation: [
                { label: "Docs", path: "/docs" },
                { label: "Hello World", path: "/docs/getting-started" },
                { label: "Syntax", path: "/docs/templates" },
                { label: "CLI", path: "/docs/cli" },
            ],
        },
        footer: {
            taglineHtml: "Components in one file.<br>Compiled in Go.",
            madeWith: "made with pierrot, obviously ◆",
            columns: [
                {
                    title: "learn",
                    links: [
                        { label: "hello world in 30s", path: "/docs/getting-started" },
                        { label: "template syntax", path: "/docs/templates" },
                        { label: "components & props", path: "/docs/components" },
                        { label: "CLI & settings", path: "/docs/cli" },
                    ],
                },
                {
                    title: "project",
                    links: [
                        { label: "github ↗", url: "https://github.com/ViitoJooj/pierrot" },
                        { label: "issues ↗", url: "https://github.com/ViitoJooj/pierrot/issues" },
                        { label: "MIT license", url: "https://github.com/ViitoJooj/pierrot/blob/main/LICENSE.md" },
                    ],
                },
            ],
        },
        documentationSidebar: {
            brandSuffix: "docs",
            groups: [
                {
                    title: "Getting started",
                    items: [
                        { label: "Overview", path: "/docs" },
                        { label: "Hello World", path: "/docs/getting-started" },
                        { label: "Project structure", path: "/docs/project-structure" },
                    ],
                },
                {
                    title: "The .pierrot language",
                    items: [
                        { label: "Template syntax", path: "/docs/templates" },
                        { label: "Components & props", path: "/docs/components" },
                        { label: "Reactivity", path: "/docs/reactivity" },
                        { label: "Script API", path: "/docs/script-api" },
                    ],
                },
                {
                    title: "Tooling",
                    items: [
                        { label: "CLI & configuration", path: "/docs/cli" },
                    ],
                },
            ],
        },
        errorPage: {
            kickerPrefix: "— ERROR / ",
            notFoundTitleHtml: "This route <em>does not compile.</em>",
            notFoundDescriptionPrefix: "The page you are looking for does not exist — or has not been written yet. Heading back home in ",
            genericTitleHtml: "Something <em>broke.</em>",
            genericDescriptionPrefix: "Unexpected error. Heading back home in ",
            descriptionSuffix: ".",
            backHomeLabel: "back to home →",
            reportIssueLabel: "report an issue ↗",
        },
    },
    pt: {
        header: {
            switchLanguageLabel: "EN",
            navigation: [
                { label: "Docs", path: "/docs" },
                { label: "Hello World", path: "/docs/getting-started" },
                { label: "Sintaxe", path: "/docs/templates" },
                { label: "CLI", path: "/docs/cli" },
            ],
        },
        footer: {
            taglineHtml: "Componentes em um arquivo.<br>Compilados em Go.",
            madeWith: "feito com pierrot, óbvio ◆",
            columns: [
                {
                    title: "aprenda",
                    links: [
                        { label: "hello world em 30s", path: "/docs/getting-started" },
                        { label: "sintaxe de template", path: "/docs/templates" },
                        { label: "componentes e props", path: "/docs/components" },
                        { label: "CLI e settings", path: "/docs/cli" },
                    ],
                },
                {
                    title: "projeto",
                    links: [
                        { label: "github ↗", url: "https://github.com/ViitoJooj/pierrot" },
                        { label: "issues ↗", url: "https://github.com/ViitoJooj/pierrot/issues" },
                        { label: "licença MIT", url: "https://github.com/ViitoJooj/pierrot/blob/main/LICENSE.md" },
                    ],
                },
            ],
        },
        documentationSidebar: {
            brandSuffix: "docs",
            groups: [
                {
                    title: "Primeiros passos",
                    items: [
                        { label: "Visão geral", path: "/docs" },
                        { label: "Hello World", path: "/docs/getting-started" },
                        { label: "Estrutura do projeto", path: "/docs/project-structure" },
                    ],
                },
                {
                    title: "A linguagem .pierrot",
                    items: [
                        { label: "Sintaxe de template", path: "/docs/templates" },
                        { label: "Componentes e props", path: "/docs/components" },
                        { label: "Reatividade", path: "/docs/reactivity" },
                        { label: "Script API", path: "/docs/script-api" },
                    ],
                },
                {
                    title: "Ferramentas",
                    items: [
                        { label: "CLI e configuração", path: "/docs/cli" },
                    ],
                },
            ],
        },
        errorPage: {
            kickerPrefix: "— ERRO / ",
            notFoundTitleHtml: "Essa rota <em>não compila.</em>",
            notFoundDescriptionPrefix: "A página que você procura não existe — ou ainda não foi escrita. Voltando pra home em ",
            genericTitleHtml: "Algo <em>quebrou.</em>",
            genericDescriptionPrefix: "Erro inesperado. Voltando pra home em ",
            descriptionSuffix: ".",
            backHomeLabel: "voltar pra home →",
            reportIssueLabel: "reportar problema ↗",
        },
    },
};

// activeLocale reads the first URL segment: "/pt-br/..." is Portuguese.
// Return type is the key union so uiTranslations[...] indexes cleanly.
function activeLocale(): "en" | "pt" {
    return get.Path(1) === "pt-br" ? "pt" : "en";
}

// localeBasePath is the prefix internal links need ("" for English,
// "/pt-br" for Portuguese).
function localeBasePath(): string {
    return activeLocale() === "pt" ? "/pt-br" : "";
}

// uiText returns the whole translation table for the current locale.
function uiText() {
    return uiTranslations[activeLocale()];
}
