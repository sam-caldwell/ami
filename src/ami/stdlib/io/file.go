package io

import (
    stdio "io"
    "errors"
    "os"
    "path/filepath"
    "strings"
    "time"
)

// ErrClosed is returned when an operation occurs on a closed handle.
var ErrClosed = errors.New("io.FHO: handle is closed")

// FHO is a file/stream handle abstraction for deterministic I/O semantics.
// Methods mirror Go's file operations but return (n int, err error) where applicable.
// After Close(), all operations fail with ErrClosed.
type FHO struct {
    f     *os.File
    name  string
    stdio bool // true if wrapping process stdio; Close() will not close os.Stdin/out/err
}

// Open opens an existing file for reading (like os.Open).
func Open(name string) (*FHO, error) {
    f, err := os.Open(name)
    if err != nil { return nil, err }
    return &FHO{f: f, name: f.Name()}, nil
}

// Create creates or truncates a file for writing (like os.Create).
func Create(name string) (*FHO, error) {
    f, err := os.Create(name)
    if err != nil { return nil, err }
    return &FHO{f: f, name: f.Name()}, nil
}

// OpenFile is a general opener (like os.OpenFile) returning an FHO.
func OpenFile(name string, flag int, perm os.FileMode) (*FHO, error) {
    f, err := os.OpenFile(name, flag, perm)
    if err != nil { return nil, err }
    return &FHO{f: f, name: f.Name()}, nil
}

// Close closes the underlying handle; subsequent operations return ErrClosed.
func (h *FHO) Close() error {
    if h.f == nil { return nil }
    var err error
    if !h.stdio { // don't close process stdio
        err = h.f.Close()
    }
    h.f = nil
    return err
}

func (h *FHO) ensure() (*os.File, error) {
    if h == nil || h.f == nil { return nil, ErrClosed }
    return h.f, nil
}

// Read reads into p, returning bytes read and error.
func (h *FHO) Read(p []byte) (int, error) {
    f, err := h.ensure()
    if err != nil { return 0, err }
    return f.Read(p)
}

// ReadBytes is an alias for Read; included for API clarity.
func (h *FHO) ReadBytes(p []byte) (int, error) { return h.Read(p) }

// Write writes p to the file, returning bytes written and error.
func (h *FHO) Write(p []byte) (int, error) {
    f, err := h.ensure()
    if err != nil { return 0, err }
    return f.Write(p)
}

// WriteBytes is an alias for Write; included for API clarity.
func (h *FHO) WriteBytes(p []byte) (int, error) { return h.Write(p) }

// Seek sets the offset for the next Read/Write on file per POSIX whence.
func (h *FHO) Seek(offset int64, whence int) (int64, error) {
    f, err := h.ensure()
    if err != nil { return 0, err }
    return f.Seek(offset, whence)
}

// Pos returns the current file position for read/write operations.
func (h *FHO) Pos() (int64, error) { return h.Seek(0, stdio.SeekCurrent) }

// Length returns the current file length in bytes.
func (h *FHO) Length() (int64, error) {
    f, err := h.ensure()
    if err != nil { return 0, err }
    st, err := f.Stat()
    if err != nil { return 0, err }
    return st.Size(), nil
}

// Truncate truncates the file to size n.
func (h *FHO) Truncate(n int64) error {
    f, err := h.ensure()
    if err != nil { return err }
    return f.Truncate(n)
}

// Flush flushes file buffers to stable storage (fsync).
func (h *FHO) Flush() error {
    f, err := h.ensure()
    if err != nil { return err }
    return f.Sync()
}

// Name returns the last known file name (path). Safe after Close().
func (h *FHO) Name() string {
    if h == nil { return "" }
    if h.name != "" { return h.name }
    if h.f != nil { return h.f.Name() }
    return ""
}

// CreateTemp creates a new temporary file under the system temp dir.
// Optional args: [dir] or [dir, suffix]. The dir is relative to the system temp directory.
func CreateTemp(args ...string) (*FHO, error) {
    base := os.TempDir()
    var rel, suffix string
    switch len(args) {
    case 0:
        // no-op
    case 1:
        rel = args[0]
    case 2:
        rel = args[0]
        suffix = args[1]
    default:
        return nil, errors.New("CreateTemp: invalid arguments; want none, [dir], or [dir,suffix]")
    }
    dir := base
    if strings.TrimSpace(rel) != "" {
        dir = filepath.Join(base, rel)
    }
    if err := os.MkdirAll(dir, 0o755); err != nil { return nil, err }
    pattern := "ami-*" + suffix
    f, err := os.CreateTemp(dir, pattern)
    if err != nil { return nil, err }
    return &FHO{f: f, name: f.Name()}, nil
}

// CreateTempDir creates a unique temporary directory under the system temp dir and returns its path.
func CreateTempDir() (string, error) {
    base := os.TempDir()
    return os.MkdirTemp(base, "ami-*")
}

// FileInfo is a simplified stat structure.
type FileInfo struct {
    Name    string
    Size    int64
    Mode    os.FileMode
    ModTime time.Time
    IsDir   bool
}

// Stat returns file information for fileName.
func Stat(fileName string) (FileInfo, error) {
    st, err := os.Stat(fileName)
    if err != nil { return FileInfo{}, err }
    return FileInfo{
        Name:    st.Name(),
        Size:    st.Size(),
        Mode:    st.Mode(),
        ModTime: st.ModTime(),
        IsDir:   st.IsDir(),
    }, nil
}

// Stdin, Stdout, Stderr are special FHO wrappers for process stdio.
var (
    Stdin  = &FHO{f: os.Stdin, name: "stdin", stdio: true}
    Stdout = &FHO{f: os.Stdout, name: "stdout", stdio: true}
    Stderr = &FHO{f: os.Stderr, name: "stderr", stdio: true}
)
