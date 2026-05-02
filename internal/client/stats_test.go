package client

import (
	"encoding/json"
	"testing"
)

// TestStats_DecodesDashboardShape covers the major reshape in #7 — the
// backend's admin.DashboardStats response replaced the flat total_* counters
// with per-resource sub-objects. The previous client.Stats struct read every
// field as zero against the v1 backend.
func TestStats_DecodesDashboardShape(t *testing.T) {
	body := []byte(`{
		"users": 42,
		"organizations": 7,
		"scm_providers": 3,
		"downloads": 12345,
		"modules": {
			"total": 100,
			"versions": 350,
			"downloads": 5000,
			"by_system": [
				{"system": "aws", "count": 60},
				{"system": "azurerm", "count": 40}
			]
		},
		"providers": {
			"total": 25,
			"total_versions": 120,
			"manual": 5,
			"manual_versions": 30,
			"mirrored": 20,
			"mirrored_versions": 90,
			"downloads": 7000
		},
		"provider_mirrors": {
			"total": 4,
			"healthy": 3,
			"failed": 1
		},
		"binary_mirrors": {
			"total": 2,
			"healthy": 2,
			"failed": 0,
			"syncing": 0,
			"downloads": 345,
			"platforms": 18,
			"by_tool": [
				{"tool": "terraform", "count": 1},
				{"tool": "opentofu", "count": 1}
			]
		},
		"recent_syncs": [
			{
				"mirror_name": "hashicorp",
				"mirror_type": "provider",
				"status": "success",
				"triggered_by": "scheduler",
				"started_at": "2026-04-30T00:00:00Z",
				"completed_at": "2026-04-30T00:05:00Z",
				"versions_synced": 12,
				"platforms_synced": 0
			}
		]
	}`)

	var s Stats
	if err := json.Unmarshal(body, &s); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if s.Users != 42 {
		t.Errorf("Users = %d, want 42", s.Users)
	}
	if s.Modules.Total != 100 {
		t.Errorf("Modules.Total = %d, want 100", s.Modules.Total)
	}
	if s.Modules.Downloads != 5000 {
		t.Errorf("Modules.Downloads = %d, want 5000", s.Modules.Downloads)
	}
	if got := len(s.Modules.BySystem); got != 2 {
		t.Errorf("Modules.BySystem len = %d, want 2", got)
	}
	if s.Providers.Manual != 5 || s.Providers.Mirrored != 20 {
		t.Errorf("Providers manual/mirrored = %d/%d, want 5/20",
			s.Providers.Manual, s.Providers.Mirrored)
	}
	if s.ProviderMirrors.Healthy != 3 {
		t.Errorf("ProviderMirrors.Healthy = %d, want 3", s.ProviderMirrors.Healthy)
	}
	if s.BinaryMirrors.Platforms != 18 {
		t.Errorf("BinaryMirrors.Platforms = %d, want 18", s.BinaryMirrors.Platforms)
	}
	if got := len(s.RecentSyncs); got != 1 {
		t.Fatalf("RecentSyncs len = %d, want 1", got)
	}
	if s.RecentSyncs[0].MirrorType != "provider" {
		t.Errorf("RecentSyncs[0].MirrorType = %q, want provider", s.RecentSyncs[0].MirrorType)
	}
}
