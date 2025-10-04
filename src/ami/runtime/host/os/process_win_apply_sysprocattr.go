//go:build windows

package os

import "os/exec"

func applySysProcAttr(c *exec.Cmd) {}

