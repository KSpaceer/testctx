package unusedimport

import (
	"context"
	"testing"
)

func testFunc2(t *testing.T) {
	ctx := context.Background() // want "using context.Background/context.TODO when testing.TB is present"
	_ = ctx
	_ = t
}