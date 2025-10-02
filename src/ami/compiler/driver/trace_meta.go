package driver

type traceMeta struct {
    Traceparent string `json:"traceparent"`
    Tracestate  string `json:"tracestate"`
}

