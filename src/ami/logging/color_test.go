package logging

import "testing"

func TestLevelColor_Mapping(t *testing.T) {
    if levelColor(LevelInfo) != ansiGreen { t.Fatalf("info color") }
    if levelColor(LevelWarn) != ansiYellow { t.Fatalf("warn color") }
    if levelColor(LevelError) != ansiRed { t.Fatalf("error color") }
}

