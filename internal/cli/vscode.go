package cli

import (
	"github.com/ViitoJooj/pierrot/internal/workers"

	"github.com/spf13/cobra"
)

var vscodeCmd = &cobra.Command{
	Use:   "vscode",
	Short: "Manage the Pierrot VS Code extension",
}

var vscodeInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the Pierrot VS Code extension",
	Long:  `Install the bundled Pierrot VS Code extension (.pierrot syntax highlighting, snippets and icons) using the VS Code CLI.`,
	Run:   workers.InstallVSCodeExtension,
}

func init() {
	vscodeCmd.AddCommand(vscodeInstallCmd)
	rootCmd.AddCommand(vscodeCmd)
}
