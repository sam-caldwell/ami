package driver

type astPipe struct {
    Name  string        `json:"name"`
    Steps []astPipeStep `json:"steps"`
    Pos   *dbgPos       `json:"pos,omitempty"`
    Kind  string        `json:"kind"`
}

