package i18n

import "testing"

func TestI18nFallback(t *testing.T) {
	val := T("tray_status")
	if val == "" {
		t.Error("expected non-empty translation for tray_status")
	}
}

func TestI18nMultipleLangs(t *testing.T) {
	// Test Portuguese
	SetLanguage("pt")
	val := T("tray_status")
	if val == "" || val == "tray_status" {
		t.Errorf("expected Portuguese translation, got '%s'", val)
	}

	// Test English fallback
	SetLanguage("en")
	valEn := T("tray_status")
	if valEn != "Status: Running" {
		t.Errorf("expected 'Status: Running', got '%s'", valEn)
	}

	// Test Spanish
	SetLanguage("es")
	valEs := T("tray_status")
	if valEs != "Estado: En ejecución" {
		t.Errorf("expected 'Estado: En ejecución', got '%s'", valEs)
	}

	// Test French
	SetLanguage("fr")
	valFr := T("tray_status")
	if valFr != "Statut : En cours" {
		t.Errorf("expected 'Statut : En cours', got '%s'", valFr)
	}

	// Reset to default
	SetLanguage("pt")
}

func TestI18nUnknownKey(t *testing.T) {
	SetLanguage("en")
	val := T("unknown_key_xyz")
	if val != "unknown_key_xyz" {
		t.Errorf("expected key returned for unknown key, got '%s'", val)
	}
}
