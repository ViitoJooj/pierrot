package files

// Arquivo total do CSS global, injetado em cada página renderizada. Define variáveis de tema e estilos globais.
// O conteúdo é mínimo e serve como ponto de partida para o usuário customizar. Ele pode ser expandido com mais variáveis, reset de estilos, etc.
const GlobalCSS = `* {
    padding: 0;
    margin: 0;
    font-family: Inter, sans-serif;
}

:root[data-theme="dark"] {
    --primary-bg-color: ;
    --secondary-bg-color: ;
    --thirdary-bg-color: ;

    --primary-accent-color: ;
    --secondary-accent-color: ;
    --thirdary-accent-color: ;

    --primary-border-color: ;
    --secondary-border-color: ;
    --thirdary-border-color: ;

    --primary-font-color: ;
    --secondary-font-color: ;
    --thirdary-font-color: ;

    --border-radius-sml: ;
    --border-radius-med: ;
    --border-radius-lrg: ;
}

:root[data-theme="light"] {
    --primary-bg-color: ;
    --secondary-bg-color: ;
    --thirdary-bg-color: ;

    --primary-accent-color: ;
    --secondary-accent-color: ;
    --thirdary-accent-color: ;

    --primary-border-color: ;
    --secondary-border-color: ;
    --thirdary-border-color: ;

    --primary-font-color: ;
    --secondary-font-color: ;
    --thirdary-font-color: ;

    --border-radius-sml: ;
    --border-radius-med: ;
    --border-radius-lrg: ;
}`
