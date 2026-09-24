# Warhammer 40k: Darktide - Mod Settings Sync

A lightweight, standalone desktop tool for Warhammer 40,000: Darktide players to easily share, inspect, and selectively merge mod settings with friends without overwriting keybindings, graphics, or audio settings.

tiny update
update 2
No installation, runtime, or dependencies required — just a single executable.

---

## 🚀 Quick Start (For Players)

### 1. Download
Go to the **[Releases](https://github.com/null-rune/darktide-settings-sync/releases/latest)** tab on GitHub and download `darktide-settings-sync.exe`.

### 2. Run the App
Double-click `darktide-settings-sync.exe`. 
- **Auto-Detection**: The app automatically detects your active Darktide installation (Steam, Microsoft Store / Xbox Game Pass, or Linux Proton).
- If multiple installations are found, select which profile you want to modify from the dropdown.

### 3. Share & Merge Settings
1. **Load Friend's File**: Click **"Browse Friend's File..."** and select the `user_settings.config` your friend sent you.
2. **Pick Your Mods**: A checklist will display all mods found in their file. You can search, filter, and pick only the specific mods you want to copy (e.g. crosshair colors, scoreboard configs, hud tweaks).
   - Mods you already have installed will display `[✓ Installed]`.
   - Mods you don't have installed will display `[⚠️ NOT IN YOUR ACTIVE CONFIG]`. If you attempt to import settings for a mod you don't have, the app will warn you before proceeding.
3. **Merge**: Click **"Merge Selected Mod Settings"**.
   - An automatic backup (`user_settings.config.<timestamp>.bak`) is saved in your settings folder before any changes are made.
   - Only the selected mod configurations are updated. Your personal resolution, graphics quality, audio volume, keybindings, and other mods remain completely untouched!

---

## 🛡️ Safety & Backups

- **Automatic Backups**: Every time you merge, the tool automatically creates a timestamped copy (`user_settings.config.YYYYMMDD_HHMMSS.bak`) and a standard `user_settings.config.bak` in your Darktide settings folder.
- **Manual Backup**: Click the **"Manual Backup..."** button at any time to export a copy of your active settings to any folder or drive you choose.
- **Stingray SJSON Engine-Safe**: Darktide's config format (Stingray SJSON) is safely parsed into an Abstract Syntax Tree. Only the specified mod sub-blocks are updated, eliminating the risk of config corruption.

---

## 📂 Supported Installation Locations

| Platform | Default Path |
| :--- | :--- |
| **Windows (Steam)** | `%APPDATA%\FatShark\Darktide\user_settings.config` |
| **Windows (MS Store / Xbox App)** | `%APPDATA%\FatShark\MicrosoftStore\Darktide\user_settings.config` |
| **Linux (Steam / Proton)** | `~/.steam/steam/steamapps/compatdata/1361210/pfx/drive_c/users/steamuser/AppData/Roaming/Fatshark/Darktide/user_settings.config` |

*(You can also click "Browse Active..." to select any custom config location).*

---

## 🛠️ For Developers & Building From Source

If you want to modify the code or build the binary yourself:

### Prerequisites
- [Go 1.22+](https://go.dev/)
- A C compiler for Cgo (e.g., MinGW-w64 / GCC on Windows, `gcc` on Linux)

### Build Standalone Binary

#### Windows
Run the included build script:
```cmd
build.bat
```
Or with `go build`:
```bash
go build -ldflags "-s -w -H windowsgui" -o darktide-settings-sync.exe .
```

#### Linux
```bash
go build -ldflags "-s -w" -o darktide-settings-sync .
```

### Cross-Compilation
Use [fyne-cross](https://github.com/fyne-io/fyne-cross):
```bash
go install github.com/fyne-io/fyne-cross@latest
fyne-cross windows -arch=amd64
fyne-cross linux -arch=amd64
```

### Run Tests
```bash
go test -v ./...
```
*(Runs unit tests against synthetic samples and real 14,000+ line Darktide configuration files).*

---

## License

MIT License. See [LICENSE](LICENSE) for details.
