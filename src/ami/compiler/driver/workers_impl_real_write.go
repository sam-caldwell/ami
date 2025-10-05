package driver

import (
    "os"
    "path/filepath"
    "sort"
)

// writeWorkersImplRealC writes build/debug/ir/<pkg>/workers_real.c containing
// symbols ami_worker_impl_<sanitized> for each worker.
// The current implementation returns an error string "unimplemented". Codegen/LLVM
// can later replace the body to call into compiled worker logic.
func writeWorkersImplRealC(pkg string, workers []string) (string, error) {
    if len(workers) == 0 { return "", nil }
    uniq := map[string]struct{}{}
    var list []string
    for _, w := range workers { if w != "" { if _, ok := uniq[w]; !ok { uniq[w] = struct{}{}; list = append(list, w) } } }
    sort.Strings(list)
    var src string
    src += "#include <stdlib.h>\n#include <string.h>\n\n"
    for _, w := range list {
        impl := sanitizeForCSymbol("ami_worker_impl_", w)
        src += "const char* " + impl + "(const char* in_json, int in_len, int* out_len, const char** err) { (void)in_json; (void)in_len; (void)out_len; if (err) *err = strdup(\"unimplemented\"); return NULL; }\n"
    }
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    out := filepath.Join(dir, "workers_real.c")
    if err := os.WriteFile(out, []byte(src), 0o644); err != nil { return "", err }
    return out, nil
}

