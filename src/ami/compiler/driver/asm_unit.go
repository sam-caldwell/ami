package driver

// ASMUnit contains generated assembly text for a compilation unit.
type ASMUnit struct {
    Package string
    Unit    string // file path
    Text    string
}

