package main

import (
	"os"
	"testing"
)

func TestParseSampleSnippets(t *testing.T) {
	snippet := `
adapter_index = 0
aspect_ratio = -1
borderless_fullscreen = false
last_fullscreen_resolution = [
	1920
	1080
]
master_render_settings = {
	dlss = 0
	graphics_quality = "high"
	ffx_frame_gen = 1
}
mods_settings = {
	"A la Mode" = {
		alm_open_setup = false
		"content/items/weapons/melee" = [
			255
			169
		]
	}
	AccurateCurioNames = {
		show_perks = true
	}
}
`

	cfg, err := ParseConfigFile(snippet)
	if err != nil {
		t.Fatalf("Failed to parse snippet: %v", err)
	}

	if len(cfg.Entries) != 6 {
		t.Fatalf("Expected 6 entries, got %d", len(cfg.Entries))
	}

	// Verify serialization round-trip
	serialized := cfg.Serialize()
	cfg2, err := ParseConfigFile(serialized)
	if err != nil {
		t.Fatalf("Failed to re-parse serialized config: %v", err)
	}

	if len(cfg2.Entries) != 6 {
		t.Fatalf("Expected 6 entries in re-parsed config, got %d", len(cfg2.Entries))
	}
}

func TestParseFullRealConfigFile(t *testing.T) {
	samplePath := "user_settings - raph.config"
	data, err := os.ReadFile(samplePath)
	if err != nil {
		t.Skipf("Sample file %s not found, skipping full test", samplePath)
	}

	cfg, err := ParseConfigFile(string(data))
	if err != nil {
		t.Fatalf("Failed to parse real Darktide config file (%d bytes): %v", len(data), err)
	}

	t.Logf("Successfully parsed %d top-level sections from %s", len(cfg.Entries), samplePath)

	// Check that mods_settings exists and count mods
	var modsCount int
	for _, entry := range cfg.Entries {
		if entry.Key == "mods_settings" {
			if obj, ok := entry.Value.(*ObjectNode); ok {
				modsCount = len(obj.Entries)
			}
		}
	}

	if modsCount == 0 {
		t.Fatalf("Expected to find mods in mods_settings, found 0")
	}

	t.Logf("Found %d mods inside mods_settings!", modsCount)
}
