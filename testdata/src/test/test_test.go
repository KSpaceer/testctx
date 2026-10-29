package test_test

import (
	"context"
	"fmt"
	"testing"
)

func TestOK(t *testing.T) { // OK
	ctx := t.Context()
	call(ctx, "ok")
}

func TestBackground(t *testing.T) {
	ctx := context.Background() // want "using context.Background/context.TODO when testing.TB is present"
	call(ctx, "background")
}

func TestTODO(t *testing.T) {
	ctx := context.TODO() // want "using context.Background/context.TODO when testing.TB is present"
	call(ctx, "todo")
}

func OK_NoTB() { // OK
	ctx := context.Background()
	call(ctx, "background_notb")
	ctx2 := context.TODO()
	call(ctx2, "todo_notb")
}

func TestOK_AnonFuncWithinCleanup(t *testing.T) { // OK
	t.Cleanup(func() {
		ctx := context.Background()
		call(ctx, "ok_anonfunc_within_cleanup")
	})
}

func TestOK_FuncPassedToCleanup(t *testing.T) { // OK
	t.Cleanup(cleanupWrapper)
}

func TestOK_FunctionCalledInAnonFunc(t *testing.T) { // OK
	t.Cleanup(func() {
		cleanupHelper(t)
	})
}

func BenchmarkTODO(b *testing.B) {
	ctx := context.TODO() // want "using context.Background/context.TODO when testing.TB is present"
	call(ctx, "benchmark_todo")
}

func cleanupWrapper() {
	ctx := context.Background()
	call(ctx, "wrapped")
}

func call(_ context.Context, name string) {
	fmt.Println(name)
}

func cleanupHelper(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Log(ctx)
}
