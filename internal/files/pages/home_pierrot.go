package pages

const HomePagePierrot = `<script>
    import "./styles.css";
    import "./script.ts";

    function Click() {
        let count: number = sum();
    }
</script>

<p>${count}</p>
<button @click={Click}>Click me</button>`
