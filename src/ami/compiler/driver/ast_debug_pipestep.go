package driver

type astPipeStep struct {
    Name  string    `json:"name"`
    Args  []string  `json:"args,omitempty"`
    Attrs []astAttr `json:"attrs,omitempty"`
    Pos   *dbgPos   `json:"pos,omitempty"`
    Kind  string    `json:"kind"`
}

