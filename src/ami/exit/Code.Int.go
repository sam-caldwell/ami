package exit

// Int returns the int value of the exit code.
func (c Code) Int() int {
	return int(c)
}
