package test_test

import (
	"context"
	"testing"
)

type pointerCleanup struct {
	t *testing.T
}

func TestOK_BoundPointerMethodCleanup(t *testing.T) {
	s := &pointerCleanup{t: t}
	t.Cleanup(s.cleanup)
}

func (s *pointerCleanup) cleanup() {
	s.t.Log(context.Background())
	s.helper()
}

func (s *pointerCleanup) helper() {
	s.t.Log(context.TODO())
}

type valueCleanup struct {
	t *testing.T
}

func TestOK_BoundValueMethodCleanup(t *testing.T) {
	s := valueCleanup{t: t}
	callback := s.cleanup
	t.Cleanup(callback)
}

func (s valueCleanup) cleanup() {
	s.t.Log(context.TODO())
}

func TestOrdinaryBoundMethod(t *testing.T) {
	s := &pointerCleanup{t: t}
	callback := s.ordinary
	callback()
}

func (s *pointerCleanup) ordinary() {
	s.t.Log(context.Background()) // want "using context.Background/context.TODO when testing.TB is present"
}
