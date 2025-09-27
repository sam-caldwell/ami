package edge

import "fmt"

// Pipeline references another pipeline by name with an optional payload type string.
// Type is a future-facing placeholder until types are wired; it records the declared
// payload type for cross-pipeline checks.
type Pipeline struct {
    Name string
    Type string // optional textual type (e.g., "Event<T>")
}

func (p Pipeline) Kind() Kind { return KindPipeline }

func (p Pipeline) Validate() error {
    if p.Name == "" { return fmt.Errorf("pipeline name required") }
    return nil
}

