package driver

type pipelineOp struct {
    Name      string         `json:"name"`
    ID        int            `json:"id,omitempty"`
    Args      []string       `json:"args,omitempty"`
    Edge      *edgeAttrs     `json:"edge,omitempty"`
    Merge     []pipeMergeAttr `json:"merge,omitempty"`
    MergeNorm *pipeMergeNorm `json:"mergeNorm,omitempty"`
    MultiPath *pipeMultiPath `json:"multipath,omitempty"`
    Attrs     []pipeAttr     `json:"attrs,omitempty"`
}

