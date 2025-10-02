package driver

type astDecorator struct {
    Name string   `json:"name"`
    Args []string `json:"args,omitempty"`
    Pos  *dbgPos  `json:"pos,omitempty"`
    Kind string   `json:"kind"`
}

