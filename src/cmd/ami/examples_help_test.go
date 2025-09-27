package main

import (
    "bytes"
    "testing"
)

func TestBuild_Help_IncludesExamples(t *testing.T) {
    cmd := newBuildCmd()
    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetArgs([]string{"--help"})
    _ = cmd.Execute()
    s := buf.String()
    if !bytes.Contains([]byte(s), []byte("Examples")) || !bytes.Contains([]byte(s), []byte("ami build")) {
        t.Fatalf("build help missing examples; got: %s", s)
    }
}

func TestTest_Help_IncludesExamples(t *testing.T) {
    cmd := newTestCmd()
    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetArgs([]string{"--help"})
    _ = cmd.Execute()
    s := buf.String()
    if !bytes.Contains([]byte(s), []byte("Examples")) || !bytes.Contains([]byte(s), []byte("ami test")) {
        t.Fatalf("test help missing examples; got: %s", s)
    }
}

func TestLint_Help_IncludesExamples(t *testing.T) {
    cmd := newLintCmd()
    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetArgs([]string{"--help"})
    _ = cmd.Execute()
    s := buf.String()
    if !bytes.Contains([]byte(s), []byte("Examples")) || !bytes.Contains([]byte(s), []byte("ami lint")) {
        t.Fatalf("lint help missing examples; got: %s", s)
    }
}

func TestModGet_Help_IncludesExamples(t *testing.T) {
    cmd := newModGetCmd()
    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetArgs([]string{"--help"})
    _ = cmd.Execute()
    s := buf.String()
    if !bytes.Contains([]byte(s), []byte("Examples")) || !bytes.Contains([]byte(s), []byte("ami mod get")) {
        t.Fatalf("mod get help missing examples; got: %s", s)
    }
}

func TestModList_Help_IncludesExamples(t *testing.T) {
    cmd := newModListCmd()
    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetArgs([]string{"--help"})
    _ = cmd.Execute()
    s := buf.String()
    if !bytes.Contains([]byte(s), []byte("Examples")) || !bytes.Contains([]byte(s), []byte("ami mod list")) {
        t.Fatalf("mod list help missing examples; got: %s", s)
    }
}

func TestPipelineVisualize_Help_IncludesExamples(t *testing.T) {
    cmd := newPipelineVisualizeCmd()
    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetArgs([]string{"--help"})
    _ = cmd.Execute()
    s := buf.String()
    if !bytes.Contains([]byte(s), []byte("Examples")) || !bytes.Contains([]byte(s), []byte("ami pipeline visualize")) {
        t.Fatalf("pipeline visualize help missing examples; got: %s", s)
    }
}

