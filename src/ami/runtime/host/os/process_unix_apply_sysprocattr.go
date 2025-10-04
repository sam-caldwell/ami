//go:build !windows

package os

import (
    "os/exec"
    "syscall"
)

func applySysProcAttr(c *exec.Cmd) {
    c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

