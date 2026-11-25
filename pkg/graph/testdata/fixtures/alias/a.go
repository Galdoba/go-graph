package alias

import (
	"fmt"
	fx "github.com/Galdoba/go-graph/pkg/graph/testdata/fixtures/foreign"
)

func Use() error {
	fx.Export()
	return fmt.Errorf("boom")
}
