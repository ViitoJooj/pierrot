// Code samples for /docs/reactivity (see note in /docs/snippets.ts).
const snipEvent: string = `<script>
    let count: number = 0;
    function increment() { count++; }
<\/script>

<button @click={increment}>+1</button>
<p>\${count}</p>`;

const snipArgs: string = `<button @click={Click(command)}>copy</button>`;

const snipBind: string = `<script>
    let name: string = "";
<\/script>

<input @bind={name} placeholder="your name" />
<p>Hello, \${name}!</p>`;

const snipRenderHtml: string = `<div class="code-highlight">
    @render html(hl(code));
</div>`;

const snipRenderMd: string = `<script>
    let doc: string = "# Title\\n\\nParagraph with **bold** and \`code\`.";
<\/script>

@render markdown(doc);`;

const snipRenderPierrot: string = `<textarea @bind={code}></textarea>
@render pierrot(code);`;

const snipAll: string = `<script>
    let query: string = "";
    let results = ["alpha", "beta", "gamma"];

    function filtered() {
        return results.filter(r => r.includes(query));
    }
<\/script>

<input @bind={query} placeholder="filter..." />

<ul>
@for r in filtered()
    <li>\${r}</li>
@end
</ul>`;
