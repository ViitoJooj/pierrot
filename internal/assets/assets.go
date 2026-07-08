package assets

import _ "embed"

// VSIX is the packaged Pierrot VS Code extension, embedded so the CLI
// can install it without needing the repository checkout.
//
//go:embed pierrot-lang.vsix
var VSIX []byte

// VSIXName is the file name used when writing the embedded extension to disk.
const VSIXName = "pierrot-lang.vsix"
