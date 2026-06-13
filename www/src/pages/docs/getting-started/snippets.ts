// Code samples for /docs/getting-started (see note in /docs/snippets.ts).
const snipInstall: string = `go install github.com/pierrot/cmd/pierrot@latest`;
const snipVerify: string = `pierrot --help`;
const snipInit: string = `pierrot init myapp`;
const snipTree: string = `myapp/
├── src/
│   ├── assets/                # robots.txt + favicon.ico
│   ├── components/            # (empty, ready for your components)
│   ├── pages/
│   │   ├── errors/            # fallback page (404)
│   │   └── home/              # default page ("/")
│   ├── globals.css
│   └── main.pierrot           # global layout
└── settings.pierrot.json      # configuration`;
const snipRun: string = `cd myapp
pierrot dev`;
const snipCounter: string = `<script>
    import "./styles.css";
    import "./script.ts";

    let count: number = 0;

    function increment() {
        count++;
    }
<\/script>

<h1>Clicks: \${count}</h1>
<button @click={increment}>+1</button>`;
const snipBuild: string = `pierrot build`;
