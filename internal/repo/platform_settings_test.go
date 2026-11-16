package repo

import "testing"

func TestPlatformInterfaceLanguage(t *testing.T) {
	s, err := OpenTest()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if got := s.PlatformInterfaceLanguage(); got != InterfaceLanguageUserChoice {
		t.Fatalf("default = %q", got)
	}
	if err := s.SetPlatformInterfaceLanguage("de", 1); err != ErrInvalidInterfaceLanguage {
		t.Fatalf("invalid language error = %v", err)
	}
	if err := s.SetPlatformInterfaceLanguage(" ZH ", 1); err != nil {
		t.Fatal(err)
	}
	if got := s.PlatformInterfaceLanguage(); got != InterfaceLanguageChinese {
		t.Fatalf("saved = %q", got)
	}
	if err := s.SetPlatformInterfaceLanguage("en", 1); err != nil {
		t.Fatal(err)
	}
	if got := s.PlatformInterfaceLanguage(); got != InterfaceLanguageEnglish {
		t.Fatalf("updated = %q", got)
	}

	// A value written outside the API (e.g. by hand in SQL) never locks the
	// interface to an unsupported language.
	if err := s.UsersRepo.DB.Model(&platformSettingsRow{}).Where("id = ?", platformSettingsID).Update("interface_language", "fr").Error; err != nil {
		t.Fatal(err)
	}
	if got := s.PlatformInterfaceLanguage(); got != InterfaceLanguageUserChoice {
		t.Fatalf("unknown stored value = %q", got)
	}
}
