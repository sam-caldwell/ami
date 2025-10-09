package driver

type gpuBlock struct {
    family string
    name   string
    source string
    n      int
    grid   [3]int
    tpg    [3]int
    args   string
}
