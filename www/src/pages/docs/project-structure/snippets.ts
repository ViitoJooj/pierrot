// Code samples for /docs/project-structure (see note in /docs/snippets.ts).
const snipTree: string = `myapp/
├── src/
│   ├── assets/                     # static files
│   │   ├── robots.txt              # crawler instructions
│   │   └── favicon.ico             # site icon
│   ├── components/                 # reusable components
│   │   └── header/
│   │       ├── index.pierrot       # template + script
│   │       ├── script.ts           # logic (optional)
│   │       └── styles.css          # styles (optional)
│   ├── pages/                      # application routes
│   │   ├── errors/                 # fallback page (404)
│   │   │   └── index.pierrot
│   │   └── home/                   # "/" route (default page)
│   │       └── index.pierrot
│   ├── globals.css                 # global styles and variables
│   └── main.pierrot                # layout / entry point
└── settings.pierrot.json           # project configuration`;

const snipAnatomy: string = `<script>
    /* 1. imports */
    import "./styles.css";                                  // page/component CSS
    import "./script.ts";                                   // TS/JS bundled in
    import { Card } from "../../components/card/index.pierrot"; // component

    /* 2. metadata (only takes effect in the layout/page) */
    set.Title("My page");

    /* 3. props — let WITH NO value (components only) */
    let title: string;

    /* 4. state and logic */
    let count: number = 0;
    function increment() { count++; }
<\/script>

<!-- 5. template: HTML + directives -->
<h1>\${title}</h1>
<button @click={increment}>\${count}</button>`;

const snipLayout: string = `<script>
    import "./globals.css";
    import { Home } from "./pages/home/index.pierrot";
    import { Errors } from "./pages/errors/index.pierrot";
    import { Header } from "./components/header/index.pierrot";
    import { Footer } from "./components/footer/index.pierrot";

    set.Title("myapp");
    set.Description("Page description");
    set.Icon("./assets/favicon.ico");
    set.Robots("./assets/robots.txt");
    set.Default(Home);
    set.Fallback(Errors);
<\/script>

<Header />
<Slot />
<Footer />`;
