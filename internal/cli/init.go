package cli

import (
	"github.com/ViitoJooj/pierrot/internal/workers"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project",
	Long:  `Initialize a new project with the specified name.`,
	Run:   workers.CreateProject,
}

func init() {
	rootCmd.AddCommand(initCmd)
}
