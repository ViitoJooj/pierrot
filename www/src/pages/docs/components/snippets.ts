// Code samples for /docs/components (see note in /docs/snippets.ts).
const snipDefine: string = `<script>
    import "./styles.css";

    // props: let WITH NO value
    let command: string;
    let animation: boolean = false;

    async function Click(cmd: string) {
        navigator.clipboard.writeText(cmd);
        animation = true;
        await time.Sleep` + `(800).msec();
        animation = false;
    }
<\/script>

<div class="cli">
    <p class="command"><span class="prompt">$</span> {command}</p>
    <button @click={Click(command)}>copy</button>

    @if animation == true
        <div class="stars">✦ ✦ ✦</div>
    @endif
</div>`;

const snipUse: string = `<script>
    import { CliComponent } from "../../components/cli-div/index.pierrot";
<\/script>

<CliComponent command="go install github.com/pierrot/cmd/pierrot@latest" />`;

const snipSlot: string = `<Header />
<Slot />     <!-- the rendered page goes here -->
<Footer />`;

const snipProps: string = `<script>
    let children: string;
    let url: string;

    function goTo(url: string) {
        client.Redirect(url).newTab();
    }
<\/script>

<button class="secondary-btn" @click={goTo(url)}>{children}</button>`;

const snipUseProps: string = `<ButtonSecondary url="https://github.com/ViitoJooj/pierrot" children="github ↗" />`;

const snipExprAttr: string = `<!-- ✅ inside @for: \${expr} is evaluated per iteration -->
@for c in cards
    <div style="color: \${c.color}">\${c.title}</div>
@end`;

const snipArg: string = `<!-- the command prop reaches the function as an argument -->
<button @click={Click(command)}>copy</button>`;
