package web

import (
	"os/exec"
	"testing"
)

// TestEnglishTranslationCoverage fails when a Chinese UI string on the main
// pages has no English entry in i18n-en.js.
func TestEnglishTranslationCoverage(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}
	if out, err := exec.Command(node, "i18n_coverage_test.js").CombinedOutput(); err != nil {
		t.Fatalf("english translation coverage: %v\n%s", err, out)
	}
}
