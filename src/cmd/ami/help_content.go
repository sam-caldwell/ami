package main

import _ "embed"

// helpDoc contains the end-user help content embedded into the binary.
// Single source of truth for CLI help lives under this package to support go:embed.
//go:embed helpdata/README.md
var helpDoc string

// getHelpDoc returns the embedded help documentation content.
func getHelpDoc() string { return helpDoc }
