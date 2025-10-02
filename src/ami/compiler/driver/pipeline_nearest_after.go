package driver

func nearestOccAfterOcc(arr []occ, idx int) int {
    best := 0
    bestIdx := 1<<30
    for _, o := range arr {
        if o.stmtIdx >= idx && o.stmtIdx <= bestIdx { bestIdx = o.stmtIdx; best = o.id }
    }
    return best
}

