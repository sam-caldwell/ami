package backend

// LinkOptions describe how to link a set of object files into a binary.
type LinkOptions struct {
    // Env is the os/arch pair (e.g., linux/arm64) for layout purposes.
    Env string
    // Triple is the LLVM/target triple used for codegen (e.g., aarch64-unknown-linux-gnu).
    Triple string
    // OutputDir is the base directory for produced binaries (e.g., build/<env>/).
    OutputDir string
    // Name is the desired output binary name (e.g., project name).
    Name string
    // Objects is a list of object file paths to link.
    Objects []string
    // Runtime is a list of runtime object or bitcode paths required to link.
    Runtime []string
    // StripSymbols requests removal of symbols and non-deterministic sections when supported.
    StripSymbols bool
}

// LinkProducts captures outputs of the linker.
type LinkProducts struct {
    // Binary is the path to the produced executable (workspace-relative).
    Binary string
}

// Linker is implemented by backends that can link objects for a target.
type Linker interface {
    Link(opts LinkOptions) (LinkProducts, error)
}

// NoopLinker is a stub that performs no linking and returns an empty product.
type NoopLinker struct{}

func (NoopLinker) Link(opts LinkOptions) (LinkProducts, error) { return LinkProducts{}, nil }

