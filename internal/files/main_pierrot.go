package files

const MainPierrot = `<script>
	import "./globals.css";
	import { Home } from "./pages/home/index.pierrot"
	import { Errors } from "./pages/errors/index.pierrot";

	set.Title("%s");
	set.Description("This is my page description");
	set.Charset("UTF-8");
	set.Icon("./assets/favicon.ico");
	set.Robots("./assets/robots.txt");
	set.Default(Home);
	set.Fallback(Errors);
</script>

<Slot />
`
