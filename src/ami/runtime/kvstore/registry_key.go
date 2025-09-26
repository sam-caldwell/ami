package kvstore

import "fmt"

func key(pipeline, node string) string { return fmt.Sprintf("%s\x1f%s", pipeline, node) }

