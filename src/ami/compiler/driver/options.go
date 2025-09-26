package driver

// Options controls compilation driver behaviors and debug artifacts.
type Options struct {
    // SemDiags, when true, includes semantic diagnostics from the analyzer.
    SemDiags bool
    // EffectiveConcurrency, when > 0, is propagated into IR/codegen so
    // generated ASM includes a "; concurrency <n>" header.
    EffectiveConcurrency int
}

