package llvm

import (
    "errors"
    "os/exec"
)

// LinkObjects links the provided object files into an executable binary using clang.
// The target triple should match the objects' triple; when empty, DefaultTriple is used.
// extra allows passing additional compiler/linker flags (e.g., -Wl,-dead_strip).
func LinkObjects(clang string, objs []string, outBin, targetTriple string, extra ...string) error {
    if clang == "" { return errors.New("clang path required") }
    if len(objs) == 0 { return errors.New("no objects to link") }
    if targetTriple == "" { targetTriple = DefaultTriple }
    args := []string{"-target", targetTriple}
    if len(extra) > 0 { args = append(args, extra...) }
    args = append(args, objs...)
    args = append(args, "-o", outBin)
    cmd := exec.Command(clang, args...)
    return cmd.Run()
}
