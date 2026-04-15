package marionette

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
)

// Suite is embedded in user test structs to provide assertion methods.
// Each exported method on the embedding struct is discovered as a test.
type Suite struct {
	t *T
}

// inject is called by the runner to wire the active T into the suite before each test.
func (s *Suite) inject(t *T) {
	s.t = t
}

// T carries per-test state: failures, skips, and artifact paths.
// It is also exposed directly in RunCases callbacks.
type T struct {
	testName      string
	theoryCaseName string
	failures      []failure
	artifactPaths []string
	skipped       bool
	skipReason    string
	skipFile      string
	skipLine      int
}

type failure struct {
	file      string
	line      int
	assertion string
	message   string
	expected  string
	actual    string
}

func newT(testName string) *T {
	return &T{testName: testName}
}

func (t *T) displayName() string {
	if t.theoryCaseName != "" {
		return t.testName + "[" + t.theoryCaseName + "]"
	}
	return t.testName
}

func (t *T) record(assertion, message, expected, actual string) {
	_, file, line, _ := runtime.Caller(2)
	t.failures = append(t.failures, failure{
		file:      file,
		line:      line,
		assertion: assertion,
		message:   message,
		expected:  expected,
		actual:    actual,
	})
}

// AssertTrue fails if condition is false.
func (s *Suite) AssertTrue(condition bool, message string) {
	if !condition {
		s.t.record("AssertTrue", message, "true", "false")
	}
}

// AssertFalse fails if condition is true.
func (s *Suite) AssertFalse(condition bool, message string) {
	if condition {
		s.t.record("AssertFalse", message, "false", "true")
	}
}

// AssertEqual fails if expected != actual.
func (s *Suite) AssertEqual(expected, actual any, message string) {
	if !reflect.DeepEqual(expected, actual) {
		s.t.record("AssertEqual", message, fmt.Sprintf("%v", expected), fmt.Sprintf("%v", actual))
	}
}

// AssertNotEqual fails if expected == actual.
func (s *Suite) AssertNotEqual(expected, actual any, message string) {
	if reflect.DeepEqual(expected, actual) {
		s.t.record("AssertNotEqual", message, fmt.Sprintf("not %v", expected), fmt.Sprintf("%v", actual))
	}
}

// AssertNear fails if |expected - actual| > tolerance.
// All three arguments must be the same float type.
func (s *Suite) AssertNear(expected, actual, tolerance float64, message string) {
	if math.Abs(expected-actual) > tolerance {
		s.t.record(
			"AssertNear",
			message,
			fmt.Sprintf("%v ± %v", expected, tolerance),
			fmt.Sprintf("%v (diff %v)", actual, math.Abs(expected-actual)),
		)
	}
}

// AssertSequenceEqual fails if the two slices differ in length or element order.
func (s *Suite) AssertSequenceEqual(expected, actual any, message string) {
	ev := reflect.ValueOf(expected)
	av := reflect.ValueOf(actual)

	if ev.Kind() != reflect.Slice && ev.Kind() != reflect.Array {
		s.t.record("AssertSequenceEqual", message, "slice or array", fmt.Sprintf("%T", expected))
		return
	}
	if av.Kind() != reflect.Slice && av.Kind() != reflect.Array {
		s.t.record("AssertSequenceEqual", message, "slice or array", fmt.Sprintf("%T", actual))
		return
	}

	if ev.Len() != av.Len() {
		s.t.record(
			"AssertSequenceEqual",
			message,
			fmt.Sprintf("len=%d", ev.Len()),
			fmt.Sprintf("len=%d", av.Len()),
		)
		return
	}

	for i := 0; i < ev.Len(); i++ {
		if !reflect.DeepEqual(ev.Index(i).Interface(), av.Index(i).Interface()) {
			s.t.record(
				"AssertSequenceEqual",
				message,
				fmt.Sprintf("[%d]=%v", i, ev.Index(i).Interface()),
				fmt.Sprintf("[%d]=%v", i, av.Index(i).Interface()),
			)
			return
		}
	}
}

// Fail records an unconditional failure.
func (s *Suite) Fail(message string) {
	s.t.record("Fail", message, "", "")
}

// Skip marks the test as skipped with a reason. Remaining test body is still executed.
func (s *Suite) Skip(reason string) {
	_, file, line, _ := runtime.Caller(1)
	s.t.skipped = true
	s.t.skipReason = reason
	s.t.skipFile = file
	s.t.skipLine = line
}

// WriteArtifact writes content to <repoRoot>/out/test-artifacts/<TestName>/<artifactName>.txt.
// Returns the path written, or empty string on failure (also records a failure).
func (s *Suite) WriteArtifact(artifactName, content string) string {
	path, err := writeTextArtifact(s.t.testName, artifactName, content)
	if err != nil {
		s.t.record("WriteArtifact", fmt.Sprintf("failed to write artifact %q: %v", artifactName, err), "", "")
		return ""
	}
	s.t.artifactPaths = append(s.t.artifactPaths, path)
	return path
}

// ArtifactDirectory returns the directory where artifacts for this test are written.
func (s *Suite) ArtifactDirectory() string {
	return artifactDirectory(s.t.testName)
}

// RunCases iterates over cases and runs fn for each, reporting failures per case.
// cases must be a slice or array whose elements have a Name string field.
func (s *Suite) RunCases(cases any, fn func(t *T, caseValue any)) {
	cv := reflect.ValueOf(cases)
	if cv.Kind() != reflect.Slice && cv.Kind() != reflect.Array {
		s.Fail("RunCases: cases must be a slice or array")
		return
	}

	for i := 0; i < cv.Len(); i++ {
		caseVal := cv.Index(i).Interface()

		nameField := cv.Index(i).FieldByName("Name")
		caseName := fmt.Sprintf("case%d", i)
		if nameField.IsValid() && nameField.Kind() == reflect.String {
			caseName = nameField.String()
		}

		caseT := &T{
			testName:       s.t.testName,
			theoryCaseName: caseName,
		}

		fn(caseT, caseVal)

		s.t.failures = append(s.t.failures, caseT.failures...)
		s.t.artifactPaths = append(s.t.artifactPaths, caseT.artifactPaths...)

		if caseT.skipped {
			s.t.skipped = true
			s.t.skipReason = caseT.skipReason
			s.t.skipFile = caseT.skipFile
			s.t.skipLine = caseT.skipLine
			return
		}
	}
}
