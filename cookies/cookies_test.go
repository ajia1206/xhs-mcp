package cookies

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveCookiesCreatesParentDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "cookies.json")
	cookieStore := NewLoadCookie(path)

	if err := cookieStore.SaveCookies([]byte("{}")); err != nil {
		t.Fatalf("save cookies: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved cookies: %v", err)
	}

	if string(data) != "{}" {
		t.Fatalf("expected saved cookies to match input, got %q", string(data))
	}
}

func TestGetCookiesFilePathPrefersEnvVar(t *testing.T) {
	t.Setenv("COOKIES_PATH", "/custom/cookies.json")

	tmpRoot := t.TempDir()
	t.Setenv("TMPDIR", tmpRoot)

	legacyPath := filepath.Join(tmpRoot, "cookies.json")
	if err := os.WriteFile(legacyPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("write legacy cookies: %v", err)
	}

	if got := GetCookiesFilePath(); got != "/custom/cookies.json" {
		t.Fatalf("expected env path, got %q", got)
	}
}

func TestGetCookiesFilePathFallsBackToLegacyPath(t *testing.T) {
	t.Setenv("COOKIES_PATH", "")

	tmpRoot := t.TempDir()
	t.Setenv("TMPDIR", tmpRoot)

	legacyPath := filepath.Join(tmpRoot, "cookies.json")
	if err := os.WriteFile(legacyPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("write legacy cookies: %v", err)
	}

	if got := GetCookiesFilePath(); got != legacyPath {
		t.Fatalf("expected legacy path %q, got %q", legacyPath, got)
	}
}

func TestGetCookiesFilePathDefaultsToRepoRootFile(t *testing.T) {
	t.Setenv("COOKIES_PATH", "")

	tmpRoot := t.TempDir()
	t.Setenv("TMPDIR", tmpRoot)

	if got := GetCookiesFilePath(); got != "cookies.json" {
		t.Fatalf("expected default cookies.json, got %q", got)
	}
}
