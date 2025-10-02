package main

func isSorted(ss []string) bool {
    for i := 1; i < len(ss); i++ { if ss[i-1] > ss[i] { return false } }
    return true
}

