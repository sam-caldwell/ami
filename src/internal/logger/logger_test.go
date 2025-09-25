package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

var tsRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z `)

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()
	f()
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func captureStderr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = old }()
	f()
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func TestLogger_Human_TimestampPrefix_Verbose(t *testing.T) {
	t.Cleanup(func() { Setup(false, false, false) })
	Setup(false, true, false)
	out := captureStdout(func() { Info("hello world", nil) })
	if !tsRe.MatchString(out) {
		t.Fatalf("expected ISO8601 UTC millis prefix; got %q", out)
	}
}

func TestLogger_Human_NoTimestamp_WhenNotVerbose(t *testing.T) {
	t.Cleanup(func() { Setup(false, false, false) })
	Setup(false, false, false)
	out := captureStdout(func() { Info("hello world", nil) })
	if tsRe.MatchString(out) {
		t.Fatalf("did not expect timestamp prefix; got %q", out)
	}
}

func TestLogger_Human_Multiline_EachLinePrefixed(t *testing.T) {
	t.Cleanup(func() { Setup(false, false, false) })
	Setup(false, true, false)
	msg := "first line\nsecond line\nthird"
	out := captureStderr(func() { Error(msg, nil) })
	// Check each line has timestamp prefix
	scanner := bufio.NewScanner(strings.NewReader(out))
	count := 0
	for scanner.Scan() {
		count++
		line := scanner.Text()
		if !tsRe.MatchString(line + " ") { // add a space to satisfy regex trailing space
			t.Fatalf("line not prefixed: %q", line)
		}
	}
	if count != 3 {
		t.Fatalf("expected 3 lines, got %d", count)
	}
}

func TestLogger_JSON_ContainsTimestamp(t *testing.T) {
	t.Cleanup(func() { Setup(false, false, false) })
	Setup(true, true, true) // color ignored in JSON
	out := captureStdout(func() { Info("json event", map[string]interface{}{"k": "v"}) })
	// Parse JSON and check timestamp field
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); err != nil {
		t.Fatalf("unmarshal: %v; out=%q", err, out)
	}
	if _, ok := m["timestamp"]; !ok {
		t.Fatalf("missing timestamp field in JSON: %v", m)
	}
	if sch, ok := m["schema"].(string); !ok || sch != "diag.v1" {
		t.Fatalf("unexpected schema: %v", m["schema"])
	}
}
