package tester

import (
	"encoding/json"
)

// KVOpt mutates a KVConfig for composing kvstore directives in InputJSON.
type KVOpt func(*KVConfig)

// KVConfig holds kvstore meta directives for test input composition.
type KVConfig struct {
	Pipeline string
	Node     string
	PutKey   string
	PutVal   any
	GetKey   string
	Emit     bool
}

// WithKV sets pipeline and node identifiers for kvstore namespacing.
func WithKV(pipeline, node string) KVOpt {
	return func(c *KVConfig) { c.Pipeline, c.Node = pipeline, node }
}

// KVPut adds a kvstore Put directive.
func KVPut(key string, val any) KVOpt { return func(c *KVConfig) { c.PutKey, c.PutVal = key, val } }

// KVGet adds a kvstore Get directive.
func KVGet(key string) KVOpt { return func(c *KVConfig) { c.GetKey = key } }

// KVEmit enables kvstore metrics emission during the run.
func KVEmit() KVOpt { return func(c *KVConfig) { c.Emit = true } }

// BuildKVInput composes a JSON Input string with kvstore meta and optional payload fields.
// If payload is nil, only kv meta fields are included.
func BuildKVInput(payload map[string]any, opts ...KVOpt) (string, error) {
	c := KVConfig{}
	for _, o := range opts {
		o(&c)
	}
	m := map[string]any{}
	for k, v := range payload {
		m[k] = v
	}
	if c.Pipeline != "" {
		m["kv_pipeline"] = c.Pipeline
	}
	if c.Node != "" {
		m["kv_node"] = c.Node
	}
	if c.PutKey != "" {
		m["kv_put_key"] = c.PutKey
		m["kv_put_val"] = c.PutVal
	}
	if c.GetKey != "" {
		m["kv_get_key"] = c.GetKey
	}
	if c.Emit {
		m["kv_emit"] = true
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
