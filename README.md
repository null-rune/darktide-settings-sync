# Warhammer 40,000: Darktide - Mod Settings Sync

A standalone desktop GUI application written in Go with the [Fyne](https://fyne.io/) toolkit that helps Warhammer 40k: Darktide players share, inspect, and selectively merge their mod settings.

---

## Features

### 1. Automatic Settings Discovery
- Automatically detects Darktide's active `user_settings.config` across:
  - **Windows (Steam)**: `%APPDATA%\FatShark\Darktide\user_settings.config`
  - **Windows (Microsoft Store / Game Pass)**: `%APPDATA%\FatShark\MicrosoftStore\Darktide\user_settings.config`
  - **Linux (Steam / Proton)**: `~/.steam/steam/steamapps/compatdata/1361210/pfx/drive_c/users/steamuser/AppData/Roaming/Fatshark/Darktide/user_settings.config`
- When multiple installations exist on the machine (e.g., both Steam and Game Pass), a profile selector dropdown allows you to choose which installation to modify.
- Native file picker button ("Browse Active...") allows manual selection of custom paths.

### 2. Robust Backup Functionality
- **Automatic Pre-Modification Backup**: Prior to making any changes, the application automatically creates:
  1. A timestamped backup: `user_settings.config.<timestamp>.bak` in the active configuration directory.
  2. A static backup: `user_settings.config.bak` in the active configuration directory.
- **Manual Backup**: A "Manual Backup..." button allows exporting your active settings file to any directory or drive via a native file save dialog.

### 3. Safe Stingray SJSON / Config Parser
- Darktide configs are written in Fatshark's Stingray SJSON (Simplified JSON), using nested dictionaries `{ ... }` and array blocks `[ ... ]` delimited by newlines without mandatory commas.
- A custom high-fidelity zero-dependency AST parser reads and serializes this exact format.
- Only the selected mods inside the `mods_settings` block are merged or updated; all unrelated game settings (video options, resolution, audio volume, keybindings, input curves) remain 100% untouched.

### 4. Mod Checklist, Search & Missing Mod Validation
- Parses your friend's `user_settings.config` and extracts all mod configurations.
- Displays a scrollable checklist with live search filtering, "Select All", and "Deselect All".
- Cross-references each mod against your active file:
  - `[✓ Installed]`: Mod exists in your configuration.
  - `[⚠️ NOT IN YOUR ACTIVE CONFIG]`: Highlights mods you don't currently have installed.
- **Validation Warning**: If you select uninstalled mods to import, the UI presents an explicit confirmation warning before proceeding to prevent unintended clutter.

---

## Building and Running

The application compiles into a single, standalone executable with zero external runtime dependencies.

### Windows (`windows/amd64`)

#### Option A: Quick Build Script
Double-click or run:
```cmd
build.bat
```

#### Option B: Go Build
```bash
go build -ldflags "-s -w -H windowsgui" -o darktide-settings-sync.exe .
```
*(The `-H windowsgui` flag suppresses the background command prompt window).*

### Linux (`linux/amd64`)

On Linux (ensure standard X11/OpenGL development headers are installed, e.g. `libgl1-mesa-dev xorg-dev`):
```bash
go build -ldflags "-s -w" -o darktide-settings-sync .
```

### Cross-Compilation

To cross-compile for both Windows and Linux from any OS, install `fyne-cross`:
```bash
go install github.com/fyne-io/fyne-cross@latest
fyne-cross windows -arch=amd64
fyne-cross linux -arch=amd64
```

---

## Running the Unit Tests

The test suite validates the parser, mod merger, and discovery logic directly against synthetic cases and real Darktide configuration files:
```bash
go test -v ./...
```
