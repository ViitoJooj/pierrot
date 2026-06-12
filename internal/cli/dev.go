package cli

import (
	"github.com/ViitoJooj/pierrot/internal/workers"

	"github.com/spf13/cobra"
)

// pierrot dev

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start the development server",
	Long:  `Compile .pierrot pages on each request and serve them at http://localhost:3000.`,
	Run:   workers.DevServer,
}

func init() {
	rootCmd.AddCommand(devCmd)
}
