package main

import (
	"fmt"
	"os"
	"sort"
)

// ModInfo stores information about a mod extracted from configuration
type ModInfo struct {
	Name        string
	Node        Node
	IsInstalled bool // True if present in the user's active config
}

type ModSettingsService struct{}

func NewModSettingsService() *ModSettingsService {
	return &ModSettingsService{}
}

// GetMods extracts all mods configured in mods_settings
func (s *ModSettingsService) GetMods(cfg *ConfigFile) map[string]Node {
	mods := make(map[string]Node)
	if cfg == nil {
		return mods
	}

	for _, entry := range cfg.Entries {
		if entry.Key == "mods_settings" {
			if obj, ok := entry.Value.(*ObjectNode); ok {
				for _, modEntry := range obj.Entries {
					mods[modEntry.Key] = modEntry.Value
				}
			}
		}
	}
	return mods
}

// GetSortedModNames returns mod names sorted alphabetically
func (s *ModSettingsService) GetSortedModNames(mods map[string]Node) []string {
	keys := make([]string, 0, len(mods))
	for k := range mods {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// CheckMissingMods returns which selected mods do not exist in the active config
func (s *ModSettingsService) CheckMissingMods(activeCfg *ConfigFile, selectedMods []string) []string {
	activeMods := s.GetMods(activeCfg)
	var missing []string

	for _, modName := range selectedMods {
		if _, exists := activeMods[modName]; !exists {
			missing = append(missing, modName)
		}
	}
	return missing
}

// MergeModSettings copies selected mod configurations from sourceCfg into activeCfg
func (s *ModSettingsService) MergeModSettings(activeCfg, sourceCfg *ConfigFile, selectedMods []string) error {
	if activeCfg == nil || sourceCfg == nil {
		return fmt.Errorf("active and source configs cannot be nil")
	}

	sourceMods := s.GetMods(sourceCfg)
	selectedSet := make(map[string]bool)
	for _, m := range selectedMods {
		selectedSet[m] = true
	}

	var modsSection *ObjectNode
	var modsSectionIdx = -1

	for idx, entry := range activeCfg.Entries {
		if entry.Key == "mods_settings" {
			if obj, ok := entry.Value.(*ObjectNode); ok {
				modsSection = obj
				modsSectionIdx = idx
				break
			}
		}
	}

	// If active file didn't have mods_settings yet, create it
	if modsSection == nil {
		modsSection = &ObjectNode{}
		activeCfg.Entries = append(activeCfg.Entries, ObjectEntry{
			Key:       "mods_settings",
			KeyQuoted: false,
			Value:     modsSection,
		})
		modsSectionIdx = len(activeCfg.Entries) - 1
	}

	// Update or append selected mods
	for modName := range selectedSet {
		srcNode, hasSrc := sourceMods[modName]
		if !hasSrc {
			continue
		}

		replaced := false
		for i, entry := range modsSection.Entries {
			if entry.Key == modName {
				modsSection.Entries[i].Value = srcNode.Clone()
				replaced = true
				break
			}
		}

		if !replaced {
			modsSection.Entries = append(modsSection.Entries, ObjectEntry{
				Key:       modName,
				KeyQuoted: needsQuoting(modName),
				Value:     srcNode.Clone(),
			})
		}
	}

	activeCfg.Entries[modsSectionIdx].Value = modsSection
	return nil
}

func LoadConfigFile(path string) (*ConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseConfigFile(string(data))
}

func SaveConfigFile(path string, cfg *ConfigFile) error {
	serialized := cfg.Serialize()
	return os.WriteFile(path, []byte(serialized), 0644)
}
