package intent

import (
	"testing"
)

func TestTryLocalMatch(t *testing.T) {
	tests := []struct {
		input       string
		wantAction  string
		wantTarget  string
		wantCategory string
	}{
		// English
		{"install nginx", "install", "nginx", "package_management"},
		{"remove docker", "remove", "docker", "package_management"},
		{"uninstall apache2", "remove", "apache2", "package_management"},
		{"update", "update", "all", "package_management"},
		{"upgrade", "update", "all", "package_management"},
		{"restart nginx", "restart", "nginx", "service_management"},
		{"start docker", "start", "docker", "service_management"},
		{"stop redis", "stop", "redis", "service_management"},
		{"ports", "ports", "network", "system_info"},
		{"disk", "disk", "storage", "system_info"},
		{"ram", "memory", "ram", "system_info"},

		// Turkish
		{"nginx kur", "install", "nginx", "package_management"},
		{"docker kaldır", "remove", "docker", "package_management"},
		{"apache2 sil", "remove", "apache2", "package_management"},
		{"güncelle", "update", "all", "package_management"},
		{"sistemi güncelle", "update", "all", "package_management"},
		{"paketleri güncelle", "update", "all", "package_management"},
		{"nginx yeniden başlat", "restart", "nginx", "service_management"},
		{"docker başlat", "start", "docker", "service_management"},
		{"redis durdur", "stop", "redis", "service_management"},
		{"açık portlar", "ports", "network", "system_info"},
		{"disk alanı", "disk", "storage", "system_info"},
		{"ram durumu", "memory", "ram", "system_info"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			res := TryLocalMatch(tt.input)
			if res == nil {
				t.Fatalf("expected match for %q, got nil", tt.input)
			}
			if string(res.Action) != tt.wantAction {
				t.Errorf("action: got %q, want %q", res.Action, tt.wantAction)
			}
			if res.Parameters.Target != tt.wantTarget {
				t.Errorf("target: got %q, want %q", res.Parameters.Target, tt.wantTarget)
			}
			if string(res.Category) != tt.wantCategory {
				t.Errorf("category: got %q, want %q", res.Category, tt.wantCategory)
			}
			if len(res.SuggestedCommands) == 0 {
				t.Errorf("expected suggested commands, got none")
			}
		})
	}
}
