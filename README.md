# marionette-go

A lightweight test harness for Go. No `go test`, no `t.Errorf` format strings, no external dependencies. Designed for LLM-first test authoring.

## Why

Go's built-in test story requires you to manually construct diagnostic messages, has no first-class theory parameterization, and no artifact concept. Marionette-go replaces it entirely with an xUnit-style suite model: named test methods, structured assertions with automatic expected/actual formatting, named theory cases, and file-based diagnostic artifacts.

## Getting started

```sh
go get github.com/yuechen-li-dev/marionette-go
```

Your test binary needs a `main.go` that calls `marionette.Run()`:

```go
package main

import (
    marionette "github.com/yuechen-li-dev/marionette-go"
    _ "yourmodule/mytests"
)

func main() {
    marionette.Run()
}
```

The blank import pulls in your test packages via their `init()` functions. Add one blank import per test package.

## Writing tests

Create a struct that embeds `marionette.Suite`. Every exported method with no parameters and no return value is discovered as a test. Register the suite in `init()`.

```go
package mytests

import marionette "github.com/yuechen-li-dev/marionette-go"

func init() {
    marionette.Register(&MathTests{})
}

type MathTests struct{ marionette.Suite }

func (s *MathTests) AdditionWorks() {
    s.AssertEqual(5, 2+3, "basic addition should hold")
}

func (s *MathTests) SubtractionWorks() {
    s.AssertEqual(3, 5-2, "basic subtraction should hold")
}
```

Output:

```
[PASS] MathTests.AdditionWorks
[PASS] MathTests.SubtractionWorks

Summary: 2 test(s), 2 passed, 0 skipped, 0 failed
```

## Assertions

### `AssertTrue` / `AssertFalse`

```go
s.AssertTrue(condition, "must be true")
s.AssertFalse(condition, "must be false")
```

### `AssertEqual` / `AssertNotEqual`

```go
s.AssertEqual(expected, actual, "values should match")
s.AssertNotEqual(expected, actual, "values should differ")
```

Uses `reflect.DeepEqual` — works for structs, slices, maps, and primitives.

### `AssertNear`

```go
s.AssertNear(10.0, measured, 0.1, "difference should stay within tolerance")
```

Fails if `|expected - actual| > tolerance`. All three arguments are `float64`.

### `AssertSequenceEqual`

```go
s.AssertSequenceEqual(expectedSlice, actualSlice, "ordered sequence should match")
```

Compares length and element order. Elements are compared with `reflect.DeepEqual`.

### `Fail`

```go
if fatalCondition {
    s.Fail("fatal condition encountered")
}
```

Records an unconditional failure. The test method continues executing.

### `Skip`

```go
if !preconditionAvailable {
    s.Skip("precondition unavailable in this environment")
}
```

Marks the test as skipped. Appears in output as `[SKIP]` with the reason. Does not count as a failure.

## Theory: parameterized tests

`RunCases` runs the same assertion logic across a collection of named cases. The case struct must have a `Name string` field — this becomes the case label in failure output.

```go
func (s *MathTests) AdditionTheory() {
    type Case struct {
        Name     string
        Left     int
        Right    int
        Expected int
    }

    s.RunCases([]Case{
        {"zeros", 0, 0, 0},
        {"small-positive", 2, 3, 5},
        {"mixed-sign", 5, -2, 3},
    }, func(t *marionette.T, caseValue any) {
        c := caseValue.(Case)
        s.AssertEqual(c.Expected, c.Left+c.Right, "addition should match")
    })
}
```

If a case fails, output identifies it by name: `[FAIL] MathTests.AdditionTheory[mixed-sign]`.

## Artifacts

Write diagnostic files that persist after the test run. Useful for capturing structured output — traces, diffs, serialized state — that would be too large or too structured for an assertion message.

```go
func (s *MathTests) WritesTrace() {
    trace := buildTrace()
    s.WriteArtifact("trace", trace)
}
```

Artifacts are written to `<repoRoot>/out/test-artifacts/<SuiteName>_<TestName>/<artifactName>.txt`. On pass, the artifact path is printed below the `[PASS]` line. On failure, it is printed alongside the failure detail.

To control the repo root, set the `MARIONETTE_REPO_ROOT` environment variable, or call `marionette.SetRepoRoot(path)` before `Run()`. If neither is set, the current working directory is used.

```go
func (s *MathTests) InspectsArtifactPath() {
    dir := s.ArtifactDirectory()  // full path to this test's artifact directory
    path := s.WriteArtifact("summary", `{"status": "ok"}`)
    s.AssertTrue(path != "", "artifact write should succeed")
}
```

## Running tests

```sh
# Run all tests
MARIONETTE_REPO_ROOT=$(pwd) go run ./

# Run tests whose name contains a filter substring
MARIONETTE_REPO_ROOT=$(pwd) go run ./ Math

# Non-zero exit code on any failure
```

Filtering matches against the full test name including suite: `MathTests.AdditionWorks`.

## Anti-patterns

- Do not use `AssertTrue` as a replacement for `AssertEqual` — the latter shows expected and actual values in failure output, which makes failures readable without re-running under a debugger.
- Do not skip writing artifacts when debugging replay or comparison mismatches — capture bounded evidence so the failure is self-contained.
- Do not use `Fail` as a substitute for assertions — prefer the specific assertion so failure output is structured.
- Do not name theory case structs without a `Name` field — unnamed cases fall back to `case0`, `case1`, which makes failure output unreadable.
