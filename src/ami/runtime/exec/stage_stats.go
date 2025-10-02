package exec

import rmerge "github.com/sam-caldwell/ami/src/ami/runtime/merge"

// StageStats pairs stage info with counters.
type StageStats struct{ Stage StageInfo; Stats rmerge.Stats }

