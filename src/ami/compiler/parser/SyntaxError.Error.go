package parser

func (e SyntaxError) Error() string { return e.Msg }

