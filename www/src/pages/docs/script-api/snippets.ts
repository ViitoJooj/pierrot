// Code samples for /docs/script-api (see note in /docs/snippets.ts).
const snipSet: string = `<script>
    import { Home } from "./pages/home/index.pierrot";
    import { Errors } from "./pages/errors/index.pierrot";

    set.Title("myapp");
    set.Description("Page description");
    set.Icon("./assets/favicon.ico");
    set.Robots("./assets/robots.txt");
    set.Default(Home);
    set.Fallback(Errors);
<\/script>`;

const snipStatus: string = `<script>
    let status: string = get.Status();
<\/script>

@if status == "404"
    <h1>Page not found</h1>
@endif`;

const snipClient: string = `<script>
    function openGithub() {
        client.Redirect("https://github.com/ViitoJooj/pierrot").newTab();
    }

    function goHome() {
        client.Redirect("/");
    }
<\/script>`;

const snipTime: string = `<script>
    let count: number = 30;

    async function countdown() {
        for (let i = 0; i < 30; i++) {
            await time.Sleep` + `(1).sec();
            count--;
        }
        client.Redirect("/");
    }
<\/script>

<p>Returning in \${count}s...</p>`;

const snipEnvJson: string = `{
  "dotenv": { "enabled": true, "path": "./.env" }
}`;

const snipEnvRead: string = `<script>
    let apiUrl: string = get.Dotenv` + `("API_URL");
<\/script>`;
