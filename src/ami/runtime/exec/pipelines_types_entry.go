package exec

type _pipeEntry struct {
    Name  string      `json:"name"`
    Steps []_pipeStep `json:"steps"`
}

