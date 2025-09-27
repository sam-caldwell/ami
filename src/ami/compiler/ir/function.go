package ir

// Function is a unit of IR with parameters, results, and blocks.
type Function struct {
    Name    string
    Params  []Value
    Results []Value
    Blocks  []Block
    Decorators []Decorator
}
