package edge

// Kind enumerates edge spec kinds.
type Kind int

const (
    // KindFIFO is a bounded/unbounded queue with configurable backpressure.
    KindFIFO Kind = iota
    // KindLIFO is a stack with configurable backpressure.
    KindLIFO
    // KindPipeline references another pipeline by name with an expected payload type.
    KindPipeline
    // KindMultiPath configures merge behavior at Collect nodes via merge.* attributes.
    KindMultiPath
)

