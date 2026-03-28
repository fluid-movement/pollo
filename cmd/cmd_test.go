package cmd_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// polloBin is the path to the compiled test binary, set by TestMain.
var polloBin string

func TestMain(m *testing.M) {
	bin, err := os.CreateTemp("", "pollo-test-*")
	if err != nil {
		panic(err)
	}
	bin.Close()
	polloBin = bin.Name()

	cmd := exec.Command("go", "build", "-o", polloBin, "github.com/yourusername/pollo")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build pollo: " + err.Error())
	}

	code := m.Run()
	os.Remove(polloBin)
	os.Exit(code)
}

// run executes the pollo binary with the given arguments and returns stdout,
// the exit code, and a parsed JSON map (nil if output is not valid JSON).
func run(args ...string) (stdout string, exitCode int, parsed map[string]any) {
	cmd := exec.Command(polloBin, args...)
	out, err := cmd.Output()
	stdout = strings.TrimRight(string(out), "\n")
	exitCode = 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
	}
	_ = json.Unmarshal([]byte(stdout), &parsed)
	return
}

// fixtureB copies fixture_b.po to a temp file and returns its path.
// The caller is responsible for cleanup (via t.TempDir or explicit remove).
func fixtureB(t *testing.T) string {
	t.Helper()
	src, err := os.ReadFile("../po/testdata/fixture_b.po")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "fixture_b.po")
	if err := os.WriteFile(path, src, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// ---- stats ----

func TestStatsBasic(t *testing.T) {
	_, code, got := run("stats", "../po/testdata/fixture_b.po")
	if code != 0 {
		t.Fatalf("exit code %d", code)
	}
	assertInt(t, got, "total", 4)
	assertInt(t, got, "translated", 1)
	assertInt(t, got, "fuzzy", 1)
	assertInt(t, got, "untranslated", 2)
	assertInt(t, got, "remaining", 3)
}

// ---- get ----

func TestGetFuzzyFirst(t *testing.T) {
	_, code, got := run("get", "../po/testdata/fixture_b.po")
	if code != 0 {
		t.Fatalf("exit code %d", code)
	}
	if got["done"] != false {
		t.Errorf("done = %v, want false", got["done"])
	}
	if got["msgid"] != "Delete selected items" {
		t.Errorf("msgid = %v, want %q", got["msgid"], "Delete selected items")
	}
	if got["state"] != "fuzzy" {
		t.Errorf("state = %v, want fuzzy", got["state"])
	}
}

func TestGetUntranslatedFirst(t *testing.T) {
	_, code, got := run("get", "../po/testdata/fixture_b.po", "--order", "untranslated-first")
	if code != 0 {
		t.Fatalf("exit code %d", code)
	}
	if got["state"] != "untranslated" {
		t.Errorf("state = %v, want untranslated", got["state"])
	}
}

func TestGetByID(t *testing.T) {
	_, code, got := run("get", "../po/testdata/fixture_b.po", "--id", "Cancel")
	if code != 0 {
		t.Fatalf("exit code %d", code)
	}
	if got["msgid"] != "Cancel" {
		t.Errorf("msgid = %v, want Cancel", got["msgid"])
	}
}

func TestGetByIDNotFound(t *testing.T) {
	_, code, got := run("get", "../po/testdata/fixture_b.po", "--id", "nonexistent")
	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
	if got["ok"] != false {
		t.Errorf("ok = %v, want false", got["ok"])
	}
}

func TestGetIDMutualExclusionWithOrder(t *testing.T) {
	// Regression: --id + --order fuzzy-first (explicit default) must still error.
	_, code, got := run("get", "../po/testdata/fixture_b.po", "--id", "Cancel", "--order", "fuzzy-first")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if got["ok"] != false {
		t.Errorf("ok = %v, want false", got["ok"])
	}
}

func TestGetIDMutualExclusionWithUntranslatedFirst(t *testing.T) {
	_, code, got := run("get", "../po/testdata/fixture_b.po", "--id", "Cancel", "--order", "untranslated-first")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if got["ok"] != false {
		t.Errorf("ok = %v, want false", got["ok"])
	}
}

func TestGetDoneWhenAllTranslated(t *testing.T) {
	_, code, got := run("get", "../po/testdata/fixture_a.po")
	if code != 0 {
		t.Fatalf("exit code %d", code)
	}
	if got["done"] != true {
		t.Errorf("done = %v, want true", got["done"])
	}
}

// ---- set ----

func TestSetSingular(t *testing.T) {
	path := fixtureB(t)
	_, code, got := run("set", path, "--id", "Cancel", "--translation", "Abbrechen")
	if code != 0 {
		t.Fatalf("exit code %d, got: %v", code, got)
	}
	if got["ok"] != true {
		t.Errorf("ok = %v, want true", got["ok"])
	}
	// Re-parse and verify fuzzy flag gone and translation set
	_, _, reparsed := run("get", path, "--id", "Cancel")
	if reparsed["state"] != "translated" {
		t.Errorf("after set, state = %v, want translated", reparsed["state"])
	}
}

func TestSetFuzzyEntry(t *testing.T) {
	path := fixtureB(t)
	_, code, got := run("set", path, "--id", "Delete selected items", "--translation", "Ausgewählte Elemente löschen")
	if code != 0 {
		t.Fatalf("exit code %d, got: %v", code, got)
	}
	if got["ok"] != true {
		t.Errorf("ok = %v, want true", got["ok"])
	}
	// Fuzzy entry should now be translated
	_, _, reparsed := run("get", path, "--id", "Delete selected items")
	if reparsed["state"] != "translated" {
		t.Errorf("state = %v, want translated", reparsed["state"])
	}
}

func TestSetPluralEntry(t *testing.T) {
	path := fixtureB(t)
	_, code, got := run("set", path, "--id", "%d item", "--translations", `["%d Element", "%d Elemente"]`)
	if code != 0 {
		t.Fatalf("exit code %d, got: %v", code, got)
	}
	if got["ok"] != true {
		t.Errorf("ok = %v, want true", got["ok"])
	}
}

func TestSetPluralLengthMismatch(t *testing.T) {
	path := fixtureB(t)
	_, code, got := run("set", path, "--id", "%d item", "--translations", `["%d Element"]`)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if got["ok"] != false {
		t.Errorf("ok = %v, want false", got["ok"])
	}
}

func TestSetSingularFlagOnPluralEntry(t *testing.T) {
	path := fixtureB(t)
	_, code, got := run("set", path, "--id", "%d item", "--translation", "Ein Element")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if got["ok"] != false {
		t.Errorf("ok = %v, want false", got["ok"])
	}
}

func TestSetEntryNotFound(t *testing.T) {
	path := fixtureB(t)
	_, code, got := run("set", path, "--id", "nonexistent", "--translation", "x")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if got["ok"] != false {
		t.Errorf("ok = %v, want false", got["ok"])
	}
}

// ---- helpers ----

func assertInt(t *testing.T, m map[string]any, key string, want int) {
	t.Helper()
	v, ok := m[key]
	if !ok {
		t.Errorf("missing key %q", key)
		return
	}
	// JSON numbers unmarshal as float64
	got, ok := v.(float64)
	if !ok {
		t.Errorf("%q = %T(%v), want int %d", key, v, v, want)
		return
	}
	if int(got) != want {
		t.Errorf("%q = %d, want %d", key, int(got), want)
	}
}
