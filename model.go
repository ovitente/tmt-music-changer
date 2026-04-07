package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Mode int

const (
	ModeView    Mode = iota // default: show current track list
	ModeAdd                 // checkbox panel for toggling tracks
	ModeReorder             // move/delete tracks
	ModeEdit                // inline edit/add track name
)

type Panel int

const (
	PanelLeft Panel = iota
	PanelRight
)

type FileState struct {
	label    string
	filePath string
	missions []Mission
	xmlBytes []byte
	dirty    bool
	sorted bool // true = sorted by number, false = original XML order

	missionCursor int
	missionScroll int
	checkCursor   int
	reorderCursor int
}

type Model struct {
	files      []FileState
	activeFile int

	mode        Mode
	activePanel Panel

	// Filter state
	filterActive    bool   // filter is on
	filterInput     string // filter text
	filterTyping    bool   // true = typing, false = navigating
	filterByMission bool   // Tab: track/mission
	filterIndices   []int  // indices into f.missions
	filterCursor    int    // position in filterIndices
	filterScroll    int    // scroll offset for filterIndices

	width  int
	height int

	statusMsg      string
	confirmQuit    bool
	confirmRestore bool
	confirmSwitch  bool
	pendingSwitchIdx int

	// Edit mode state
	editInput  string
	editIsNew  bool // true = add new track, false = edit existing
	editCursor int  // cursor position in editInput
}

type saveResultMsg struct{ err error }
type switchAfterSaveMsg struct{ idx int }

