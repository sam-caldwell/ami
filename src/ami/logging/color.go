package logging

// ANSI color codes used only in human mode when Color=true.
const (
    ansiReset  = "\x1b[0m"
    ansiGray   = "\x1b[90m"
    ansiBlue   = "\x1b[34m"
    ansiGreen  = "\x1b[32m"
    ansiYellow = "\x1b[33m"
    ansiRed    = "\x1b[31m"
    ansiMagenta= "\x1b[35m"
)

func levelColor(l Level) string {
    switch l {
    case LevelTrace:
        return ansiGray
    case LevelDebug:
        return ansiBlue
    case LevelInfo:
        return ansiGreen
    case LevelWarn:
        return ansiYellow
    case LevelError:
        return ansiRed
    case LevelFatal:
        return ansiMagenta
    default:
        return ""
    }
}

