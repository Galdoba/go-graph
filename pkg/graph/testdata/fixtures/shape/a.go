package shape

// Shape is the type a reference from another package names.
type Shape struct {
	Value int
}

func (s Shape) Read() error { return nil }

type Reader interface {
	Read() error
}

type Box[T any] struct {
	Value T
}

type hidden struct {
	Value int
}

// Shown is an exported name for an entity the package does not export.
type Shown = hidden

// Wrapper reaches Count only through the type it embeds.
type Wrapper struct {
	Inner
}

type Inner struct{}

func (Inner) Count() int { return 0 }

func Top() error { return nil }
