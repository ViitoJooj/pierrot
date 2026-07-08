package cli

import (
	"github.com/ViitoJooj/pierrot/internal/workers"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the project as a static site",
	Long:  `Render every page under pages/ to static HTML in dist/, copying css and other assets.`,
	Run:   workers.Build,
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
