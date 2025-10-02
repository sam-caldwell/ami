package main

func joinCSV(xs []string) string {
    if len(xs) == 0 { return "" }
    s := xs[0]
    for i := 1; i < len(xs); i++ { s += "," + xs[i] }
    return s
}

