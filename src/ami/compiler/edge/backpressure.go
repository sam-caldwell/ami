package edge

// BackpressurePolicy declares how producers are throttled when buffers fill.
type BackpressurePolicy string

const (
    // BackpressureBlock blocks producers until capacity is available.
    BackpressureBlock BackpressurePolicy = "block"
    // BackpressureDrop drops newest/oldest based on edge policy.
    BackpressureDrop BackpressurePolicy = "drop"
)

