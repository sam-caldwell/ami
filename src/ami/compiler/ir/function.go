package ir

// Function represents a declared function in a module.
// TypeParams surfaces generic type parameter names for tooling/IR schemas.
type Function struct{
    Name       string
    TypeParams []string
}
