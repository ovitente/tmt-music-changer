package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

const defaultModPath = `/mnt/d/SteamLibrary/steamapps/common/Terminator Dark Fate - Defiance/mods/GeneralsMusic`

type ThemeFile struct {
	Label   string
	RelPath string
}

var themeFiles = []ThemeFile{
	{Label: "Original", RelPath: "basis/scripts/themes.xml"},
	{Label: "Resistance", RelPath: "dlc/Resistance/basis/scripts/themes.xml"},
	{Label: "Legion", RelPath: "dlc/Legion/basis/scripts/themes.xml"},
}

func main() {
	modPath := flag.String("mod", defaultModPath, "path to mod root")
	flag.Parse()

	var files []FileState
	for _, tf := range themeFiles {
		fp := filepath.Join(*modPath, tf.RelPath)
		data, err := os.ReadFile(fp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot read %s (%s): %v\n", tf.Label, fp, err)
			continue
		}
		missions, err := ParseThemesXML(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot parse %s: %v\n", tf.Label, err)
			continue
		}
		// Store original parse order index on each mission
		for i := range missions {
			missions[i].OrigOrder = i
		}
		sortMissions(missions)
		files = append(files, FileState{
			label:    tf.Label,
			filePath: fp,
			missions: missions,
			xmlBytes: data,
			sorted:   true,
		})
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No valid theme files found")
		os.Exit(1)
	}

	model := NewModel(files)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
