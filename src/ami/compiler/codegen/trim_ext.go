package codegen

import "path/filepath"

func trimExt(path string) string {
    ext := filepath.Ext(path)
    return path[:len(path)-len(ext)]
}

