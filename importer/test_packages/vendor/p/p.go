package p

type Thing struct {
	value string
}

func (t *Thing) Outer() string {
	return "Outer"
}
