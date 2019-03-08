package config

import "testing"

func TestIsSupportedAction(t *testing.T) {
	testcases := map[string]bool{
		"install":   true,
		"upgrade":   true,
		"uninstall": true,
		"status":    false,
		"INSTALL":   false,
	}

	for action, wantSupported := range testcases {
		t.Run(action, func(t *testing.T) {
			gotSupported := IsSupportedAction(action)
			if wantSupported != gotSupported {
				t.Fatalf("IsSupportedAction(%q) failed, want %t, got %t", action, wantSupported, gotSupported)
			}
		})
	}
}
