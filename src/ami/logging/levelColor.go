package logging

// levelColor - return the ANSI color code (string) for a given log level
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
