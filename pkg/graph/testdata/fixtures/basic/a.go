package basic

type Struct1 struct {
	Field int
}

func Read() error {
	local := 1
	_ = local
	return nil
}

var GlobalCounter int

const MaxRetries = 3
