@echo off
echo Building Darktide Settings Sync for Windows...
go build -ldflags "-s -w -H windowsgui" -o darktide-settings-sync.exe .
if %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] Successfully built darktide-settings-sync.exe!
) else (
    echo [ERROR] Build failed with exit code %ERRORLEVEL%.
)
