package files

// Entry point padrão do Pierrot, basicamente todo projeto precisa de um tipo de
// entry point que pode ser customizado diretamente no settings.pierrot.json, porém no
// comando utilizamos esse template como base para criar o main.pierrot do projeto,
//  que é o entry point default. Ele é bem simples, mas já tem as principais tags e
//  configurações básicas, como título, descrição, charset, favicon, robots.txt, página
//  default e página de fallback, ótimo para um bom começo.
//
//  talvez no futuro ele possa ser expandido com mais configurações.
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
