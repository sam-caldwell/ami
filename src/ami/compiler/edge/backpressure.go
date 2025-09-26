package edge

// BackpressurePolicy declares how producers are throttled when buffers fill.
type BackpressurePolicy string

const (
    // BackpressureBlock blocks producers until capacity is available.
    BackpressureBlock BackpressurePolicy = "block"
    // BackpressureDropOldest drops the oldest item to make room for the new one.
    BackpressureDropOldest BackpressurePolicy = "dropOldest"
    // BackpressureDropNewest drops the incoming item (newest) when full.
    BackpressureDropNewest BackpressurePolicy = "dropNewest"
)
