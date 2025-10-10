package driver

import "testing"

func Test_parseInlineLetReturn_SimpleForms(t *testing.T) {
    // let x = 5; return x
    if rp, ok := parseInlineLetReturn("let x = 5;\nreturn x"); !ok || rp.kind != retLit || rp.lit != "5" {
        t.Fatalf("let+return literal not parsed: ok=%v rp=%+v", ok, rp)
    }
    // x := ev; return x
    if rp, ok := parseInlineLetReturn("x := ev; return x"); !ok || rp.kind != retEV {
        t.Fatalf("short assign ev not parsed: ok=%v rp=%+v", ok, rp)
    }
    // var y = 2 * 3; return y
    if rp, ok := parseInlineLetReturn("var y = 2 * 3; return y"); !ok || rp.kind != retBinOp || rp.lhs != "2" || rp.rhs != "3" || rp.op != "*" {
        t.Fatalf("var assign binop not parsed: ok=%v rp=%+v", ok, rp)
    }
}

