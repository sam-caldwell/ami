//go:build windows

package os

import goos "os"

func killProcessGroup(p *goos.Process) error { return p.Kill() }

