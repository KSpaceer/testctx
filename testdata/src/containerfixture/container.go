package containerfixture

import "testing"

type Exported struct {
	TB testing.TB
}

type Private struct {
	tb testing.TB
}