func NewModel(files []FileState) Model {
	return Model{
		files:       files,
		activePanel: PanelLeft,
		mode:        ModeView,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) file() *FileState {
	return &m.files[m.activeFile]
}

// missionModified returns true if mission themes differ from original.
func missionModified(mission *Mission) bool {
	if len(mission.Themes) != len(mission.OriginalThemes) {
		return true
	}
	for i, t := range mission.Themes {
		if t != mission.OriginalThemes[i] {
			return true
		}
	}
	return false
}

// recalcDirty recalculates the dirty flag for the active file.
func (m *Model) recalcDirty() {
	f := m.file()
	f.dirty = false
	for i := range f.missions {
		if missionModified(&f.missions[i]) {
			f.dirty = true
			return
		}
	}
}

func (m Model) currentMission() *Mission {
	f := &m.files[m.activeFile]
	if f.missionCursor >= 0 && f.missionCursor < len(f.missions) {
		return &f.missions[f.missionCursor]
	}
	return nil
}

func (m *Model) toggleUSATrack(idx int) {
	tracks := AllTracks()
	if idx < 0 || idx >= len(tracks) {
		return
	}
	mission := m.currentMission()
	if mission == nil {
		return
	}

	track := "themes/" + tracks[idx]
	isMute := idx >= len(USATracks)

	if isMute {
		// Mute tracks: always append a copy
		mission.Themes = append(mission.Themes, track)
		m.recalcDirty()
		return
	}

	// USA tracks: toggle on/off
	for i, t := range mission.Themes {
		if t == track {
			mission.Themes = append(mission.Themes[:i], mission.Themes[i+1:]...)
			m.recalcDirty()
			return
		}
	}
	mission.Themes = append(mission.Themes, track)
	m.recalcDirty()
}

func trackCount(mission *Mission, track string) int {
	if mission == nil {
		return 0
	}
	full := "themes/" + track
	count := 0
	for _, t := range mission.Themes {
		if t == full {
			count++
		}
	}
	return count
}

func (m *Model) deleteTrack(idx int) {
	mission := m.currentMission()
	if mission == nil || idx < 0 || idx >= len(mission.Themes) {
		return
	}
	mission.Themes = append(mission.Themes[:idx], mission.Themes[idx+1:]...)
	m.recalcDirty()
	if m.file().reorderCursor >= len(mission.Themes) && m.file().reorderCursor > 0 {
		m.file().reorderCursor--
	}
}

// allUniqueTracks returns all unique track names across all missions.
func (m *Model) allUniqueTracks() []string {
	seen := make(map[string]bool)
	var tracks []string
	for _, mission := range m.file().missions {
		for _, t := range mission.Themes {
			if !seen[t] {
				seen[t] = true
				tracks = append(tracks, t)
			}
		}
	}
	return tracks
}

// fuzzyMatch returns true if query chars appear in s in order (case-insensitive).
func fuzzyMatch(s, query string) bool {
	s = strings.ToLower(s)
	query = strings.ToLower(query)
	qi := 0
	for i := 0; i < len(s) && qi < len(query); i++ {
		if s[i] == query[qi] {
			qi++
		}
	}
	return qi == len(query)
}

// updateFilterIndices recalculates filtered mission indices based on current query.
func (m *Model) updateFilterIndices() {
	m.filterIndices = nil
	if m.filterInput == "" {
		// Empty query: show all missions
		for i := range m.file().missions {
			m.filterIndices = append(m.filterIndices, i)
		}
	} else if m.filterByMission {
		for i, mission := range m.file().missions {
			if fuzzyMatch(mission.Name, m.filterInput) {
				m.filterIndices = append(m.filterIndices, i)
			}
		}
	} else {
		// Filter by track name
		matchedTracks := make(map[string]bool)
		for _, t := range m.allUniqueTracks() {
			name := t
			if idx := strings.LastIndex(t, "/"); idx != -1 {
				name = t[idx+1:]
			}
			if fuzzyMatch(name, m.filterInput) {
				matchedTracks[t] = true
			}
		}
		for i, mission := range m.file().missions {
			for _, t := range mission.Themes {
				if matchedTracks[t] {
					m.filterIndices = append(m.filterIndices, i)
					break
				}
			}
		}
	}

	m.filterCursor = 0
	m.filterScroll = 0
	if len(m.filterIndices) > 0 {
		m.file().missionCursor = m.filterIndices[0]
	}
}

// clearFilter resets filter state.
func (m *Model) clearFilter() {
	m.filterActive = false
	m.filterInput = ""
	m.filterTyping = false
	m.filterIndices = nil
	m.filterCursor = 0
	m.filterScroll = 0
}

func (m *Model) moveTrack(from, to int) {
	mission := m.currentMission()
	if mission == nil {
		return
	}
	if from < 0 || from >= len(mission.Themes) || to < 0 || to >= len(mission.Themes) {
		return
	}
	mission.Themes[from], mission.Themes[to] = mission.Themes[to], mission.Themes[from]
	m.recalcDirty()
}

func (m *Model) switchToFile(idx int) {
	m.activeFile = idx
	m.mode = ModeView
	m.activePanel = PanelLeft
	m.clearFilter()
}

func (m Model) anyDirty() bool {
	for _, f := range m.files {
		if f.dirty {
			return true
		}
	}
	return false
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case saveResultMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error saving: %v", msg.err)
		} else {
			// Update stored bytes so repeated saves are correct
			f := m.file()
			f.xmlBytes = SerializeThemesXML(f.xmlBytes, f.missions)
			// Reset baselines so dots clear and Restore uses saved state
			for i := range f.missions {
				f.missions[i].OriginalThemes = make([]string, len(f.missions[i].Themes))
				copy(f.missions[i].OriginalThemes, f.missions[i].Themes)
			}
			f.dirty = false
			m.statusMsg = "Saved!"
		}
		return m, nil

	case switchAfterSaveMsg:
		m.switchToFile(msg.idx)
		return m, nil

	case tea.KeyMsg:
		// Clear status on any key
		if m.statusMsg != "" && !key.Matches(msg, keys.Save) {
			m.statusMsg = ""
		}

		// Handle confirm switch (dirty file)
		if m.confirmSwitch {
			switch msg.String() {
			case "y", "Y":
				// Save then switch
				idx := m.pendingSwitchIdx
				m.confirmSwitch = false
				return m, tea.Sequence(m.saveCmd(), func() tea.Msg {
					return switchAfterSaveMsg{idx: idx}
				})
			case "n", "N":
				// Switch without saving
				m.switchToFile(m.pendingSwitchIdx)
				m.confirmSwitch = false
			default:
				m.confirmSwitch = false
				m.statusMsg = ""
			}
			return m, nil
		}

		// Handle confirm quit
		if m.confirmQuit {
			switch msg.String() {
			case "y", "Y":
				return m, tea.Quit
			default:
				m.confirmQuit = false
				m.statusMsg = ""
				return m, nil
			}
		}

		// Handle confirm restore
		if m.confirmRestore {
			switch msg.String() {
			case "y", "Y":
				mission := m.currentMission()
				if mission != nil {
					mission.Themes = make([]string, len(mission.OriginalThemes))
					copy(mission.Themes, mission.OriginalThemes)
					m.recalcDirty()
				}
			}
			m.confirmRestore = false
			return m, nil
		}

		// File switch (only in ModeView)
		if key.Matches(msg, keys.FileSwitch) && m.mode == ModeView && len(m.files) > 1 {
			nextIdx := (m.activeFile + 1) % len(m.files)
			if m.file().dirty {
				m.confirmSwitch = true
				m.pendingSwitchIdx = nextIdx
				m.statusMsg = "Unsaved changes! Y=save & switch, N=switch without saving, other=cancel"
			} else {
				m.switchToFile(nextIdx)
			}
			return m, nil
		}

		// Filter typing and edit mode handle their own keys
		if m.filterActive && m.filterTyping {
			return m.updateFilterTyping(msg)
		}
		if m.mode == ModeEdit {
			return m.updateEdit(msg)
		}

		// Restore current mission
		if key.Matches(msg, keys.Restore) {
			mission := m.currentMission()
			if mission != nil {
				m.confirmRestore = true
				m.statusMsg = fmt.Sprintf("Restore \"%s\" to original? Y/N", mission.Name)
			}
			return m, nil
		}

		// Save
		if key.Matches(msg, keys.Save) {
			return m, m.saveCmd()
		}

		// Quit
		if key.Matches(msg, keys.Quit) {
			if m.mode != ModeView || m.activePanel == PanelRight {
				m.mode = ModeView
				m.activePanel = PanelLeft
				return m, nil
			}
			if m.anyDirty() {
				m.confirmQuit = true
				m.statusMsg = "Unsaved changes! Press Y to quit, any other key to cancel"
				return m, nil
			}
			return m, tea.Quit
		}

		switch m.mode {
		case ModeView:
			return m.updateView(msg)
		case ModeAdd:
			return m.updateAdd(msg)
		case ModeReorder:
			return m.updateReorder(msg)
		case ModeEdit:
			return m.updateEdit(msg)
		}
	}
	return m, nil
}

