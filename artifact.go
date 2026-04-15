package marionette

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	repoRootMu sync.RWMutex
	repoRoot   string
)

// SetRepoRoot sets the repository root for artifact output.
// Artifacts are written to <repoRoot>/out/test-artifacts/<TestName>/.
// If not set, the MARIONETTE_REPO_ROOT environment variable is used.
// If neither is set, the current working directory is used.
func SetRepoRoot(path string) {
	repoRootMu.Lock()
	defer repoRootMu.Unlock()
	repoRoot = path
}

func resolveRepoRoot() string {
	repoRootMu.RLock()
	r := repoRoot
	repoRootMu.RUnlock()

	if r != "" {
		return r
	}

	if env := os.Getenv("MARIONETTE_REPO_ROOT"); env != "" {
		return env
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func artifactDirectory(testName string) string {
	return filepath.Join(resolveRepoRoot(), "out", "test-artifacts", sanitizePathComponent(testName))
}

func writeTextArtifact(testName, artifactName, content string) (string, error) {
	dir := artifactDirectory(testName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	filename := sanitizePathComponent(artifactName) + ".txt"
	path := filepath.Join(dir, filename)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}

	return path, nil
}

func sanitizePathComponent(value string) string {
	var b strings.Builder
	b.Grow(len(value))

	for _, c := range value {
		isLetter := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
		isDigit := c >= '0' && c <= '9'
		isAllowed := c == '-' || c == '_'

		if isLetter || isDigit || isAllowed {
			b.WriteRune(c)
		} else {
			b.WriteRune('_')
		}
	}

	if b.Len() == 0 {
		return "unnamed"
	}
	return b.String()
}
