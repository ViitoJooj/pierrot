const vscode = require("vscode");

// import { Name } from "./path.pierrot";  — mesmo formato aceito pelo compilador (compRe em parser.go)
const COMPONENT_IMPORT_RE = /^[ \t]*import\s*\{\s*(\w+)\s*\}\s*from\s*"(.+?\.pierrot)";?[ \t]*\r?$/gm;

let diagnostics;

/**
 * Encontra imports de componente não utilizados no documento.
 * Um componente conta como usado se o nome aparece como tag (<Nome ... />)
 * ou referenciado em qualquer lugar fora da própria linha de import
 * (ex.: set.Default(Home)).
 *
 * @returns {{ name: string, start: number, end: number }[]}
 *          offsets cobrindo a linha inteira do import (incluindo o \n final)
 */
function findUnusedImports(text) {
    const imports = [];
    let m;
    COMPONENT_IMPORT_RE.lastIndex = 0;
    while ((m = COMPONENT_IMPORT_RE.exec(text)) !== null) {
        imports.push({ name: m[1], start: m.index, end: m.index + m[0].length });
    }
    if (imports.length === 0) return [];

    // texto sem nenhuma linha de import: o que sobrar é template + script
    let rest = "";
    let last = 0;
    for (const imp of imports) {
        rest += text.slice(last, imp.start);
        last = imp.end;
    }
    rest += text.slice(last);

    const unused = [];
    for (const imp of imports) {
        const usedAsTag = new RegExp("<" + imp.name + "\\b").test(rest);
        const usedInScript = new RegExp("\\b" + imp.name + "\\b").test(rest);
        if (usedAsTag || usedInScript) continue;
        // estende até o fim da linha (consome o \n para a remoção não deixar linha vazia)
        let end = imp.end;
        if (text[end] === "\r") end++;
        if (text[end] === "\n") end++;
        unused.push({ name: imp.name, start: imp.start, end });
    }
    return unused;
}

function refreshDiagnostics(doc) {
    if (doc.languageId !== "pierrot") return;
    const text = doc.getText();
    const diags = findUnusedImports(text).map((imp) => {
        const range = new vscode.Range(doc.positionAt(imp.start), doc.positionAt(imp.end));
        const d = new vscode.Diagnostic(
            range,
            `Not used.`,
            vscode.DiagnosticSeverity.Warning
        );
        d.source = "pierrot";
        d.tags = [vscode.DiagnosticTag.Unnecessary];
        return d;
    });
    diagnostics.set(doc.uri, diags);
}

function activate(context) {
    diagnostics = vscode.languages.createDiagnosticCollection("pierrot");
    context.subscriptions.push(diagnostics);

    vscode.workspace.textDocuments.forEach(refreshDiagnostics);

    context.subscriptions.push(
        vscode.workspace.onDidOpenTextDocument(refreshDiagnostics),
        vscode.workspace.onDidChangeTextDocument((e) => refreshDiagnostics(e.document)),
        vscode.workspace.onDidCloseTextDocument((doc) => diagnostics.delete(doc.uri)),

        // remove imports não utilizados ao salvar
        vscode.workspace.onWillSaveTextDocument((e) => {
            if (e.document.languageId !== "pierrot") return;
            const doc = e.document;
            const edits = findUnusedImports(doc.getText()).map((imp) =>
                vscode.TextEdit.delete(new vscode.Range(doc.positionAt(imp.start), doc.positionAt(imp.end)))
            );
            if (edits.length > 0) e.waitUntil(Promise.resolve(edits));
        })
    );
}

function deactivate() {}

module.exports = { activate, deactivate };
