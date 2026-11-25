package methods

type Struct1 struct{}

type Struct2 struct{}

func (s1 *Struct1) Read() error { return nil }

func (s2 *Struct2) Read() error { return nil }

func (s2 *Struct2) ReadLoud(loud bool) error { return nil }
