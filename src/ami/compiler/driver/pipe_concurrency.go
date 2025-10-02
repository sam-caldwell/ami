package driver

type pipeConcurrency struct {
    Workers  int            `json:"workers,omitempty"`
    Schedule string         `json:"schedule,omitempty"`
    Limits   map[string]int `json:"limits,omitempty"`
}

