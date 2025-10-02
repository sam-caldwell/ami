package driver

type collectEntry struct {
    Unit      string        `json:"unit"`
    Step      string        `json:"step"`
    MultiPath *edgeMultiPath `json:"multipath,omitempty"`
}

