package exit

// Code represents a process exit code.
// Keep values stable across the toolchain.
type Code int

const (
    // OK indicates success.
    OK Code = 0

    // Internal indicates an unexpected failure (panic/bug).
    Internal Code = 1

    // User indicates invalid CLI usage or user input.
    User Code = 2

    // IO indicates filesystem or OS interaction errors.
    IO Code = 3

    // Integrity indicates workspace or artifact integrity issues.
    Integrity Code = 4
)

// Int returns the int value of the exit code.
func (c Code) Int() int { return int(c) }

