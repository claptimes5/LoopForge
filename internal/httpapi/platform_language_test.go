package httpapi

import (
	"encoding/json"
	"strings"
	"testing"

	"agentflow/internal/repo"
)

// Rule: only a platform administrator (same-origin browser session or bearer
// token) can change the interface language; anyone, signed in or not, can
// read the language value, and nothing else, from /platform-language.js.
func TestPlatformAdminsLockInterfaceLanguage(t *testing.T) {
	s, client, owner, member := platformSettingsFixture(t)
	const origin = "http://localhost"

	script := func() string {
		t.Helper()
		w := client.call(nil, "GET", "/platform-language.js", "", "")
		if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/javascript") || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("policy script: %d %q %q", w.Code, w.Header().Get("Content-Type"), w.Header().Get("Cache-Control"))
		}
		return w.Body.String()
	}
	if got := script(); got != "window.LF_LANG_POLICY=\"\";\n" {
		t.Fatalf("default policy script = %q", got)
	}

	if w := client.call(nil, "GET", "/api/platform/language", "", ""); w.Code != 401 && w.Code != 403 {
		t.Fatalf("anonymous read: %d", w.Code)
	}
	if w := client.call(nil, "PUT", "/api/platform/language", origin, `{"language":"en"}`); w.Code != 401 && w.Code != 403 {
		t.Fatalf("anonymous save: %d", w.Code)
	}
	if w := client.call(member, "GET", "/api/platform/language", "", ""); w.Code != 403 {
		t.Fatalf("member read: %d", w.Code)
	}
	if w := client.call(member, "PUT", "/api/platform/language", origin, `{"language":"en"}`); w.Code != 403 {
		t.Fatalf("member save: %d", w.Code)
	}
	if w := client.call(owner, "PUT", "/api/platform/language", "https://attacker.invalid", `{"language":"en"}`); w.Code != 403 {
		t.Fatalf("cross-origin save: %d", w.Code)
	}
	if w := client.call(owner, "PUT", "/api/platform/language", "", `{"language":"en"}`); w.Code != 403 {
		t.Fatalf("save without origin: %d", w.Code)
	}
	for _, body := range []string{`{}`, `{"language":"fr"}`, `{"language":null}`, `not json`} {
		if w := client.call(owner, "PUT", "/api/platform/language", origin, body); w.Code != 400 {
			t.Fatalf("invalid %s: %d %s", body, w.Code, w.Body.String())
		}
	}
	if got := s.PlatformInterfaceLanguage(); got != "" {
		t.Fatalf("rejected saves changed the language to %q", got)
	}

	w := client.call(owner, "PUT", "/api/platform/language", origin, `{"language":" EN "}`)
	var status repo.PlatformInterfaceLanguageStatus
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &status) != nil || status.Language != "en" || status.UpdatedBy != "owner" || status.UpdatedAt == nil {
		t.Fatalf("save: %d %s", w.Code, w.Body.String())
	}
	if got := script(); got != "window.LF_LANG_POLICY=\"en\";\n" {
		t.Fatalf("locked policy script = %q", got)
	}
	w = client.call(owner, "GET", "/api/platform/language", "", "")
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &status) != nil || status.Language != "en" || status.UpdatedBy != "owner" {
		t.Fatalf("status: %d %s", w.Code, w.Body.String())
	}

	if w := client.call(owner, "PUT", "/api/platform/language", origin, `{"language":""}`); w.Code != 200 {
		t.Fatalf("unlock: %d %s", w.Code, w.Body.String())
	}
	if got := script(); got != "window.LF_LANG_POLICY=\"\";\n" {
		t.Fatalf("unlocked policy script = %q", got)
	}
}
