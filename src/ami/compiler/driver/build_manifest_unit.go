package driver

type bmUnit struct {
    Unit      string `json:"unit"`
    IR        string `json:"ir,omitempty"`
    LLVM      string `json:"llvm,omitempty"`
    RAII      string `json:"raii,omitempty"`
    Pipelines string `json:"pipelines,omitempty"`
    Contracts string `json:"contracts,omitempty"`
    EventMeta string `json:"eventmeta,omitempty"`
    ASM       string `json:"asm,omitempty"`
    AST       string `json:"ast,omitempty"`
    Sources   string `json:"sources,omitempty"`
}

