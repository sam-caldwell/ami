package main

import "github.com/sam-caldwell/ami/src/ami/semver"

// constraintsConflict implements a conservative overlap check for two constraints.
func constraintsConflict(a, b semver.Constraint) bool {
    ba, oka := semver.Bounds(a)
    bb, okb := semver.Bounds(b)
    if !oka || !okb { return false }
    _, ok := semver.Intersect(ba, bb)
    return !ok
}