func (m Model) updateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	f := m.file()
	switch {
	case key.Matches(msg, keys.Escape):
		if m.filterActive && !m.filterTyping {
			m.clearFilter()
		} else if m.activePanel == PanelRight {
			m.activePanel = PanelLeft
		}

	case key.Matches(msg, keys.Up):
		if m.activePanel == PanelLeft {
			if m.filterActive {
				if m.filterCursor > 0 {
					m.filterCursor--
					f.missionCursor = m.filterIndices[m.filterCursor]
				}
			} else if f.missionCursor > 0 {
				f.missionCursor--
			}
		}

	case key.Matches(msg, keys.Down):
		if m.activePanel == PanelLeft {
			if m.filterActive {
				if m.filterCursor < len(m.filterIndices)-1 {
					m.filterCursor++
					f.missionCursor = m.filterIndices[m.filterCursor]
				}
			} else if f.missionCursor < len(f.missions)-1 {
				f.missionCursor++
			}
		}

	case key.Matches(msg, keys.Toggle):
		m.mode = ModeAdd
		f.checkCursor = 0
		m.activePanel = PanelRight

	case key.Matches(msg, keys.Order):
		mission := m.currentMission()
		if mission != nil && len(mission.Themes) > 0 {
			m.mode = ModeReorder
			f.reorderCursor = 0
			m.activePanel = PanelRight
		}

	case key.Matches(msg, keys.SortToggle):
		currentName := ""
		if mission := m.currentMission(); mission != nil {
			currentName = mission.Name
		}
		if f.sorted {
			sortByOrigOrder(f.missions)
			f.sorted = false
		} else {
			sortMissions(f.missions)
			f.sorted = true
		}
		if m.filterActive {
			m.updateFilterIndices()
		}
		// Find same mission in new order
		f.missionCursor = 0
		f.missionScroll = 0
		if currentName != "" {
			for i, mis := range f.missions {
				if mis.Name == currentName {
					f.missionCursor = i
					break
				}
			}
			if m.filterActive {
				for i, idx := range m.filterIndices {
					if idx == f.missionCursor {
						m.filterCursor = i
						break
					}
				}
			}
		}

	case key.Matches(msg, keys.Search):
		m.filterActive = true
		m.filterTyping = true
		m.filterInput = ""
		m.updateFilterIndices()
	}
	return m, nil
}

