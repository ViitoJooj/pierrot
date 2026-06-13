// Code samples for /docs (overview). Kept in a .ts file on purpose: the
// Pierrot parser scans a page's own <script> for `import`/`set.X`/prop lines,
// so a multi-line sample containing those would be mistaken for real
// directives. Inside an imported .ts the samples are just data.
const snipComponent: string = `<script>
    // imports, metadata (set.X), props and TypeScript logic
    let count: number = 0;

    function increment() {
        count++;
    }
<\/script>

<!-- template: HTML + directives (@for, @if, @event, \${}) -->
<h1>Clicks: \${count}</h1>
<button @click={increment}>+1</button>`;
