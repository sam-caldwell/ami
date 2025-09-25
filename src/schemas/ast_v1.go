package schemas

import "errors"

type ASTV1 struct {
	Schema    string  `json:"schema"`
	Timestamp string  `json:"timestamp"`
	Package   string  `json:"package"`
	Version   string  `json:"version,omitempty"`
	File      string  `json:"file"`
	Root      ASTNode `json:"root"`
}

type ASTNode struct {
	Kind     string                 `json:"kind"`
	Pos      Position               `json:"pos"`
	EndPos   *Position              `json:"endPos,omitempty"`
	Fields   map[string]interface{} `json:"fields,omitempty"`
	Children []ASTNode              `json:"children,omitempty"`
}

func (a *ASTV1) Validate() error {
	if a == nil {
		return errors.New("nil ast")
	}
	if a.Schema == "" {
		a.Schema = "ast.v1"
	}
	if a.Schema != "ast.v1" {
		return errors.New("invalid schema")
	}
	if a.Root.Kind == "" {
		return errors.New("root.kind required")
	}
	return nil
}
