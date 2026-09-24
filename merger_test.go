package main

import (
	"os"
	"testing"
)

func TestMergerLogic(t *testing.T) {
	activeText := `
adapter_index = 0
sound_settings = {
	master_volume = 100
}
mods_settings = {
	ModAlpha = {
		enabled = false
		setting_1 = 10
	}
	ModBeta = {
		volume = 5
	}
}
`

	sourceText := `
adapter_index = 1
sound_settings = {
	master_volume = 50
}
mods_settings = {
	ModAlpha = {
		enabled = true
		setting_1 = 99
	}
	ModGamma = {
		hud_scale = 1.2
	}
}
`

	activeCfg, err := ParseConfigFile(activeText)
	if err != nil {
		t.Fatalf("Failed to parse active: %v", err)
	}

	sourceCfg, err := ParseConfigFile(sourceText)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	svc := NewModSettingsService()

	// 1. Verify mod extraction
	sourceMods := svc.GetMods(sourceCfg)
	if len(sourceMods) != 2 {
		t.Fatalf("Expected 2 source mods, got %d", len(sourceMods))
	}
	if _, ok := sourceMods["ModAlpha"]; !ok {
		t.Fatalf("Expected ModAlpha in source mods")
	}
	if _, ok := sourceMods["ModGamma"]; !ok {
		t.Fatalf("Expected ModGamma in source mods")
	}

	// 2. Verify missing mod detection
	selected := []string{"ModAlpha", "ModGamma"}
	missing := svc.CheckMissingMods(activeCfg, selected)
	if len(missing) != 1 || missing[0] != "ModGamma" {
		t.Fatalf("Expected ModGamma to be detected as missing, got %v", missing)
	}

	// 3. Perform merge of ModAlpha and ModGamma
	err = svc.MergeModSettings(activeCfg, sourceCfg, selected)
	if err != nil {
		t.Fatalf("MergeModSettings failed: %v", err)
	}

	// Verify active config now has ModAlpha updated and ModGamma added, while ModBeta is preserved
	activeMods := svc.GetMods(activeCfg)
	if len(activeMods) != 3 {
		t.Fatalf("Expected 3 active mods after merge, got %d", len(activeMods))
	}

	// Check ModAlpha was updated
	alphaObj, ok := activeMods["ModAlpha"].(*ObjectNode)
	if !ok {
		t.Fatalf("ModAlpha is not an ObjectNode")
	}
	foundUpdated := false
	for _, entry := range alphaObj.Entries {
		if entry.Key == "enabled" {
			if boolNode, ok := entry.Value.(*BoolNode); ok && boolNode.Value {
				foundUpdated = true
			}
		}
	}
	if !foundUpdated {
		t.Fatalf("ModAlpha was not updated with source values")
	}

	// Check ModBeta is intact
	if _, ok := activeMods["ModBeta"]; !ok {
		t.Fatalf("ModBeta was lost during merge")
	}

	// Check non-mod settings are intact
	for _, entry := range activeCfg.Entries {
		if entry.Key == "adapter_index" {
			if numNode, ok := entry.Value.(*NumberNode); !ok || numNode.Value != "0" {
				t.Fatalf("adapter_index corrupted, expected 0, got %v", entry.Value)
			}
		}
	}
}

func TestMergerWithRealFile(t *testing.T) {
	data, err := os.ReadFile("user_settings - raph.config")
	if err != nil {
		t.Skip("user_settings - raph.config not found")
	}

	sourceCfg, err := ParseConfigFile(string(data))
	if err != nil {
		t.Fatalf("Failed to parse real source config: %v", err)
	}

	svc := NewModSettingsService()
	sourceMods := svc.GetMods(sourceCfg)
	if len(sourceMods) < 100 {
		t.Fatalf("Expected >100 mods, got %d", len(sourceMods))
	}

	// Create a mock active config with only 1 mod
	activeCfg, err := ParseConfigFile(`
adapter_index = 0
mods_settings = {
	AccurateCurioNames = {
		show_perks = false
	}
}
`)
	if err != nil {
		t.Fatalf("Failed to parse mock active: %v", err)
	}

	// Select 2 mods: AccurateCurioNames (installed) and "A la Mode" (uninstalled)
	selected := []string{"AccurateCurioNames", "A la Mode"}
	missing := svc.CheckMissingMods(activeCfg, selected)
	if len(missing) != 1 || missing[0] != "A la Mode" {
		t.Fatalf("Expected 'A la Mode' to be detected as missing, got %v", missing)
	}

	// Execute merge
	if err := svc.MergeModSettings(activeCfg, sourceCfg, selected); err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Serialize and re-parse to verify syntax validity
	output := activeCfg.Serialize()
	reparsed, err := ParseConfigFile(output)
	if err != nil {
		t.Fatalf("Re-parsing merged config failed: %v\nOutput snippet:\n%s", err, output[:500])
	}

	reparsedMods := svc.GetMods(reparsed)
	if len(reparsedMods) != 2 {
		t.Fatalf("Expected 2 mods in reparsed active, got %d", len(reparsedMods))
	}
}
