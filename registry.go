package marionette

import (
	"reflect"
	"strings"
	"sync"
)

// SuiteInstance is a pointer to a user-defined struct embedding Suite.
type SuiteInstance interface {
	inject(*T)
}

type registeredSuite struct {
	name     string
	instance SuiteInstance
	typ      reflect.Type
	methods  []reflect.Method
}

var (
	registryMu sync.RWMutex
	registry   []*registeredSuite
)

// Register adds a suite to the global test registry.
// Call from init() in each test file.
// suite must be a pointer to a struct that embeds marionette.Suite.
func Register(suite SuiteInstance) {
	rs := buildRegisteredSuite(suite)
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = append(registry, rs)
}

func buildRegisteredSuite(suite SuiteInstance) *registeredSuite {
	t := reflect.TypeOf(suite)
	name := t.Elem().Name()

	var methods []reflect.Method
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if !isTestMethod(m) {
			continue
		}
		methods = append(methods, m)
	}

	return &registeredSuite{
		name:     name,
		instance: suite,
		typ:      t,
		methods:  methods,
	}
}

// isTestMethod returns true for exported methods with no parameters (beyond receiver)
// and no return values, excluding Suite infrastructure methods.
func isTestMethod(m reflect.Method) bool {
	if !m.IsExported() {
		return false
	}

	// Exclude embedded Suite methods by name
	switch m.Name {
	case "AssertTrue", "AssertFalse", "AssertEqual", "AssertNotEqual",
		"AssertNear", "AssertSequenceEqual", "Fail", "Skip",
		"WriteArtifact", "ArtifactDirectory", "RunCases":
		return false
	}

	// Must have only receiver, no other params, no return values
	mt := m.Type
	if mt.NumIn() != 1 {
		return false
	}
	if mt.NumOut() != 0 {
		return false
	}

	return true
}

func allSuites() []*registeredSuite {
	registryMu.RLock()
	defer registryMu.RUnlock()
	result := make([]*registeredSuite, len(registry))
	copy(result, registry)
	return result
}

func matchesFilter(testName, filter string) bool {
	if filter == "" {
		return true
	}
	return strings.Contains(testName, filter)
}
