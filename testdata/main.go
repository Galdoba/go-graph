package main

import (
	"fmt"
	str "strings"
)

type Struct1 struct {
	Field1 int
	Field2 string
	Field3 func(int) int
	Reader
}

// Alias of a declared type.
type Mirror = Struct1

type Name string

// Alias of Name.
type NameAlias = Name

// Alias of a predeclared identifier, which names an entity that is not a node.
type Counter = int

type Reader interface {
	Read() error
}

// An interface satisfied only through a pointer receiver.
type Adder interface {
	Add5() int
}

// Empty interface.
type EmptyInterface interface{}

type Base struct{}

// A declared type with the same name as a type in the second ground, so the same node name appears
// in two packages and in two folders.
type Thing struct {
	Value int
}

type Outer struct {
	Base
}

type Box[T any] struct {
	Value T
}

// Defined type with a map underlying.
type StringMap map[string]string

// Functional option type.
type Option func(*Struct1) error

const (
	Five = 5
	two  = 2
)

const (
	LayoutMono int = iota
	LayoutStereo
)

var zeroStruct *Struct1

var (
	first  = 1
	second = 2
)

// Unnamed function held by a package-level variable.
var namedFunc = func() int { return 0 }

func NewStruct1() *Struct1 {
	s1 := &Struct1{}
	s1.Field1 = 3
	s1.Field3 = func(v int) int {
		return s1.Field1 + v
	}
	return s1
}

func (s1 *Struct1) Add5() int {
	return s1.Field1 + 5
}

// Method with its own type parameters.
func (b Box[T]) PickVal[U any](x U) U { return x }

func (b Base) Read() error {
	return nil
}

func (b Base) Describe() string {
	return "base"
}

func (t Thing) Read() int { return t.Value }

// A package-level declaration sharing a name with a method in the same package, which is what keeps
// the receiver part of a name.
func Read() int { return 0 }

// Method written with an alias name as its receiver.
func (Mirror) DescribeAlias() string {
	return "alias"
}

func Pick[T any](b *Box[T]) T {
	return b.Value
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func describe(r Reader) error {
	return r.Read()
}

func init() {
	fmt.Println("init")
}

func main() {
	zeroStruct.Add5()
	s1 := NewStruct1()
	s1.Field3(4)
	fmt.Println(s1)

	local := 0
	_ = local
	anon := func() int { return min(first, second) }
	_ = anon()
	_ = str.ToLower("testground")
	_ = fmt.Sprint(Five, two, LayoutMono, LayoutStereo)

	outer := Outer{Base{}}
	_ = outer.Read()

	box := &Box[int]{Value: first}
	_ = Pick(box)

	_ = describe(s1.Reader)
	_ = Name("testground")
	_ = second
	_ = namedFunc()
	_ = Mirror{}
	_ = NameAlias("")
	_ = StringMap{}
	_ = Counter(3)
	_ = Option(nil)
	_ = EmptyInterface(nil)
}
