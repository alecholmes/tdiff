package p

type Thing struct {
	Inner int
}

func (t *Thing) Nested() string {
	return "Nested"
}
