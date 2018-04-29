package aaa

import (
	"p"

	"github.com/alecholmes/tdiff/importer/test_packages/b"
)

// func TestAAA(t *testing.T) {
// 	thing := new(p.Thing)
// 	// fmt.Println(thing.Nested())
// 	thing.Nested()

// 	b.Hello()
// }

func helper() {
	thing := new(p.Thing)
	// fmt.Println(thing.Nested())
	thing.Nested()

	b.Hello()
}
