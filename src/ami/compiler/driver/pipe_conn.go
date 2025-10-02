package driver

type pipeConn struct {
    HasEdges         bool         `json:"hasEdges"`
    IngressToEgress  bool         `json:"ingressToEgress"`
    Disconnected     []string     `json:"disconnected,omitempty"`
    UnreachableFromIngress []string `json:"unreachableFromIngress,omitempty"`
    CannotReachEgress      []string `json:"cannotReachEgress,omitempty"`
    UnreachableFromIngressIDs []pipeNodeRef `json:"unreachableFromIngressIDs,omitempty"`
    CannotReachEgressIDs      []pipeNodeRef `json:"cannotReachEgressIDs,omitempty"`
}

