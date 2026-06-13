package workers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ViitoJooj/pierrot/internal/files"
	"github.com/ViitoJooj/pierrot/internal/files/pages"

	"github.com/spf13/cobra"
)

func CreateProject(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatal("Project name is required")
	}
	name := args[0]

	dirs := []string{
		filepath.Join(name, "src", "pages", "home"),
		filepath.Join(name, "src", "pages", "errors"),
		filepath.Join(name, "src", "components"),
		filepath.Join(name, "src", "assets"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			log.Fatal(err)
		}
	}

	scaffold := map[string][]byte{
		filepath.Join(name, "settings.pierrot.json"):                   fmt.Appendf(nil, files.ConfJson, name),
		filepath.Join(name, "src", "main.pierrot"):                     fmt.Appendf(nil, files.MainPierrot, name),
		filepath.Join(name, "src", "globals.css"):                      []byte(files.GlobalCSS),
		filepath.Join(name, "src", "pages", "home", "index.pierrot"):   []byte(pages.HomePagePierrot),
		filepath.Join(name, "src", "pages", "home", "styles.css"):      []byte(pages.HomeCss),
		filepath.Join(name, "src", "pages", "home", "script.ts"):       []byte(pages.HomeTS),
		filepath.Join(name, "src", "pages", "errors", "index.pierrot"): []byte(pages.ErrorPierrot),
		filepath.Join(name, "src", "pages", "errors", "styles.css"):    []byte(pages.ErrorCSS),
		filepath.Join(name, "src", "pages", "errors", "script.ts"):     []byte(pages.ErrorTS),
		// referenciados pelo set.Robots/set.Icon do main.pierrot do scaffold
		filepath.Join(name, "src", "assets", "robots.txt"):  []byte(files.RobotsTxt),
		filepath.Join(name, "src", "assets", "favicon.ico"): files.Favicon(),
	}
	for path, content := range scaffold {
		if err := os.WriteFile(path, content, 0o644); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("projeto %s criado\n\n  cd %s\n  pierrot dev\n", name, name)
}
