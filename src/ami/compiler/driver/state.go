package driver

// lowerState holds per-function lowering state.
type lowerState struct {
    temp int
    varTypes map[string]string
    funcResults map[string][]string
    funcParams  map[string][]string
}
