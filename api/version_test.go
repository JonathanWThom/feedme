package api

import "testing"

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{"equal", "1.0.0", "1.0.0", 0},
		{"v1 greater major", "2.0.0", "1.0.0", 1},
		{"v1 lesser major", "1.0.0", "2.0.0", -1},
		{"v1 greater minor", "1.2.0", "1.1.0", 1},
		{"v1 greater patch", "1.0.2", "1.0.1", 1},
		{"with v prefix", "v1.1.0", "v1.0.0", 1},
		{"mixed v prefix", "v1.1.0", "1.0.0", 1},
		{"different lengths", "1.0.0", "1.0", 0},
		{"shorter v1 less", "1.0", "1.0.1", -1},
		{"equal no patch", "1.0", "1.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d",
					tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestUpdateInfo_HasUpdate(t *testing.T) {
	tests := []struct {
		name string
		info *UpdateInfo
		want bool
	}{
		{"nil info", nil, false},
		{"empty latest", &UpdateInfo{CurrentVersion: "1.0.0", LatestVersion: ""}, false},
		{"same version", &UpdateInfo{CurrentVersion: "1.0.0", LatestVersion: "1.0.0"}, false},
		{
			"newer available",
			&UpdateInfo{CurrentVersion: "1.0.0", LatestVersion: "1.1.0"},
			true,
		},
		{
			"older latest (downgrade)",
			&UpdateInfo{CurrentVersion: "2.0.0", LatestVersion: "1.0.0"},
			false,
		},
		{
			"with v prefix",
			&UpdateInfo{CurrentVersion: "v1.0.0", LatestVersion: "v1.1.0"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.info.HasUpdate()
			if got != tt.want {
				t.Errorf("HasUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateInfo_FormatUpdateMessage(t *testing.T) {
	tests := []struct {
		name string
		info *UpdateInfo
		want string
	}{
		{
			"has update",
			&UpdateInfo{CurrentVersion: "1.0.0", LatestVersion: "1.1.0"},
			"Update available: 1.0.0 → 1.1.0",
		},
		{
			"no update",
			&UpdateInfo{CurrentVersion: "1.0.0", LatestVersion: "1.0.0"},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.info.FormatUpdateMessage()
			if got != tt.want {
				t.Errorf("FormatUpdateMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}
