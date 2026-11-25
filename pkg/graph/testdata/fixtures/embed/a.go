package embed

type Base struct{}

func (b *Base) Read() error { return nil }

type Outer struct {
	Base
}

func Use(o Outer) error { return o.Read() }