// updateFilterTyping handles keyboard input while typing a filter query.
func (m Model) updateFilterTyping(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.clearFilter()

	case key.Matches(msg, keys.Enter):
		if len(m.filterIndices) > 0 {
			m.filterTyping = false
		}

	case key.Matches(msg, keys.Tab):
		m.filterByMission = !m.filterByMission
		m.updateFilterIndices()

	case msg.Type == tea.KeyBackspace:
		if len(m.filterInput) > 0 {
			m.filterInput = m.filterInput[:len(m.filterInput)-1]
			m.updateFilterIndices()
		}

	case msg.Type == tea.KeyRunes:
		m.filterInput += string(msg.Runes)
		m.updateFilterIndices()
	}
	return m, nil
}

func (m Model) updateAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	f := m.file()
	switch {
	case key.Matches(msg, keys.Escape):
		m.mode = ModeView
		m.activePanel = PanelLeft

	case key.Matches(msg, keys.Up):
		if f.checkCursor > 0 {
			f.checkCursor--
		}

	case key.Matches(msg, keys.Down):
		if f.checkCursor < len(AllTracks())-1 {
			f.checkCursor++
		}

	case key.Matches(msg, keys.Space):
		m.toggleUSATrack(f.checkCursor)

	case key.Matches(msg, keys.Order):
		mission := m.currentMission()
		if mission != nil && len(mission.Themes) > 0 {
			m.mode = ModeReorder
			f.reorderCursor = 0
		}

	case key.Matches(msg, keys.AddTrack):
		m.editInput = ""
		m.editIsNew = true
		m.editCursor = 0
		m.mode = ModeEdit
	}
	return m, nil
}

func (m Model) updateReorder(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	mission := m.currentMission()
	if mission == nil {
		return m, nil
	}
	f := m.file()

	switch {
	case key.Matches(msg, keys.Escape):
		m.mode = ModeView
		m.activePanel = PanelLeft

	case key.Matches(msg, keys.Up):
		if f.reorderCursor > 0 {
			f.reorderCursor--
		}

	case key.Matches(msg, keys.Down):
		if f.reorderCursor < len(mission.Themes)-1 {
			f.reorderCursor++
		}

	case key.Matches(msg, keys.MoveUp):
		if f.reorderCursor > 0 {
			m.moveTrack(f.reorderCursor, f.reorderCursor-1)
			f.reorderCursor--
		}

	case key.Matches(msg, keys.MoveDown):
		if f.reorderCursor < len(mission.Themes)-1 {
			m.moveTrack(f.reorderCursor, f.reorderCursor+1)
			f.reorderCursor++
		}

	case key.Matches(msg, keys.Delete):
		if len(mission.Themes) > 0 {
			m.deleteTrack(f.reorderCursor)
			if len(mission.Themes) == 0 {
				m.mode = ModeView
			}
		}

	case key.Matches(msg, keys.Toggle):
		m.mode = ModeAdd
		f.checkCursor = 0
		m.activePanel = PanelRight

	case key.Matches(msg, keys.Edit):
		if len(mission.Themes) > 0 {
			track := strings.TrimPrefix(mission.Themes[f.reorderCursor], "themes/")
			m.editInput = track
			m.editIsNew = false
			m.editCursor = len(track)
			m.mode = ModeEdit
		}

	case key.Matches(msg, keys.AddTrack):
		m.editInput = ""
		m.editIsNew = true
		m.editCursor = 0
		m.mode = ModeEdit
	}
	return m, nil
}

