package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pierrot",
	Short: "Pierrot is a web framework with single-file components",
	Long:  `Pierrot is a web framework with single-file components. It allows you to build web applications using a component-based architecture, where each component is defined in a single file. This makes it easy to manage and organize your code, and allows for better reusability and maintainability.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
