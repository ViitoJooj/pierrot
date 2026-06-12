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

	files := map[string]string{
		filepath.Join(name, "settings.pierrot.json"):                   fmt.Sprintf(files.ConfJson, name),
		filepath.Join(name, "src", "main.pierrot"):                     fmt.Sprintf(files.MainPierrot, name),
		filepath.Join(name, "src", "globals.css"):                      files.GlobalCSS,
		filepath.Join(name, "src", "pages", "home", "index.pierrot"):   pages.HomePagePierrot,
		filepath.Join(name, "src", "pages", "home", "styles.css"):      pages.HomeCss,
		filepath.Join(name, "src", "pages", "home", "script.ts"):       pages.HomeTS,
		filepath.Join(name, "src", "pages", "errors", "index.pierrot"): pages.ErrorPierrot,
		filepath.Join(name, "src", "pages", "errors", "styles.css"):    pages.ErrorCSS,
		filepath.Join(name, "src", "pages", "errors", "script.ts"):     pages.ErrorTS,
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("projeto %s criado\n\n  cd %s\n  pierrot dev\n", name, name)
}