func (m Model) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.mode = ModeReorder

	case key.Matches(msg, keys.Enter):
		if m.editInput != "" {
			f := m.file()
			mission := m.currentMission()
			if mission != nil {
				value := "themes/" + m.editInput
				if m.editIsNew {
					mission.Themes = append(mission.Themes, value)
					f.reorderCursor = len(mission.Themes) - 1
				} else {
					mission.Themes[f.reorderCursor] = value
				}
				m.recalcDirty()
			}
		}
		m.mode = ModeReorder

	case msg.Type == tea.KeyBackspace:
		if m.editCursor > 0 {
			m.editInput = m.editInput[:m.editCursor-1] + m.editInput[m.editCursor:]
			m.editCursor--
		}

	case msg.Type == tea.KeyLeft:
		if m.editCursor > 0 {
			m.editCursor--
		}

	case msg.Type == tea.KeyRight:
		if m.editCursor < len(m.editInput) {
			m.editCursor++
		}

	case msg.Type == tea.KeyRunes:
		ch := string(msg.Runes)
		m.editInput = m.editInput[:m.editCursor] + ch + m.editInput[m.editCursor:]
		m.editCursor += len(ch)
	}
	return m, nil
}

func (m Model) saveCmd() tea.Cmd {
	f := m.files[m.activeFile]
	return func() tea.Msg {
		result := SerializeThemesXML(f.xmlBytes, f.missions)
		err := os.WriteFile(f.filePath, result, 0644)
		if err != nil {
			return saveResultMsg{err: err}
		}
		return saveResultMsg{err: nil}
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	totalWidth := m.width
	// Each panel has: 1 border left + 1 pad left + content + 1 pad right + 1 border right = content + 4
	// Plus 1 gap between panels. Total overhead = 4 + 1 + 4 = 9
	leftWidth := (totalWidth - 9) / 4
	rightWidth := totalWidth - 9 - leftWidth
	if leftWidth < 20 {
		leftWidth = 20
	}
	if rightWidth < 25 {
		rightWidth = 25
	}

	// Title + status bar take 4 lines (title 1 + panels gap 0 + status border 3)
	contentHeight := m.height - 6
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Title
	title := appTitleStyle.Render("Music Themes Changer")

	var leftPanel, rightPanel string

	leftPanel = m.renderMissionList(leftWidth, contentHeight)
	switch m.mode {
	case ModeView:
		rightPanel = m.renderViewPanel(rightWidth, contentHeight)
	case ModeAdd:
		rightPanel = m.renderCheckboxPanel(rightWidth, contentHeight)
	case ModeReorder:
		rightPanel = m.renderReorderPanel(rightWidth, contentHeight)
	case ModeEdit:
		rightPanel = m.renderEditPanel(rightWidth, contentHeight)
	}

	// Apply panel styles
	leftStyle := panelStyle.Width(leftWidth).Height(contentHeight)
	rightStyle := panelStyle.Width(rightWidth).Height(contentHeight)
	if m.filterActive && m.filterTyping {
		leftStyle = activePanelStyle.Width(leftWidth).Height(contentHeight)
	} else if m.mode == ModeAdd || m.mode == ModeReorder || m.mode == ModeEdit {
		rightStyle = activePanelStyle.Width(rightWidth).Height(contentHeight)
	} else {
		leftStyle = activePanelStyle.Width(leftWidth).Height(contentHeight)
	}

	panels := lipgloss.JoinHorizontal(lipgloss.Top,
		leftStyle.Render(leftPanel),
		" ",
		rightStyle.Render(rightPanel),
	)

	panelsWidth := lipgloss.Width(panels)
	status := m.renderStatus(panelsWidth)

	screen := lipgloss.JoinVertical(lipgloss.Left, title, panels, status)

	return screen
}

func (m Model) renderMissionList(width, height int) string {
	var sb strings.Builder
	f := &m.files[m.activeFile]
	textWidth := width - 2 // panel padding

	// Header
	if m.filterActive {
		modeLabel := "Track"
		if m.filterByMission {
			modeLabel = "Mission"
		}
		header := titleStyle.Render("Find") + helpSepStyle.Render(" | ") + titleStyle.Render(modeLabel)
		count := fmt.Sprintf("%d", len(m.filterIndices)) + helpSepStyle.Render("/") + fmt.Sprintf("%d", len(f.missions))
		pad := textWidth - lipgloss.Width(header) - lipgloss.Width(count)
		if pad < 1 {
			pad = 1
		}
		sb.WriteString(header + strings.Repeat(" ", pad) + count)
		sb.WriteString("\n")

		if m.filterTyping {
			sb.WriteString(searchInputStyle.Render(m.filterInput + "█"))
		} else {
			sb.WriteString(checkedStyle.Render(m.filterInput + " ✓"))
		}
		sb.WriteString("\n")
	} else {
		title := "Missions"
		if f.dirty {
			title += " [modified]"
		}
		count := fmt.Sprintf("%d", len(f.missions))
		pad := textWidth - len(title) - len(count)
		if pad < 1 {
			pad = 1
		}
		sb.WriteString(titleStyle.Render(title) + strings.Repeat(" ", pad) + titleStyle.Render(count))
		sb.WriteString("\n\n")
	}

	// Determine visible missions
	var indices []int
	if m.filterActive {
		indices = m.filterIndices
	} else {
		indices = make([]int, len(f.missions))
		for i := range f.missions {
			indices[i] = i
		}
	}

	visibleHeight := height - 3
	if m.filterActive {
		visibleHeight = height - 3 // header takes 2 lines + gap
	}
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	// Scroll management
	var cursor *int
	var scroll *int
	if m.filterActive {
		cursor = &m.filterCursor
		scroll = &m.filterScroll
	} else {
		cursor = &f.missionCursor
		scroll = &f.missionScroll
	}

	if *cursor < *scroll {
		*scroll = *cursor
	}
	if *cursor >= *scroll+visibleHeight {
		*scroll = *cursor - visibleHeight + 1
	}

	for vi := *scroll; vi < len(indices) && vi < *scroll+visibleHeight; vi++ {
		idx := indices[vi]
		name := f.missions[idx].Name
		if len(name) > textWidth-2 {
			name = name[:textWidth-5] + "..."
		}

		hasUSA := false
		for _, t := range f.missions[idx].Themes {
			for _, u := range USATracks {
				if t == "themes/"+u {
					hasUSA = true
					break
				}
			}
			if hasUSA {
				break
			}
		}

		modified := missionModified(&f.missions[idx])

		gutter := " "
		if modified {
			gutter = activeTabStyle.Render("·")
		}

		isSelected := false
		if m.filterActive {
			isSelected = vi == m.filterCursor
		} else {
			isSelected = idx == f.missionCursor
		}

		if isSelected {
			name = reorderHighlight.Render(name)
			sb.WriteString(selectorPrefix.Render("▶ ") + name + "\n")
		} else {
			if hasUSA {
				name = checkedStyle.Render(name)
			}
			sb.WriteString(gutter + " " + name + "\n")
		}
	}

	return sb.String()
}

func (m Model) renderViewPanel(width, height int) string {
	var sb strings.Builder

	mission := m.currentMission()
	if mission == nil {
		sb.WriteString(titleStyle.Render("No mission selected"))
		return sb.String()
	}

	sb.WriteString(titleStyle.Render("Tracks"))
	sb.WriteString("\n\n")

	if len(mission.Themes) == 0 {
		sb.WriteString(commentStyle.Render("  (empty)"))
		sb.WriteString("\n")
	}

	// Reserve space for comment at bottom
	commentLines := 0
	if mission.Comment != "" {
		commentLines = 2 // blank line + comment line
	}
	visibleHeight := height - 3 - commentLines
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	for i := 0; i < len(mission.Themes) && i < visibleHeight; i++ {
		raw := mission.Themes[i]
		theme := strings.TrimPrefix(raw, "themes/")
		if len(theme) > width-4 {
			theme = theme[:width-7] + "..."
		}

		// Highlight USA tracks
		isUSA := false
		for _, u := range USATracks {
			if raw == "themes/"+u {
				isUSA = true
				break
			}
		}

		idx := fmt.Sprintf("%2d. ", i+1)
		if isUSA {
			sb.WriteString(checkedStyle.Render(idx + theme))
		} else {
			sb.WriteString(idx + theme)
		}
		sb.WriteString("\n")
	}

	// Comment at bottom left
	if mission.Comment != "" {
		// Pad to push comment to bottom
		tracksRendered := len(mission.Themes)
		if tracksRendered > visibleHeight {
			tracksRendered = visibleHeight
		}
		remaining := visibleHeight - tracksRendered
		for i := 0; i < remaining; i++ {
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
		sb.WriteString("  Comment " + commentStyle.Render(mission.Comment))
	}

	return sb.String()
}

func (m Model) renderCheckboxPanel(width, height int) string {
	var sb strings.Builder
	f := &m.files[m.activeFile]

	mission := m.currentMission()

	sb.WriteString(titleStyle.Render("Toggle"))
	sb.WriteString("\n\n")

	tracks := AllTracks()
	for i, track := range tracks {
		isMute := i >= len(USATracks)

		if i == len(USATracks) {
			sb.WriteString(commentStyle.Render("  ── pauses ──"))
			sb.WriteString("\n")
		}

		isCursor := i == f.checkCursor && m.activePanel == PanelRight
		prefix := "  "
		if isCursor {
			prefix = selectorPrefix.Render("▶ ")
		}

		var line string
		var hasValue bool
		if isMute {
			count := trackCount(mission, track)
			if count > 0 {
				line = fmt.Sprintf("%s[%dx] %s", prefix, count, track)
				hasValue = true
			} else {
				line = fmt.Sprintf("%s[ ] %s", prefix, track)
			}
		} else {
			checked := " "
			if mission != nil && HasUSATrack(mission.Themes, track) {
				checked = "x"
				hasValue = true
			}
			line = fmt.Sprintf("%s[%s] %s", prefix, checked, track)
		}

		if isCursor {
			sb.WriteString(reorderHighlight.Render(line))
		} else if hasValue {
			sb.WriteString(checkedStyle.Render(line))
		} else {
			sb.WriteString(line)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderReorderPanel(width, height int) string {
	var sb strings.Builder
	f := &m.files[m.activeFile]

	mission := m.currentMission()
	if mission == nil {
		return "No mission selected"
	}

	sb.WriteString(titleStyle.Render("Order"))
	sb.WriteString("\n\n")

	visibleHeight := height - 3
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	scroll := 0
	if f.reorderCursor >= visibleHeight {
		scroll = f.reorderCursor - visibleHeight + 1
	}

	for i := scroll; i < len(mission.Themes) && i < scroll+visibleHeight; i++ {
		theme := strings.TrimPrefix(mission.Themes[i], "themes/")
		if len(theme) > width-4 {
			theme = theme[:width-7] + "..."
		}

		idx := fmt.Sprintf("%2d. ", i+1)
		prefix := "  "
		if i == f.reorderCursor {
			sb.WriteString(reorderHighlight.Render("▶ "+idx+theme) + "\n")
		} else {
			sb.WriteString(prefix + idx + theme + "\n")
		}
	}

	return sb.String()
}

func (m Model) renderEditPanel(width, height int) string {
	var sb strings.Builder
	f := &m.files[m.activeFile]

	mission := m.currentMission()
	if mission == nil {
		return "No mission selected"
	}

	// Input header
	label := "Edit: "
	if m.editIsNew {
		label = "Add: "
	}
	before := m.editInput[:m.editCursor]
	after := m.editInput[m.editCursor:]
	inputLine := label + before + "█" + after
	sb.WriteString(searchInputStyle.Render(inputLine))
	sb.WriteString("\n\n")

	visibleHeight := height - 3
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	scroll := 0
	if f.reorderCursor >= visibleHeight {
		scroll = f.reorderCursor - visibleHeight + 1
	}

	for i := scroll; i < len(mission.Themes) && i < scroll+visibleHeight; i++ {
		theme := strings.TrimPrefix(mission.Themes[i], "themes/")
		if len(theme) > width-4 {
			theme = theme[:width-7] + "..."
		}

		idx := fmt.Sprintf("%2d. ", i+1)
		if i == f.reorderCursor {
			sb.WriteString(reorderHighlight.Render("▶ "+idx+theme) + "\n")
		} else {
			sb.WriteString("  " + idx + theme + "\n")
		}
	}

	return sb.String()
}


// h formats a help entry: key highlighted, description dimmed.
func h(key, desc string) string {
	return helpKeyStyle.Render(key) + helpDescStyle.Render(": "+desc)
}

func helpLine(entries ...string) string {
	return strings.Join(entries, "  ")
}

func (m Model) renderStatus(totalWidth int) string {
	var left string
	if m.statusMsg != "" {
		left = reorderHighlight.Render(m.statusMsg)
	} else if m.filterActive && m.filterTyping {
		left = helpLine(h("type", "filter"), h("Enter", "apply"), h("Tab", "mode"), h("Esc", "cancel"))
	} else if m.filterActive {
		hp := activeTabStyle.Render("p") + helpDescStyle.Render(": profile")
		left = helpLine(h("j/k", "navigate"), h("t", "toggle"), h("o", "order"), h("r", "sort"), hp, h("R", "restore"), h("s", "save"), h("Esc", "clear filter"))
	} else {
		switch m.mode {
		case ModeView:
			hp := activeTabStyle.Render("p") + helpDescStyle.Render(": profile")
			left = helpLine(h("j/k", "navigate"), h("t", "toggle"), h("o", "order"), h("f", "find"), h("r", "sort"), hp, h("R", "restore"), h("s", "save"), h("q", "quit"))
		case ModeAdd:
			left = helpLine(h("j/k", "navigate"), h("Space", "toggle"), h("a", "add"), h("o", "order"), h("R", "restore"), h("s", "save"), h("Esc", "back"))
		case ModeReorder:
			left = helpLine(h("j/k", "navigate"), h("J/K", "move"), h("d", "delete"), h("e", "edit"), h("a", "add"), h("t", "toggle"), h("s", "save"), h("Esc", "back"))
		case ModeEdit:
			left = helpLine(h("type", "edit track"), h("Enter", "confirm"), h("Esc", "cancel"))
		}
	}

	// Profile tabs on the right
	right := m.renderTabBar()

	// statusBarStyle has border(2) + padding(2). Width() includes padding but not border.
	// To match panelsWidth: rendered = Width + border(2) = panelsWidth → Width = panelsWidth - 2
	styleWidth := totalWidth - 2
	// Text area inside = Width - padding(2)
	textWidth := styleWidth - 2
	if textWidth < 40 {
		textWidth = 40
	}

	// Place help left, profiles right
	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	gap := textWidth - leftLen - rightLen
	if gap < 1 {
		gap = 1
	}

	content := left + strings.Repeat(" ", gap) + right
	return statusBarStyle.Width(styleWidth).Render(content)
}

func (m Model) renderTabBar() string {
	if len(m.files) <= 1 {
		return ""
	}
	var tabs []string
	for i, f := range m.files {
		label := f.label
		if f.dirty {
			label += "*"
		}
		if i == m.activeFile {
			tabs = append(tabs, activeTabStyle.Render(label))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(label))
		}
	}
	return strings.Join(tabs, helpSepStyle.Render(" | "))
}
