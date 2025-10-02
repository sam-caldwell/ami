package driver

type pipeEdge struct {
    From  string `json:"from"`
    To    string `json:"to"`
    FromID int   `json:"fromId"`
    ToID   int   `json:"toId"`
}

