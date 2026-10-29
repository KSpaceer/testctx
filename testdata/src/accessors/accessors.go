package accessors

import (
	"containerfixture"
	"context"
	"testing"
)

func pointerReceiver(suite *containerfixture.Suite) context.Context {
	return context.Background() // want "using context.Background/context.TODO when testing.TB is present"
}

func addressableValue(suite containerfixture.Suite) context.Context {
	return context.TODO() // want "using context.Background/context.TODO when testing.TB is present"
}

func valueReceiver(suite containerfixture.ValueSuite) context.Context {
	return context.Background() // want "using context.Background/context.TODO when testing.TB is present"
}

func benchmarkResult(suite *containerfixture.BenchmarkSuite) context.Context {
	return context.TODO() // want "using context.Background/context.TODO when testing.TB is present"
}

func promoted(suite struct{ *containerfixture.Suite }) context.Context {
	return context.Background() // want "using context.Background/context.TODO when testing.TB is present"
}

func nested(suite struct{ Env containerfixture.Suite }) context.Context {
	return context.TODO() // want "using context.Background/context.TODO when testing.TB is present"
}

func interfaceAccessor(suite interface{ T() *testing.T }) context.Context {
	return context.Background() // want "using context.Background/context.TODO when testing.TB is present"
}

func preferDirect(suite *containerfixture.Suite, t *testing.T) context.Context {
	return context.TODO() // want "using context.Background/context.TODO when testing.TB is present"
}

func preferField(suite *containerfixture.Suite, env containerfixture.Exported) context.Context {
	return context.Background() // want "using context.Background/context.TODO when testing.TB is present"
}

type embeddedSuite struct{ *containerfixture.Suite }

func (suite *embeddedSuite) receiver() context.Context {
	return context.TODO() // want "using context.Background/context.TODO when testing.TB is present"
}

func cleanup(suite *containerfixture.Suite) {
	suite.T().Cleanup(func() {
		_ = context.Background()
	})
}

func requiresArgument(suite *containerfixture.ArgumentSuite) context.Context {
	return context.Background()
}

func variadic(suite *containerfixture.VariadicSuite) context.Context {
	return context.TODO()
}

func multipleResults(suite *containerfixture.MultipleResultsSuite) context.Context {
	return context.Background()
}

func noResults(suite *containerfixture.NoResultsSuite) context.Context {
	return context.TODO()
}

func wrongResult(suite *containerfixture.WrongResultSuite) context.Context {
	return context.Background()
}

func nonAddressableResult(suite *containerfixture.ValueResultSuite) context.Context {
	return context.TODO()
}

func wrongName(suite *containerfixture.WrongNameSuite) context.Context {
	return context.Background()
}

func functionField(suite struct{ T func() *testing.T }) context.Context {
	return context.TODO()
}

func shadowed(suite *containerfixture.Suite) context.Context {
	{
		suite := 1
		_ = suite
		_ = context.Background()
	}
	return context.TODO() // want "using context.Background/context.TODO when testing.TB is present"
}

func declaredLater() context.Context {
	ctx := context.Background()
	var suite containerfixture.Suite
	_ = suite
	return ctx
}
