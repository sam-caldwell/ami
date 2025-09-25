package scanner

type Scanner struct {
	src     string
	off     int
	line    int
	column  int
	pending []Comment
}
