// doc.ts — helpers shared by every documentation page.
//
// The docs render code samples with `@render html(docHl(src))`. Going through
// `@render html` (instead of writing the snippet straight into the template)
// is deliberate: the Pierrot compiler treats a literal `<script>` and every
// `${...}` in the template as markup it must process, so a raw code sample
// embedded in the HTML would be mangled. As a JavaScript string handed to
// `@render html`, the snippet is opaque to the compiler and only the browser
// ever touches it.
//
// `docEsc` is the plain HTML escaper; `docHl` is the same tokenizer the live
// editor uses (components/code-example), so code blocks across the site share
// one syntax-highlight palette.

// docEsc makes a string safe to drop inside HTML text.
function docEsc(src: string): string {
    return src
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;");
}

// docHl tokenizes a .pierrot / TypeScript snippet and wraps each token in a
// <span class="tok-*">. It returns the inner HTML only — the page wraps it in
// its own <pre><code> so surrounding template whitespace never leaks in.
function docHl(src: string): string {
    const esc = docEsc(src);
    const span = (cls: string, txt: string) => '<span class="' + cls + '">' + txt + "</span>";
    // comment | string | ${}/@directive | keyword | :type | number | tag | fn-name
    const re = /(\/\/[^\n]*)|("(?:[^"\\\n]|\\.)*"|'(?:[^'\\\n]|\\.)*'|`(?:[^`\\]|\\.)*`)|(\$\{[^}\n]*\}|@\w+)|\b(let|const|var|function|return|if|else|for|of|in|import|from|new|typeof|async|await)\b|(:\s*)([A-Za-z_]\w*)|\b(\d+(?:\.\d+)?)\b|(&lt;\/?)([\w-]+)|(\/?&gt;)|([A-Za-z_]\w*)(?=\s*\()/g;
    return esc.replace(re, function (m, com, str, dir, kw, colon, type, num, tagp, tagn, gt, fn) {
        if (com) return span("tok-comment", com);
        if (str) return span("tok-string", str);
        if (dir) return span("tok-directive", dir);
        if (kw) return span("tok-keyword", kw);
        if (type) return colon + span("tok-type", type);
        if (num) return span("tok-number", num);
        if (tagn) return span("tok-punct", tagp) + span("tok-tag", tagn);
        if (gt) return span("tok-punct", gt);
        if (fn) return span("tok-fn", fn);
        return m;
    });
}

// docCode returns a complete, highlighted code block. The whole <pre> is built
// inside this string so no template whitespace leaks into the preformatted
// text. Use it on its own line: `@render html(docCode(snippet));`.
function docCode(src: string): string {
    return '<pre class="doc-code"><code>' + docHl(src) + "</code></pre>";
}

// docCodeFile is docCode with a filename chip above the block.
function docCodeFile(file: string, src: string): string {
    return '<span class="doc-file">' + docEsc(file) + "</span>" + docCode(src);
}
