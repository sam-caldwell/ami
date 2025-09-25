package tester

// Case describes a Phase 2 runtime test case for a compiled AMI pipeline.
type Case struct {
    Name        string
    Pipeline    string // pipeline name or entry
    InputJSON   string // input payload serialized as JSON
    ExpectJSON  string // expected output payload as JSON
    ExpectError string // expected runtime error code (optional)
    TimeoutMs   int    // per-case timeout
}

// Result captures outcome for a single case.
type Result struct {
    Name       string
    Status     string // pass|fail|skip
    DurationMs int64
    Error      string // when fail/skip, a description
}

// Runner scaffolds a runtime executor for AMI pipelines.
// Phase 2: will load compiled pipelines and execute deterministically.
type Runner struct{}

func New() *Runner { return &Runner{} }

// Execute runs the provided cases against the named pipeline.
// Phase 2 scaffold: returns skip for each case with a canned reason.
func (r *Runner) Execute(pipeline string, cases []Case) ([]Result, error) {
    res := make([]Result, 0, len(cases))
    for _, c := range cases {
        res = append(res, Result{Name: c.Name, Status: "skip", Error: "runtime disabled"})
    }
    return res, nil
}

