package delta

import (
	"testing"
)

func TestAddition(t *testing.T) {
	a := `
aaa
ccc
  `
	b := `
aaa
bbb
ccc
  `
	d := Diff(a, b)
	t.Fatal(d.HTML())
}

func TestChange(t *testing.T) {
	a := `
aaa
bbb
ccc
  `
	b := `
aaa
ddd
ccc
  `
	d := Diff(a, b)
	t.Fatal(d.HTML())
}

func TestTranspose(t *testing.T) {
	a := `
aaa
bbb
ccc
  `
	b := `
aaa
ccc
bbb
  `
	d := Diff(a, b)
	t.Fatal(d.HTML())
}
