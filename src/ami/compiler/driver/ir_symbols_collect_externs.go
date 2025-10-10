package driver

import (
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// collectExterns scans IR for operations that imply runtime extern usage and returns a deterministic list
// of extern symbol names (not full LLVM decls).
func collectExterns(m ir.Module) []string {
    seen := map[string]bool{}
    add := func(s string) { if s != "" { seen[s] = true } }
    for _, f := range m.Functions {
        for _, b := range f.Blocks {
            for _, ins := range b.Instr {
                if ex, ok := ins.(ir.Expr); ok {
                    op := ex.Op
                    if op == "panic" { add("ami_rt_panic") }
                    if op == "alloc" || ex.Callee == "ami_rt_alloc" { add("ami_rt_alloc") }
                    switch ex.Callee {
                    case "ami_rt_zeroize":
                        add("ami_rt_zeroize")
                    case "ami_rt_owned_len":
                        add("ami_rt_owned_len")
                    case "ami_rt_owned_ptr":
                        add("ami_rt_owned_ptr")
                    case "ami_rt_owned_new":
                        add("ami_rt_owned_new")
                    case "ami_rt_zeroize_owned":
                        add("ami_rt_zeroize_owned")
                    case "ami_rt_sleep_ms":
                        add("ami_rt_sleep_ms")
                    case "ami_rt_time_now":
                        add("ami_rt_time_now")
                    case "ami_rt_time_add":
                        add("ami_rt_time_add")
                    case "ami_rt_time_delta":
                        add("ami_rt_time_delta")
                    case "ami_rt_time_unix":
                        add("ami_rt_time_unix")
                    case "ami_rt_time_unix_nano":
                        add("ami_rt_time_unix_nano")
                    case "ami_rt_signal_register":
                        add("ami_rt_signal_register")
                    case "ami_rt_os_signal_enable":
                        add("ami_rt_os_signal_enable")
                    case "ami_rt_os_signal_disable":
                        add("ami_rt_os_signal_disable")
                    case "ami_rt_install_handler_thunk":
                        add("ami_rt_install_handler_thunk")
                    case "ami_rt_get_handler_thunk":
                        add("ami_rt_get_handler_thunk")
                    // bufio runtime shims
                    case "ami_rt_bufio_reader_read":
                        add("ami_rt_bufio_reader_read")
                    case "ami_rt_bufio_reader_peek":
                        add("ami_rt_bufio_reader_peek")
                    case "ami_rt_bufio_reader_unread_byte":
                        add("ami_rt_bufio_reader_unread_byte")
                    case "ami_rt_bufio_writer_write":
                        add("ami_rt_bufio_writer_write")
                    case "ami_rt_bufio_writer_flush":
                        add("ami_rt_bufio_writer_flush")
                    case "ami_rt_bufio_scanner_scan":
                        add("ami_rt_bufio_scanner_scan")
                    case "ami_rt_bufio_scanner_text":
                        add("ami_rt_bufio_scanner_text")
                    case "ami_rt_bufio_scanner_bytes":
                        add("ami_rt_bufio_scanner_bytes")
                    case "ami_rt_bufio_scanner_err":
                        add("ami_rt_bufio_scanner_err")
                    }
                } else if d, ok := ins.(ir.Defer); ok {
                    ex := d.Expr
                    if strings.ToLower(ex.Op) == "call" {
                        switch ex.Callee {
                        case "ami_rt_panic":
                            add("ami_rt_panic")
                        case "ami_rt_alloc":
                            add("ami_rt_alloc")
                        case "ami_rt_zeroize":
                            add("ami_rt_zeroize")
                        case "ami_rt_owned_len":
                            add("ami_rt_owned_len")
                        case "ami_rt_owned_ptr":
                            add("ami_rt_owned_ptr")
                        case "ami_rt_owned_new":
                            add("ami_rt_owned_new")
                        case "ami_rt_zeroize_owned":
                            add("ami_rt_zeroize_owned")
                        case "ami_rt_sleep_ms":
                            add("ami_rt_sleep_ms")
                        case "ami_rt_time_now":
                            add("ami_rt_time_now")
                        case "ami_rt_time_add":
                            add("ami_rt_time_add")
                        case "ami_rt_time_delta":
                            add("ami_rt_time_delta")
                        case "ami_rt_time_unix":
                            add("ami_rt_time_unix")
                        case "ami_rt_time_unix_nano":
                            add("ami_rt_time_unix_nano")
                        case "ami_rt_signal_register":
                            add("ami_rt_signal_register")
                        case "ami_rt_os_signal_enable":
                            add("ami_rt_os_signal_enable")
                        case "ami_rt_os_signal_disable":
                            add("ami_rt_os_signal_disable")
                        case "ami_rt_install_handler_thunk":
                            add("ami_rt_install_handler_thunk")
                        case "ami_rt_get_handler_thunk":
                            add("ami_rt_get_handler_thunk")
                        // bufio runtime shims
                        case "ami_rt_bufio_reader_read":
                            add("ami_rt_bufio_reader_read")
                        case "ami_rt_bufio_reader_peek":
                            add("ami_rt_bufio_reader_peek")
                        case "ami_rt_bufio_reader_unread_byte":
                            add("ami_rt_bufio_reader_unread_byte")
                        case "ami_rt_bufio_writer_write":
                            add("ami_rt_bufio_writer_write")
                        case "ami_rt_bufio_writer_flush":
                            add("ami_rt_bufio_writer_flush")
                        case "ami_rt_bufio_scanner_scan":
                            add("ami_rt_bufio_scanner_scan")
                        case "ami_rt_bufio_scanner_text":
                            add("ami_rt_bufio_scanner_text")
                        case "ami_rt_bufio_scanner_bytes":
                            add("ami_rt_bufio_scanner_bytes")
                        case "ami_rt_bufio_scanner_err":
                            add("ami_rt_bufio_scanner_err")
                        }
                    }
                }
            }
        }
    }
    var out []string
    for k := range seen { out = append(out, k) }
    sortStrings(out)
    return out
}
