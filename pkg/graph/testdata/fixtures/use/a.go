package use

import (
	"strings"

	shp "github.com/Galdoba/go-graph/pkg/graph/testdata/fixtures/shape"
)

func Use() error {
	var s shp.Shape
	if err := s.Read(); err != nil {
		return err
	}
	var shown shp.Shown
	_ = shown
	var box shp.Box[int]
	_ = box
	var w shp.Wrapper
	_ = w.Count()
	_ = strings.ToLower("x")
	return nil
}
