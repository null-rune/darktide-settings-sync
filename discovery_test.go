package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupFunctionality(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dt_backup_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "user_settings.config")
	content := "adapter_index = 0\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test automatic backup
	bakPath, err := CreateAutoBackup(testFile)
	if err != nil {
		t.Fatalf("CreateAutoBackup failed: %v", err)
	}

	if !fileExists(bakPath) {
		t.Fatalf("Expected backup file %s to exist", bakPath)
	}

	staticBak := filepath.Join(tempDir, "user_settings.config.bak")
	if !fileExists(staticBak) {
		t.Fatalf("Expected static backup %s to exist", staticBak)
	}

	// Test manual export backup
	exportDest := filepath.Join(tempDir, "my_custom_backup.config")
	if err := ExportBackup(testFile, exportDest); err != nil {
		t.Fatalf("ExportBackup failed: %v", err)
	}
	if !fileExists(exportDest) {
		t.Fatalf("Expected exported backup %s to exist", exportDest)
	}
}

func TestDetectSettingsFiles(t *testing.T) {
	// Should execute without crashing regardless of whether files are on this machine
	detected := DetectSettingsFiles()
	t.Logf("Detected %d installation(s)", len(detected))
	for _, inst := range detected {
		t.Logf("Found %s at %s", inst.Platform, inst.Path)
	}
}
