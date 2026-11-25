package attrs

type Counter struct {
	Value int
	Named Struct1
}

type Struct1 struct{}

type Reader interface {
	Read() error
}

func Use(r Reader) error {
	local := 0
	_ = local
	return r.Read()
}
