package driver

// lowerState holds per-function lowering state.
type lowerState struct {
    temp int
    varTypes map[string]string
    funcResults map[string][]string
    funcParams  map[string][]string
    funcParamNames map[string][]string
    currentFn string
    // methodRecv caches synthesized receiver temporaries for method-form calls
    methodRecv map[string]irValue
}

// small wrapper to avoid import cycle in state.go
type irValue struct{ id, typ string }
