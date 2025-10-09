package driver

import (
    "os"
    "path/filepath"
    "sort"
)

// writeWorkersImplC writes build/debug/ir/<pkg>/workers_impl.c containing wrappers
// for each worker name. Each wrapper uses dlsym(RTLD_DEFAULT) to resolve a real
// implementation symbol named "ami_worker_impl_<sanitized>" and forwards the call
// when available. Otherwise, it returns a malloc'd error string "not implemented".
func writeWorkersImplC(pkg string, workers []string) (string, error) {
    if len(workers) == 0 { return "", nil }
    // dedupe and stable order
    uniq := map[string]struct{}{}
    var list []string
    for _, w := range workers { if w != "" { if _, ok := uniq[w]; !ok { uniq[w] = struct{}{}; list = append(list, w) } } }
    sort.Strings(list)
    // build content
    var src string
    src += "#include <stdlib.h>\n#include <string.h>\n\n"
    src += "#if defined(__APPLE__) || defined(__linux__)\n#include <dlfcn.h>\n#endif\n\n"
    src += "typedef const char* (*ami_worker_impl_fn)(const char*, int, int*, const char**);\n\n"
    for _, w := range list {
        sym := sanitizeForCSymbol("ami_worker_", w)
        impl := sanitizeForCSymbol("ami_worker_impl_", w)
        src += "const char* " + sym + "(const char* in_json, int in_len, int* out_len, const char** err) {\n"
        src += "#if defined(__APPLE__) || defined(__linux__)\n"
        src += "    void* h = RTLD_DEFAULT;\n"
        src += "    const char* name = \"" + impl + "\";\n"
        src += "    void* f = dlsym(h, name);\n"
        src += "    if (f) { ami_worker_impl_fn g = (ami_worker_impl_fn)f; return g(in_json, in_len, out_len, err); }\n"
        src += "    if (err) *err = strdup(\"not implemented\");\n    return NULL;\n"
        src += "#else\n    (void)in_json; (void)in_len; (void)out_len; if (err) *err = strdup(\"not implemented\"); return NULL;\n#endif\n}\n\n"
    }
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    out := filepath.Join(dir, "workers_impl.c")
    if err := os.WriteFile(out, []byte(src), 0o644); err != nil { return "", err }
    return out, nil
}

// sanitizeForCSymbol defined in sanitize_csymbol.go
