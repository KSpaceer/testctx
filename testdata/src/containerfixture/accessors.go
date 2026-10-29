package containerfixture

import "testing"

// Keep testing fields private so consumers must use the accessors.
type Suite struct{ t *testing.T }

func (suite *Suite) T() *testing.T { return suite.t }

type ValueSuite struct{ tb testing.TB }

func (suite ValueSuite) T() testing.TB { return suite.tb }

type BenchmarkSuite struct{ b *testing.B }

func (suite *BenchmarkSuite) T() *testing.B { return suite.b }

type ArgumentSuite struct{}

func (*ArgumentSuite) T(int) *testing.T { return nil }

type VariadicSuite struct{}

func (*VariadicSuite) T(...int) *testing.T { return nil }

type MultipleResultsSuite struct{}

func (*MultipleResultsSuite) T() (*testing.T, error) { return nil, nil }

type NoResultsSuite struct{}

func (*NoResultsSuite) T() {}

type WrongResultSuite struct{}

func (*WrongResultSuite) T() string { return "" }

type ValueResultSuite struct{}

func (*ValueResultSuite) T() testing.T { return testing.T{} }

type WrongNameSuite struct{}

func (*WrongNameSuite) Testing() *testing.T { return nil }
