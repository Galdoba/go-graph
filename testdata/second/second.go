// Package second is the part of the test ground whose declared name is also the last
// segment of its package path, which is the case the printed form states twice.
package second

type Thing struct {
	Value int
}

func (t Thing) Read() int { return t.Value }

func New() Thing { return Thing{} }
