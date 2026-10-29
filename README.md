# testctx
testctx is a linter that detects context.Background/context.TODO usages in tests when testing.T.Context is available

## Install

```bash
go install github.com/KSpaceer/testctx/cmd/testctx@latest
```

## How it works

testctx analyzer searches for context.Background or context.TODO calls, then it checks
if testing.TB is available here.

testing.TB is available if:

- any variable in the scope implements it
- any field (including embedded or nested) of struct implements it
- any variable/field has T() <testing.TB implementation> method (e.g. testify Suite)

If the call may be directly or indirectly done within testing.TB.Cleanup, no issue is raised as testing.TB.Context is already cancelled and context.Background/context.TODO usage is justified
