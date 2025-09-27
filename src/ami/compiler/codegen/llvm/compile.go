package llvm

import (
    "errors"
    "os/exec"
)

// CompileLLToObject invokes clang to compile a textual LLVM IR file (.ll) into an object file (.o).
// The target triple should be something like "arm64-apple-macosx".
func CompileLLToObject(clang, llPath, outObj, targetTriple string) error {
    if clang == "" { return errors.New("clang path required") }
    if targetTriple == "" { targetTriple = DefaultTriple }
    // -x ir forces LLVM IR input; -target sets the triple; -c compiles to obj
    cmd := exec.Command(clang, "-x", "ir", "-target", targetTriple, "-c", llPath, "-o", outObj)
    return cmd.Run()
}
