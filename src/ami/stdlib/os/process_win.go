//go:build windows

package os

import (
    goos "os"
    "os/exec"
)

func applySysProcAttr(c *exec.Cmd) {}

func killProcessGroup(p *goos.Process) error { return p.Kill() }
