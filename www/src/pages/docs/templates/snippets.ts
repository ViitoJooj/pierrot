// Code samples for /docs/templates (see note in /docs/snippets.ts).
const snipInterp: string = `<script>
    let name: string = "world";
    let user = { age: 22 };
<\/script>

<h1>Hello, \${name}!</h1>
<p>Age: \${user.age}</p>
<p>In 10 years: \${user.age + 10}</p>`;

const snipComment: string = `// this comment is gone from the final HTML
<p>visible</p>`;

const snipFor: string = `<script>
    let items = [
        { label: "no node_modules", color: "#FFB300" },
        { label: "no virtual DOM", color: "#C8553D" },
        { label: "no eval", color: "#1A1410" },
    ];
<\/script>

<ul>
@for item in items
    <li style="color: \${item.color}">\${item.label}</li>
@end
</ul>`;

const snipIf: string = `<script>
    let status: string = get.Status();
<\/script>

@if status == "404"
    <h1>404 — not found</h1>
    <p>That route doesn't exist.</p>
@else
    <h1>Error \${status}</h1>
    <p>Something broke.</p>
@endif`;

const snipNest: string = `@for col in columns
    <div class="col">
        <h3>\${col.title}</h3>
@for link in col.links
        <a href="\${link.href}">\${link.label}</a>
@end
    </div>
@end`;
