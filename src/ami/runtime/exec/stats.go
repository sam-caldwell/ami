package exec

import rmerge "github.com/sam-caldwell/ami/src/ami/runtime/merge"

// StageInfo identifies a stage in the simulated pipeline.
type StageInfo struct{ Name, Kind string; Index int }

// StageStats pairs stage info with counters.
type StageStats struct{ Stage StageInfo; Stats rmerge.Stats }

