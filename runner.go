package marionette

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Run executes all registered suites using os.Args for filter and exits with
// the appropriate code. Call this from main().
func Run() {
	os.Exit(RunWithArgs(os.Args[1:]))
}

// RunWithArgs executes all registered suites and returns an exit code.
// args may contain an optional filter substring as the first element.
func RunWithArgs(args []string) int {
	filter := ""
	if len(args) >= 1 {
		filter = args[0]
	}

	suites := allSuites()

	total := 0
	passed := 0
	skipped := 0
	failed := 0

	for _, rs := range suites {
		for _, method := range rs.methods {
			testName := rs.name + "." + method.Name
			if !matchesFilter(testName, filter) {
				continue
			}

			result := runOne(rs, method, testName)
			total++

			switch {
			case result.skipped:
				skipped++
				printSkip(result)
			case len(result.failures) > 0:
				failed++
				printFailures(result)
			default:
				passed++
				printPass(result)
			}
		}
	}

	fmt.Printf("\nSummary: %d test(s), %d passed, %d skipped, %d failed\n",
		total, passed, skipped, failed)

	if failed > 0 {
		return 1
	}
	return 0
}

type testResult struct {
	testName      string
	failures      []failure
	artifactPaths []string
	skipped       bool
	skipReason    string
	skipFile      string
	skipLine      int
}

func runOne(rs *registeredSuite, method reflect.Method, testName string) testResult {
	t := newT(testName)

	// Create a fresh instance of the suite type for each test so tests are isolated
	freshVal := reflect.New(rs.typ.Elem())
	fresh := freshVal.Interface().(SuiteInstance)
	fresh.inject(t)

	method.Func.Call([]reflect.Value{freshVal})

	return testResult{
		testName:      testName,
		failures:      t.failures,
		artifactPaths: t.artifactPaths,
		skipped:       t.skipped,
		skipReason:    t.skipReason,
		skipFile:      t.skipFile,
		skipLine:      t.skipLine,
	}
}

func printPass(r testResult) {
	fmt.Printf("[PASS] %s\n", r.testName)
	for _, path := range r.artifactPaths {
		fmt.Printf("    artifact: %s\n", path)
	}
}

func printSkip(r testResult) {
	fmt.Printf("[SKIP] %s\n", r.testName)
	fmt.Printf("  SKIP %s at %s:%d\n", r.testName, shortPath(r.skipFile), r.skipLine)
	fmt.Printf("    reason: %s\n", r.skipReason)
}

func printFailures(r testResult) {
	fmt.Printf("[FAIL] %s\n", r.testName)
	for _, f := range r.failures {
		fmt.Printf("  FAIL %s at %s:%d\n", f.assertion, shortPath(f.file), f.line)
		fmt.Printf("    message: %s\n", f.message)
		if f.expected != "" {
			fmt.Printf("    expected: %s\n", f.expected)
		}
		if f.actual != "" {
			fmt.Printf("    actual:   %s\n", f.actual)
		}
	}
	for _, path := range r.artifactPaths {
		fmt.Printf("    artifact: %s\n", path)
	}
}

func shortPath(path string) string {
	// Trim everything up to and including the module root for readable output
	for _, sep := range []string{"/marionette-go/", "/smoke_tests/"} {
		if idx := strings.LastIndex(path, sep); idx >= 0 {
			return path[idx+1:]
		}
	}
	return path
}
