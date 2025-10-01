package time

import (
    stdtime "time"
    "testing"
)

func TestNowDeltaAddUnix(t *testing.T) {
    t1 := Now()
    Sleep(10 * stdtime.Millisecond)
    t2 := Now()
    if Delta(t1, t2) <= 0 { t.Fatalf("Delta should be >0") }
    t3 := Add(t1, 10*stdtime.Millisecond)
    if t3.Sub(t1) != 10*stdtime.Millisecond { t.Fatalf("Add mismatch") }
    if t1.Unix() <= 0 || t1.UnixNano() <= 0 { t.Fatalf("Unix timestamps should be positive") }
}

func TestTicker_Start_Stop_Register(t *testing.T) {
    var count int
    tk := NewTicker(5 * stdtime.Millisecond)
    tk.Register(func(){ count++ })
    tk.Start()
    Sleep(20 * stdtime.Millisecond)
    tk.Stop()
    if count == 0 { t.Fatalf("ticker did not fire") }
}

