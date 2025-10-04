package gpu

// Library represents a Metal library.
type Library struct{
    valid bool
    libId int
}

func (l *Library) Release() error {
    if l == nil || !l.valid { return ErrInvalidHandle }
    if l.libId > 0 { metalReleaseLibrary(l.libId) }
    l.valid = false
    l.libId = 0
    return nil
}

