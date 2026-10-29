package indirectcontainer

import (
	"containerfixture"
	"context"
)

// This package only imports testing transitively through containerfixture.
func indirect(suite containerfixture.Exported) context.Context {
	return context.Background() // want "using context.Background/context.TODO when testing.TB is present"
}
