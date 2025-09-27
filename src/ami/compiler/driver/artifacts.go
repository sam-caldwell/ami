package driver

// Artifacts lists debug outputs produced during compilation.
type Artifacts struct {
    // IR holds paths to written IR JSON debug files (per unit).
    IR []string
}

