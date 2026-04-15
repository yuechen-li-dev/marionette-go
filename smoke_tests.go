package smoketests

import (
	"os"
	"path/filepath"

	marionette "github.com/yuechen-li-dev/marionette-go"
)

func init() {
	marionette.Register(&SmokeTests{})
}

type SmokeTests struct{ marionette.Suite }

func (s *SmokeTests) FactPasses() {
	s.AssertTrue(true, "basic true assertion should pass")
}

func (s *SmokeTests) SupportsRichAssertions() {
	isStable := true
	expectedCount := 3
	actualCount := 3
	left := "host"
	right := "candidate"

	s.AssertTrue(isStable, "true assertions can carry an explicit message")
	s.AssertFalse(false, "false assertions can carry an explicit message")
	s.AssertEqual(expectedCount, actualCount, "equal assertions show expected and actual when they diverge")
	s.AssertNotEqual(left, right, "not-equal assertions confirm distinct values")
}

func (s *SmokeTests) CanBeFailedDeliberately() {
	enableIntentionalFailure := false

	if enableIntentionalFailure {
		s.Fail("flip enableIntentionalFailure to true when you want to inspect failure output manually")
	}

	s.AssertTrue(true, "default smoke run stays green")
}

func (s *SmokeTests) TheorySupportsNamedCases() {
	type AdditionCase struct {
		Name     string
		Left     int
		Right    int
		Expected int
	}

	cases := []AdditionCase{
		{"zeros", 0, 0, 0},
		{"small-positive", 2, 3, 5},
		{"mixed-sign", 5, -2, 3},
	}

	s.RunCases(cases, func(t *marionette.T, caseValue any) {
		c := caseValue.(AdditionCase)
		s.AssertEqual(c.Expected, c.Left+c.Right,
			"theory cases should reuse the same assertion logic across multiple named rows")
	})
}

func (s *SmokeTests) CanBeSkipped() {
	s.Skip("example skipped tests stay visible without failing the default run")
}

func (s *SmokeTests) SupportsSequenceAssertions() {
	expected := []int{1, 2, 3, 5, 8}
	actual := []int{1, 2, 3, 5, 8}

	s.AssertSequenceEqual(expected, actual, "sequence equality should compare size and element order")
}

func (s *SmokeTests) WritesDeterministicArtifacts() {
	contents := "{\n  \"status\": \"ok\",\n  \"test\": \"WritesDeterministicArtifacts\"\n}\n"

	path := s.WriteArtifact("summary", contents)
	s.AssertTrue(path != "", "artifact helper should return a non-empty path")

	expectedPath := filepath.Join(s.ArtifactDirectory(), "summary.txt")
	s.AssertEqual(expectedPath, path, "artifact path should be deterministic")

	_, err := os.Stat(path)
	s.AssertTrue(err == nil, "artifact file should exist at the returned path")

	actualBytes, err := os.ReadFile(path)
	s.AssertTrue(err == nil, "artifact file should be readable after it is written")
	s.AssertEqual(contents, string(actualBytes), "artifact contents should be deterministic and overwritten in place")
}

func (s *SmokeTests) NearAssertionPassesWithinTolerance() {
	s.AssertNear(10.0, 10.05, 0.1, "value within tolerance should pass")
}

func (s *SmokeTests) NearAssertionPassesOnBoundary() {
	s.AssertNear(10.0, 10.1, 0.1, "value exactly on tolerance boundary should pass")
}
