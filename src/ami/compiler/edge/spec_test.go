package edge

// Compile-time assertions that concrete specs implement Spec.
var _ Spec = FIFO{}
var _ Spec = LIFO{}
var _ Spec = Pipeline{}
var _ Spec = MultiPath{}

