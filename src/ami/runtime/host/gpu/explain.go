package gpu

// Explain formats a deterministic message for GPU stub errors.
func Explain(backend, op string, err error) string {
    msg := "ok"
    switch err {
    case nil:
        msg = "ok"
    case ErrInvalidHandle:
        msg = "invalid handle"
    case ErrUnavailable:
        msg = "backend unavailable"
    case ErrUnimplemented:
        msg = "unimplemented"
    default:
        msg = err.Error()
    }
    return "gpu/" + backend + " " + op + ": " + msg
}

