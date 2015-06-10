package main

import (
	"bitbucket.org/pancakeio/delta/delta"
	"github.com/gopherjs/gopherjs/js"
)

func main() {
	js.Global.Set("diff", map[string]interface{}{
		"diff": Diff,
		"test": func() *string {
			a := "hello"
			return &a
		},
	})
}

func Diff(a, b string) *js.Object {
	return js.MakeWrapper(delta.Diff(a, b))
}
