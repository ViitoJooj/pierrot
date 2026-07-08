const Module = require("module");
const originalRequire = Module.prototype.require;

// Mock vscode module before requiring extension.js
const vscodeModule = {
    Range: class Range {},
    Diagnostic: class Diagnostic {},
    DiagnosticSeverity: { Warning: 1 },
    DiagnosticTag: { Unnecessary: 1 },
    languages: {
        createDiagnosticCollection: () => ({
            set: () => {},
            delete: () => {},
        }),
    },
    workspace: {
        textDocuments: [],
        onDidOpenTextDocument: () => ({ dispose: () => {} }),
        onDidChangeTextDocument: () => ({ dispose: () => {} }),
        onDidCloseTextDocument: () => ({ dispose: () => {} }),
        onWillSaveTextDocument: () => ({ dispose: () => {} }),
    },
    TextEdit: {
        delete: () => {},
    },
};

Module.prototype.require = function(id) {
    if (id === "vscode") {
        return vscodeModule;
    }
    return originalRequire.call(this, id);
};

const assert = require("node:assert");
const test = require("node:test");
const { findUnusedImports, findScriptRanges, isInScript } = require("./extension.js");

test("detecta import de componente nao usado", () => {
    const text = 'import { Home } from "./pages/home/index.pierrot";\n<p>oi</p>\n';
    const result = findUnusedImports(text);
    assert.strictEqual(result.length, 1);
    assert.strictEqual(result[0].name, "Home");
});

test("nao marca import usado como tag", () => {
    const text = 'import { Home } from "./pages/home/index.pierrot";\n<Home />\n';
    assert.strictEqual(findUnusedImports(text).length, 0);
});

test("nao marca import usado em set.Default", () => {
    const text = 'import { Home } from "./pages/home/index.pierrot";\nset.Default(Home);\n';
    assert.strictEqual(findUnusedImports(text).length, 0);
});

test("sem imports retorna lista vazia", () => {
    assert.strictEqual(findUnusedImports("<p>oi</p>").length, 0);
});

test("findScriptRanges acha um bloco de script", () => {
    const text = "<script>\ncodigo\n</script>\n<p>oi</p>";
    const ranges = findScriptRanges(text);
    assert.strictEqual(ranges.length, 1);
    const [start, end] = ranges[0];
    assert.strictEqual(text.slice(start, end), "\ncodigo\n");
});

test("findScriptRanges acha multiplos blocos", () => {
    const text = "<script>a</script><p>x</p><script>b</script>";
    assert.strictEqual(findScriptRanges(text).length, 2);
});

test("isInScript true dentro do bloco", () => {
    const text = "<script>\ncodigo\n</script>\n<p>oi</p>";
    const offset = text.indexOf("codigo");
    assert.strictEqual(isInScript(text, offset), true);
});

test("isInScript false no template", () => {
    const text = "<script>\ncodigo\n</script>\n<p>oi</p>";
    const offset = text.indexOf("<p>");
    assert.strictEqual(isInScript(text, offset), false);
});

test("isInScript false sem nenhum script no documento", () => {
    assert.strictEqual(isInScript("<p>oi</p>", 3), false);
});
