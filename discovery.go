package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type InstallationSource struct {
	Platform string
	Path     string
}

// DetectSettingsFiles searches standard paths for user_settings.config based on OS.
func DetectSettingsFiles() []InstallationSource {
	var candidates []InstallationSource

	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData != "" {
			// Windows (Steam)
			steamPath := filepath.Join(appData, "FatShark", "Darktide", "user_settings.config")
			if fileExists(steamPath) {
				candidates = append(candidates, InstallationSource{
					Platform: "Windows (Steam)",
					Path:     steamPath,
				})
			}

			// Windows (Microsoft Store / Game Pass)
			msStorePath := filepath.Join(appData, "FatShark", "MicrosoftStore", "Darktide", "user_settings.config")
			if fileExists(msStorePath) {
				candidates = append(candidates, InstallationSource{
					Platform: "Windows (Microsoft Store)",
					Path:     msStorePath,
				})
			}
		}
	} else if runtime.GOOS == "linux" {
		home, err := os.UserHomeDir()
		if err == nil {
			// Linux (Steam/Proton prefix)
			protonPath := filepath.Join(home, ".steam", "steam", "steamapps", "compatdata", "1361210",
				"pfx", "drive_c", "users", "steamuser", "AppData", "Roaming", "Fatshark", "Darktide", "user_settings.config")
			if fileExists(protonPath) {
				candidates = append(candidates, InstallationSource{
					Platform: "Linux (Steam/Proton)",
					Path:     protonPath,
				})
			}

			// Alternative Steam root ~/.local/share/Steam/
			altProtonPath := filepath.Join(home, ".local", "share", "Steam", "steamapps", "compatdata", "1361210",
				"pfx", "drive_c", "users", "steamuser", "AppData", "Roaming", "Fatshark", "Darktide", "user_settings.config")
			if fileExists(altProtonPath) && altProtonPath != protonPath {
				candidates = append(candidates, InstallationSource{
					Platform: "Linux (Steam/Proton - Local Share)",
					Path:     altProtonPath,
				})
			}
		}
	}

	return candidates
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// CreateAutoBackup creates an automatic timestamped backup and updates user_settings.config.bak.
func CreateAutoBackup(activeFilePath string) (string, error) {
	if !fileExists(activeFilePath) {
		return "", fmt.Errorf("active settings file not found: %s", activeFilePath)
	}

	dir := filepath.Dir(activeFilePath)
	base := filepath.Base(activeFilePath)
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("%s.%s.bak", base, timestamp)
	backupPath := filepath.Join(dir, backupName)

	if err := CopyFile(activeFilePath, backupPath); err != nil {
		return "", fmt.Errorf("failed to create timestamped backup: %w", err)
	}

	// Also write/update static user_settings.config.bak in the same directory
	staticBakPath := filepath.Join(dir, "user_settings.config.bak")
	_ = CopyFile(activeFilePath, staticBakPath)

	return backupPath, nil
}

// ExportBackup writes a backup copy of the active file to a target destination.
func ExportBackup(activeFilePath, targetFilePath string) error {
	return CopyFile(activeFilePath, targetFilePath)
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
