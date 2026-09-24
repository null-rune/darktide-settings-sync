package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type AppState struct {
	Window         fyne.Window
	Service        *ModSettingsService
	Detected       []InstallationSource
	ActivePath     string
	ActiveConfig   *ConfigFile
	SourcePath     string
	SourceConfig   *ConfigFile
	SourceMods     map[string]Node
	SelectedMods   map[string]bool
	FilterQuery    string
	ChecklistGroup *fyne.Container
	StatusLabel    *widget.Label
	CountLabel     *widget.Label
}

func main() {
	myApp := app.NewWithID("com.darktide.settingssync")
	window := myApp.NewWindow("Warhammer 40,000: Darktide - Mod Settings Sync")
	window.Resize(fyne.NewSize(820, 720))

	state := &AppState{
		Window:         window,
		Service:        NewModSettingsService(),
		Detected:       DetectSettingsFiles(),
		SelectedMods:   make(map[string]bool),
		ChecklistGroup: container.NewVBox(),
		StatusLabel:    widget.NewLabel("Ready. Select active settings and friend's source file."),
		CountLabel:     widget.NewLabel(""),
	}

	// Active file controls
	activePathEntry := widget.NewEntry()
	activePathEntry.SetPlaceHolder("Path to active user_settings.config...")

	browseActiveBtn := widget.NewButtonWithIcon("Browse Active...", theme.FolderOpenIcon(), func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				path := reader.URI().Path()
				activePathEntry.SetText(path)
				state.loadActiveFile(path)
			}
		}, window)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".config"}))
		fd.Show()
	})

	manualBackupBtn := widget.NewButtonWithIcon("Manual Backup...", theme.DocumentSaveIcon(), func() {
		if state.ActivePath == "" {
			dialog.ShowError(fmt.Errorf("no active settings file selected to backup"), window)
			return
		}
		fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err == nil && writer != nil {
				target := writer.URI().Path()
				if err := ExportBackup(state.ActivePath, target); err != nil {
					dialog.ShowError(err, window)
				} else {
					dialog.ShowInformation("Backup Created", fmt.Sprintf("Backup successfully written to:\n%s", target), window)
				}
			}
		}, window)
		fd.SetFileName("user_settings.config.bak")
		fd.Show()
	})

	var installSelector *widget.Select
	if len(state.Detected) > 0 {
		options := make([]string, len(state.Detected))
		for i, inst := range state.Detected {
			options[i] = fmt.Sprintf("%s (%s)", inst.Platform, inst.Path)
		}

		installSelector = widget.NewSelect(options, func(selected string) {
			for _, inst := range state.Detected {
				desc := fmt.Sprintf("%s (%s)", inst.Platform, inst.Path)
				if desc == selected {
					activePathEntry.SetText(inst.Path)
					state.loadActiveFile(inst.Path)
					break
				}
			}
		})
	}

	// Source file controls
	sourcePathEntry := widget.NewEntry()
	sourcePathEntry.SetPlaceHolder("Select friend's source user_settings.config...")

	browseSourceBtn := widget.NewButtonWithIcon("Browse Friend's File...", theme.FolderOpenIcon(), func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				path := reader.URI().Path()
				sourcePathEntry.SetText(path)
				state.loadSourceFile(path)
			}
		}, window)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".config"}))
		fd.Show()
	})

	// Mod checklist controls
	searchFilter := widget.NewEntry()
	searchFilter.SetPlaceHolder("Filter mods by name...")
	searchFilter.OnChanged = func(q string) {
		state.FilterQuery = strings.ToLower(strings.TrimSpace(q))
		state.refreshChecklist()
	}

	selectAllBtn := widget.NewButton("Select All", func() {
		for m := range state.SourceMods {
			state.SelectedMods[m] = true
		}
		state.refreshChecklist()
	})

	deselectAllBtn := widget.NewButton("Deselect All", func() {
		state.SelectedMods = make(map[string]bool)
		state.refreshChecklist()
	})

	mergeBtn := widget.NewButtonWithIcon("Merge Selected Mod Settings", theme.DocumentSaveIcon(), func() {
		state.promptAndMerge()
	})
	mergeBtn.Importance = widget.HighImportance

	// Build Layout
	activeBox := container.NewVBox(
		widget.NewLabelWithStyle("Active Installation / Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	if len(state.Detected) > 1 {
		activeBox.Add(widget.NewLabel("Multiple game installations detected. Select one to modify:"))
		activeBox.Add(installSelector)
	}

	activeBox.Add(container.NewBorder(nil, nil, nil, container.NewHBox(browseActiveBtn, manualBackupBtn), activePathEntry))

	sourceBox := container.NewVBox(
		widget.NewLabelWithStyle("Source Settings (Friend's File)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(nil, nil, nil, browseSourceBtn, sourcePathEntry),
	)

	checklistHeader := container.NewVBox(
		container.NewBorder(
			nil, nil,
			widget.NewLabelWithStyle("Mod Settings Selection", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewHBox(selectAllBtn, deselectAllBtn),
		),
		container.NewBorder(nil, nil, nil, state.CountLabel, searchFilter),
	)

	checklistScroll := container.NewVScroll(state.ChecklistGroup)
	checklistScroll.SetMinSize(fyne.NewSize(760, 300))

	mainLayout := container.NewBorder(
		container.NewVBox(
			activeBox,
			widget.NewSeparator(),
			sourceBox,
			widget.NewSeparator(),
			checklistHeader,
		),
		container.NewVBox(
			widget.NewSeparator(),
			mergeBtn,
			state.StatusLabel,
		),
		nil,
		nil,
		checklistScroll,
	)

	window.SetContent(container.NewPadded(mainLayout))

	// Pre-select detected active file
	if len(state.Detected) == 1 {
		activePathEntry.SetText(state.Detected[0].Path)
		state.loadActiveFile(state.Detected[0].Path)
	} else if len(state.Detected) > 1 {
		installSelector.SetSelected(fmt.Sprintf("%s (%s)", state.Detected[0].Platform, state.Detected[0].Path))
	}

	window.ShowAndRun()
}

func (s *AppState) loadActiveFile(path string) {
	cfg, err := LoadConfigFile(path)
	if err != nil {
		dialog.ShowError(fmt.Errorf("error reading active config: %w", err), s.Window)
		return
	}
	s.ActivePath = path
	s.ActiveConfig = cfg
	mods := s.Service.GetMods(cfg)
	s.StatusLabel.SetText(fmt.Sprintf("Loaded active settings: %d mod(s) configured.", len(mods)))
	if s.SourceConfig != nil {
		s.refreshChecklist()
	}
}

func (s *AppState) loadSourceFile(path string) {
	cfg, err := LoadConfigFile(path)
	if err != nil {
		dialog.ShowError(fmt.Errorf("error reading source config: %w", err), s.Window)
		return
	}
	s.SourcePath = path
	s.SourceConfig = cfg
	s.SourceMods = s.Service.GetMods(cfg)
	s.SelectedMods = make(map[string]bool)

	// Pre-select all mods by default
	for m := range s.SourceMods {
		s.SelectedMods[m] = true
	}

	s.refreshChecklist()
	s.StatusLabel.SetText(fmt.Sprintf("Loaded source file: found %d mod(s).", len(s.SourceMods)))
}

func (s *AppState) refreshChecklist() {
	s.ChecklistGroup.Objects = nil

	if len(s.SourceMods) == 0 {
		s.ChecklistGroup.Add(widget.NewLabel("No mod configurations found in the source file."))
		s.CountLabel.SetText("")
		s.ChecklistGroup.Refresh()
		return
	}

	activeMods := s.Service.GetMods(s.ActiveConfig)
	allSorted := s.Service.GetSortedModNames(s.SourceMods)

	var filtered []string
	for _, name := range allSorted {
		if s.FilterQuery == "" || strings.Contains(strings.ToLower(name), s.FilterQuery) {
			filtered = append(filtered, name)
		}
	}

	selectedCount := 0
	uninstalledCount := 0
	for m, sel := range s.SelectedMods {
		if sel {
			selectedCount++
			if _, ok := activeMods[m]; !ok {
				uninstalledCount++
			}
		}
	}

	s.CountLabel.SetText(fmt.Sprintf("Showing %d / %d mods (%d selected, %d uninstalled)",
		len(filtered), len(allSorted), selectedCount, uninstalledCount))

	for _, modName := range filtered {
		m := modName
		_, isInstalled := activeMods[m]

		var statusBadge string
		if isInstalled {
			statusBadge = "[✓ Installed]"
		} else {
			statusBadge = "[⚠️ NOT IN YOUR ACTIVE CONFIG]"
		}

		label := fmt.Sprintf("%-35s  %s", m, statusBadge)
		chk := widget.NewCheck(label, func(checked bool) {
			s.SelectedMods[m] = checked
			s.updateCounts()
		})
		chk.SetChecked(s.SelectedMods[m])

		s.ChecklistGroup.Add(chk)
	}

	s.ChecklistGroup.Refresh()
}

func (s *AppState) updateCounts() {
	if s.ActiveConfig == nil || s.SourceMods == nil {
		return
	}
	activeMods := s.Service.GetMods(s.ActiveConfig)
	selectedCount := 0
	uninstalledCount := 0
	for m, sel := range s.SelectedMods {
		if sel {
			selectedCount++
			if _, ok := activeMods[m]; !ok {
				uninstalledCount++
			}
		}
	}
	s.CountLabel.SetText(fmt.Sprintf("%d selected (%d uninstalled)", selectedCount, uninstalledCount))
}

func (s *AppState) promptAndMerge() {
	if s.ActiveConfig == nil || s.ActivePath == "" {
		dialog.ShowError(fmt.Errorf("please select or detect your active user_settings.config file"), s.Window)
		return
	}
	if s.SourceConfig == nil || s.SourcePath == "" {
		dialog.ShowError(fmt.Errorf("please select a source (friend's) user_settings.config file"), s.Window)
		return
	}

	var selected []string
	for mod, isChecked := range s.SelectedMods {
		if isChecked {
			selected = append(selected, mod)
		}
	}

	if len(selected) == 0 {
		dialog.ShowInformation("No Mods Selected", "Please select at least one mod setting to import.", s.Window)
		return
	}

	// Validation: Cross-reference selected mods against active file
	missing := s.Service.CheckMissingMods(s.ActiveConfig, selected)
	if len(missing) > 0 {
		warningText := fmt.Sprintf(
			"⚠️ Uninstalled Mod Warning\n\n"+
				"The following %d selected mod(s) are NOT installed in your active settings:\n\n"+
				"• %s\n\n"+
				"Importing settings for mods you don't have installed may create unused config blocks.\n\n"+
				"Do you want to proceed with importing these settings anyway?",
			len(missing),
			strings.Join(missing, "\n• "),
		)

		dialog.ShowConfirm("Warning: Uninstalled Mods Selected", warningText, func(proceed bool) {
			if proceed {
				s.executeMerge(selected)
			}
		}, s.Window)
		return
	}

	s.executeMerge(selected)
}

func (s *AppState) executeMerge(selected []string) {
	// Step 1: Automatic backup of active file
	bakPath, err := CreateAutoBackup(s.ActivePath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("automatic backup failed, aborting merge: %w", err), s.Window)
		return
	}

	// Step 2: Merge mod settings in AST
	if err := s.Service.MergeModSettings(s.ActiveConfig, s.SourceConfig, selected); err != nil {
		dialog.ShowError(fmt.Errorf("merge failed: %w", err), s.Window)
		return
	}

	// Step 3: Write out updated config
	if err := SaveConfigFile(s.ActivePath, s.ActiveConfig); err != nil {
		dialog.ShowError(fmt.Errorf("failed to save merged settings: %w", err), s.Window)
		return
	}

	s.StatusLabel.SetText(fmt.Sprintf("Successfully merged %d mod(s). Auto-backup: %s", len(selected), bakPath))
	dialog.ShowInformation(
		"Merge Complete",
		fmt.Sprintf("Successfully merged %d mod setting(s) into your configuration!\n\nAn automatic backup was created at:\n%s",
			len(selected), bakPath),
		s.Window,
	)

	// Refresh UI with updated settings
	s.loadActiveFile(s.ActivePath)
}
