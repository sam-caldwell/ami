package schemas

import "errors"

type IRV1 struct {
	Schema    string       `json:"schema"`
	Timestamp string       `json:"timestamp"`
	Package   string       `json:"package"`
	Version   string       `json:"version,omitempty"`
	File      string       `json:"file"`
	Functions []IRFunction `json:"functions"`
}

type IRFunction struct {
    Name   string    `json:"name"`
    Blocks []IRBlock `json:"blocks"`
    // Optional: parameter metadata with ownership/domain annotations.
    Params []IRParam `json:"params,omitempty"`
    // Optional: generic type parameters with optional constraints (e.g., any)
    TypeParams []IRTypeParam `json:"typeParams,omitempty"`
}

type IRBlock struct {
	Label  string    `json:"label"`
	Instrs []IRInstr `json:"instrs"`
}

type IRInstr struct {
	Op     string        `json:"op"`
	Args   []interface{} `json:"args,omitempty"`
	Result string        `json:"result,omitempty"`
}

// IRParam describes a function parameter in the IR with light annotations
// useful for memory model analysis and tooling.
// - Type: rendered type string, e.g., "Event<string>", "State", "Owned<T>".
// - Ownership: "owned" when parameter type is Owned<…>, otherwise "borrowed".
// - Domain: one of "event", "state", "ephemeral" where applicable.
type IRParam struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Ownership string `json:"ownership,omitempty"`
	Domain    string `json:"domain,omitempty"`
}

// IRTypeParam describes a generic type parameter on a function.
type IRTypeParam struct {
    Name       string `json:"name"`
    Constraint string `json:"constraint,omitempty"`
}

func (i *IRV1) Validate() error {
	if i == nil {
		return errors.New("nil ir")
	}
	if i.Schema == "" {
		i.Schema = "ir.v1"
	}
	if i.Schema != "ir.v1" {
		return errors.New("invalid schema")
	}
	return nil
}
