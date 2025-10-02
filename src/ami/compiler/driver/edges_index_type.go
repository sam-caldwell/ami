package driver

type edgesIndex struct {
    Schema   string       `json:"schema"`
    Package  string       `json:"package"`
    Edges    []edgeEntry  `json:"edges"`
    Collect  []collectEntry `json:"collect,omitempty"`
}

