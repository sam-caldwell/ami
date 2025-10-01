//go:build !windows

package os

import (
    goos "os"
    "os/exec"
    "syscall"
)

func applySysProcAttr(c *exec.Cmd) {
    c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killProcessGroup(p *goos.Process) error {
    // negative pid => kill process group
    return syscall.Kill(-p.Pid, syscall.SIGKILL)
}
