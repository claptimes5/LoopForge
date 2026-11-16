package web

import (
	"os/exec"
	"testing"
)

// TestInterfaceLanguageLock covers the platform-wide language lock in the
// browser: boot applies it before paint and the language switches hide.
func TestInterfaceLanguageLock(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}
	if out, err := exec.Command(node, "lang_lock_test.js").CombinedOutput(); err != nil {
		t.Fatalf("interface language lock: %v\n%s", err, out)
	}
}
