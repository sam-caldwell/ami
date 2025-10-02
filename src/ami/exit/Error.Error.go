package exit

func (e Error) Error() string { return e.Msg }
