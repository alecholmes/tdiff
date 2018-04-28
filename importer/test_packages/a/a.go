package a

import (
	"p"

	"github.com/alecholmes/tdiff/importer/test_packages/a/aa/aaa"
	"github.com/alecholmes/tdiff/importer/test_packages/b"
)

func Do() {
	// Package p from vendor/p
	// fmt.Println(new(p.Thing).Outer())
	new(p.Thing).Outer()

	// Package p from a/aa/aaa/vendor/p
	aaa.GetThing().Nested()

	b.Hello()
}
