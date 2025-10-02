package logging

import "testing"

func TestLevelColor(t *testing.T) {
	testData := map[Level]string{
		LevelTrace:         "\x1b[90m",
		LevelDebug:         "\x1b[34m",
		LevelInfo:          "\x1b[32m",
		LevelWarn:          "\x1b[33m",
		LevelError:         "\x1b[31m",
		LevelFatal:         "\x1b[35m",
		Level("bad level"): "",
	}
	for lvl, expected := range testData {
		if actual := levelColor(lvl); actual != expected {
			t.Fatalf("levelColor(%s): expected %s, actual %s", lvl, expected, actual)
		}
	}
}
