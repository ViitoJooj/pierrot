package workers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ViitoJooj/pierrot/internal/assets"

	"github.com/spf13/cobra"
)

// InstallVSCodeExtension writes the embedded .vsix to a temporary file and
// installs it through the VS Code CLI, so it survives VS Code updates.
func InstallVSCodeExtension(cmd *cobra.Command, args []string) {
	codeBin, err := exec.LookPath("code")
	if err != nil {
		fmt.Println("Could not find the 'code' command in PATH.")
		fmt.Println("Open VS Code, run \"Shell Command: Install 'code' command in PATH\" and try again.")
		os.Exit(1)
	}

	tmpDir, err := os.MkdirTemp("", "pierrot-vsix-")
	if err != nil {
		fmt.Println("Failed to create temporary directory:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	vsixPath := filepath.Join(tmpDir, assets.VSIXName)
	if err := os.WriteFile(vsixPath, assets.VSIX, 0o644); err != nil {
		fmt.Println("Failed to write extension file:", err)
		os.Exit(1)
	}

	install := exec.Command(codeBin, "--install-extension", vsixPath, "--force")
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr
	if err := install.Run(); err != nil {
		fmt.Println("Failed to install the extension:", err)
		os.Exit(1)
	}

	fmt.Println("Pierrot extension installed. Reload VS Code to activate it.")
}
