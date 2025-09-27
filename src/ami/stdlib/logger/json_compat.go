package logger

import "encoding/json"

func stdJSONUnmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }

