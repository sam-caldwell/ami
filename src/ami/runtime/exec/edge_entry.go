package exec

type edgeEntry struct {
    Unit     string `json:"unit"`
    Pipeline string `json:"pipeline"`
    From     string `json:"from"`
    To       string `json:"to"`
}

