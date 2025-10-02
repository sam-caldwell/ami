package io

import "os"

// Stat returns file information for fileName.
func Stat(fileName string) (FileInfo, error) {
    if err := guardFS(); err != nil { return FileInfo{}, err }
    st, err := os.Stat(fileName)
    if err != nil { return FileInfo{}, err }
    return FileInfo{Name: st.Name(), Size: st.Size(), Mode: st.Mode(), ModTime: st.ModTime(), IsDir: st.IsDir()}, nil
}

