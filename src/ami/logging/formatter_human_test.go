package logging

import (
    "regexp"
    "strings"
    "testing"
    "time"
)

func TestHumanFormatter_VerbosePrefixesEachLine(t *testing.T) {
    f := HumanFormatter{Verbose: true, Color: false}
    r := Record{
        Timestamp: time.Date(2025, 9, 24, 17, 5, 6, 789000000, time.UTC),
        Level:     LevelWarn,
        Package:   "pkg/mod",
        Message:   "first line\nsecond line",
    }
    out := string(f.Format(r))
    lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
    if len(lines) != 2 {
        t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
    }
    for i, line := range lines {
        if !strings.HasPrefix(line, "2025-09-24T17:05:06.789Z ") {
            t.Fatalf("line %d missing timestamp prefix: %q", i, line)
        }
    }
}

func TestHumanFormatter_NoPrefixWhenNotVerbose(t *testing.T) {
    f := HumanFormatter{Verbose: false, Color: false}
    r := Record{Timestamp: time.Unix(0, 0), Level: LevelInfo, Message: "hello"}
    out := string(f.Format(r))
    if strings.HasPrefix(out, "1970-01-01T00:00:00.000Z ") {
        t.Fatalf("should not prefix timestamp when not verbose: %q", out)
    }
}

func TestHumanFormatter_ColorOnlyWhenEnabled(t *testing.T) {
    r := Record{Timestamp: time.Unix(0, 0), Level: LevelError, Message: "boom"}
    colored := string(HumanFormatter{Verbose: false, Color: true}.Format(r))
    plain := string(HumanFormatter{Verbose: false, Color: false}.Format(r))
    re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
    if !re.MatchString(colored) {
        t.Fatalf("expected color codes in colored output: %q", colored)
    }
    if re.MatchString(plain) {
        t.Fatalf("did not expect color codes in plain output: %q", plain)
    }
}

func TestHumanFormatter_CRLFtoLF(t *testing.T) {
    f := HumanFormatter{Verbose: false, Color: false}
    r := Record{Timestamp: time.Unix(0, 0), Level: LevelInfo, Message: "a\r\nb"}
    out := string(f.Format(r))
    if strings.Contains(out, "\r\n") {
        t.Fatalf("CRLF should be normalized to LF: %q", out)
    }
}

