package root

type amiExpect struct {
    kind      string // "no_errors" | "no_warnings" | "error" | "warn" | "errors_count" | "warnings_count"
    code      string // for error/warn kinds
    countSet  bool
    count     int
    msgSubstr string
}

type amiCase struct {
    name       string
    file       string
    pkg        string
    expects    []amiExpect
    skipReason string
}

type amiRuntimeCase struct {
    name        string
    file        string
    pkg         string
    pipeline    string
    inputJSON   string
    expectJSON  string
    expectError string
    timeoutMs   int
    fixtures    []amiFixture
}

type amiFixture struct{ path, mode string }

