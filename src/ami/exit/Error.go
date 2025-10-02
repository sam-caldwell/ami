package exit

// Error wraps an error with an exit code.
type Error struct {
	Code Code
	Msg  string
}
