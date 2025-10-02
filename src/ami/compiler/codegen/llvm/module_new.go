package llvm

// NewModuleEmitter constructs a new emitter for the given package/unit.
func NewModuleEmitter(pkg, unit string) *ModuleEmitter {
    return &ModuleEmitter{pkg: pkg, unit: unit, triple: DefaultTriple, types: map[string]struct{}{}}
}

