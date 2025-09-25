package metrics

import (
    lg "github.com/sam-caldwell/ami/src/internal/logger"
)

// PipelineMetrics represents snapshot counters for a pipeline.
type PipelineMetrics struct {
    Pipeline   string
    QueueDepth int
    Throughput float64 // events/sec
    LatencyMs  int     // average or p50 in milliseconds
    Errors     int
}

// NodeMetrics represents snapshot counters for a specific node in a pipeline.
type NodeMetrics struct {
    Pipeline   string
    Node       string
    QueueDepth int
    Throughput float64 // events/sec
    LatencyMs  int     // average or p50 in milliseconds
    Errors     int
}

// Emit logs metrics as a diag.v1-compatible JSON record via the shared logger.
func (m PipelineMetrics) Emit() {
    lg.Info("metrics.pipeline", map[string]interface{}{
        "pipeline":   m.Pipeline,
        "queueDepth": m.QueueDepth,
        "throughput": m.Throughput,
        "latencyMs":  m.LatencyMs,
        "errors":     m.Errors,
    })
}

// Emit logs metrics as a diag.v1-compatible JSON record via the shared logger.
func (m NodeMetrics) Emit() {
    lg.Info("metrics.node", map[string]interface{}{
        "pipeline":   m.Pipeline,
        "node":       m.Node,
        "queueDepth": m.QueueDepth,
        "throughput": m.Throughput,
        "latencyMs":  m.LatencyMs,
        "errors":     m.Errors,
    })
}

