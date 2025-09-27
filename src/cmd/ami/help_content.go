package main

import _ "embed"

// helpDoc contains the end-user help content embedded into the binary.
//go:embed helpdata/README.md
var helpDoc string

// getHelpDoc returns the embedded help documentation content.
func getHelpDoc() string { return helpDoc }
