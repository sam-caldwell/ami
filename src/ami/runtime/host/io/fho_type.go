package io

import (
    stdio "io"
    "os"
)

// ErrClosed is returned when an operation occurs on a closed handle.
var ErrClosed = os.ErrClosed

// FHO is a file/stream handle abstraction for deterministic I/O semantics.
type FHO struct {
    f     *os.File
    name  string
    stdio bool // true if wrapping process stdio; Close() will not close os.Stdin/out/err
}

// Stdin, Stdout, Stderr are special FHO wrappers for process stdio.
var (
    Stdin  = &FHO{f: os.Stdin, name: "stdin", stdio: true}
    Stdout = &FHO{f: os.Stdout, name: "stdout", stdio: true}
    Stderr = &FHO{f: os.Stderr, name: "stderr", stdio: true}
)

// Close closes the underlying handle; subsequent operations return ErrClosed.
func (h *FHO) Close() error {
    if h.f == nil { return nil }
    var err error
    if !h.stdio { err = h.f.Close() }
    h.f = nil
    return err
}

func (h *FHO) ensure() (*os.File, error) {
    if h == nil || h.f == nil { return nil, ErrClosed }
    if err := guardFS(); err != nil { return nil, err }
    return h.f, nil
}

// Read reads into p, returning bytes read and error.
func (h *FHO) Read(p []byte) (int, error) { f, err := h.ensure(); if err != nil { return 0, err }; return f.Read(p) }
// ReadBytes is an alias for Read.
func (h *FHO) ReadBytes(p []byte) (int, error) { return h.Read(p) }
// Write writes p to the file, returning bytes written and error.
func (h *FHO) Write(p []byte) (int, error) { f, err := h.ensure(); if err != nil { return 0, err }; return f.Write(p) }
// WriteBytes is an alias for Write.
func (h *FHO) WriteBytes(p []byte) (int, error) { return h.Write(p) }
// Seek sets the offset for the next Read/Write on file.
func (h *FHO) Seek(offset int64, whence int) (int64, error) { f, err := h.ensure(); if err != nil { return 0, err }; return f.Seek(offset, whence) }
// Pos returns the current file position.
func (h *FHO) Pos() (int64, error) { return h.Seek(0, stdio.SeekCurrent) }
// Length returns the current file length in bytes.
func (h *FHO) Length() (int64, error) { f, err := h.ensure(); if err != nil { return 0, err }; st, err := f.Stat(); if err != nil { return 0, err }; return st.Size(), nil }
// Truncate truncates the file to size n.
func (h *FHO) Truncate(n int64) error { f, err := h.ensure(); if err != nil { return err }; return f.Truncate(n) }
// Flush flushes buffers.
func (h *FHO) Flush() error { f, err := h.ensure(); if err != nil { return err }; return f.Sync() }
// Name returns the last known file name.
func (h *FHO) Name() string { if h == nil { return "" }; if h.name != "" { return h.name }; if h.f != nil { return h.f.Name() }; return "" }

