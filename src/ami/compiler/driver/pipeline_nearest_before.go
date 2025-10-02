package driver

func nearestOccBeforeOcc(arr []occ, idx int) int {
    best := 0
    bestIdx := -1
    for _, o := range arr {
        if o.stmtIdx <= idx && o.stmtIdx >= bestIdx { bestIdx = o.stmtIdx; best = o.id }
    }
    return best
}

