package main

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Mission represents one row from the themes XML.
type Mission struct {
	Name            string
	Themes          []string // e.g. ["themes/USA_01.ogg", "themes/02.ogg"]
	OriginalThemes  []string // snapshot from file load for per-mission restore
	Comment         string   // optional 3rd column comment
	ThemesDataStart int      // byte offset of themes <Data> content start
	ThemesDataEnd   int      // byte offset of themes <Data> content end
	InsertPoint     int      // byte offset for inserting new themes cell (after first </Cell>)
	OrigOrder       int      // original parse order index for unsorted view
	HasThemesCell   bool
}

var USATracks = []string{
	"USA_01.ogg",
	"USA_02-19.ogg",
	"USA_03-02.ogg",
	"USA_04-07.ogg",
	"USA_05-12.ogg",
	"USA_06-08.ogg",
	"USA_07.ogg",
	"USA_08.ogg",
	"USA_09.ogg",
	"USA_11-01.ogg",
}

var MuteTracks = []string{
	"mute_10.ogg",
	"mute_20.ogg",
	"mute_30.ogg",
}

// AllTracks returns USA + Mute tracks combined for the checkbox panel.
func AllTracks() []string {
	return append(append([]string{}, USATracks...), MuteTracks...)
}

// ThemesString builds the comma-separated themes value for XML.
func ThemesString(themes []string) string {
	return strings.Join(themes, ",")
}

// ParseThemes splits a comma-separated themes string into a slice.
func ParseThemes(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// HasUSATrack checks if a mission has a specific USA track.
func HasUSATrack(themes []string, track string) bool {
	full := "themes/" + track
	for _, t := range themes {
		if t == full {
			return true
		}
	}
	return false
}

// ParseThemesXML parses the Excel XML Spreadsheet and extracts missions with byte offsets.
func ParseThemesXML(data []byte) ([]Mission, error) {
	var missions []Mission

	// Find all <Row...> blocks in the data (tag may have attributes like ss:AutoFitHeight)
	rowTag := []byte("<Row")
	rowEnd := []byte("</Row>")
	dataStartTag := []byte(`<Data ss:Type="String">`)
	dataEndTag := []byte("</Data>")
	cellEndTag := []byte("</Cell>")

	pos := 0
	isFirstRow := true

	for {
		rowStart := bytes.Index(data[pos:], rowTag)
		if rowStart == -1 {
			break
		}
		rowStart += pos

		rowEndIdx := bytes.Index(data[rowStart:], rowEnd)
		if rowEndIdx == -1 {
			break
		}
		rowEndIdx += rowStart + len(rowEnd)

		rowData := data[rowStart:rowEndIdx]

		// Skip header row
		if isFirstRow {
			isFirstRow = false
			pos = rowEndIdx
			continue
		}

		// Find first cell's Data content (context_name)
		d1Start := bytes.Index(rowData, dataStartTag)
		if d1Start == -1 {
			pos = rowEndIdx
			continue
		}
		d1ContentStart := d1Start + len(dataStartTag)
		d1End := bytes.Index(rowData[d1ContentStart:], dataEndTag)
		if d1End == -1 {
			pos = rowEndIdx
			continue
		}
		contextName := string(rowData[d1ContentStart : d1ContentStart+d1End])

		// Find the end of the first </Cell> for insert point
		firstCellEnd := bytes.Index(rowData, cellEndTag)
		insertPoint := -1
		if firstCellEnd != -1 {
			insertPoint = rowStart + firstCellEnd + len(cellEndTag)
		}

		// Find second cell's Data content (themes) - search after first </Cell>
		m := Mission{
			Name:          contextName,
			InsertPoint:   insertPoint,
			HasThemesCell: false,
		}

		if firstCellEnd != -1 {
			afterFirstCell := rowData[firstCellEnd+len(cellEndTag):]
			d2Start := bytes.Index(afterFirstCell, dataStartTag)
			if d2Start != -1 {
				// Calculate absolute offsets
				d2ContentStartRel := d2Start + len(dataStartTag)
				absContentStart := rowStart + (firstCellEnd + len(cellEndTag)) + d2ContentStartRel
				d2End := bytes.Index(afterFirstCell[d2ContentStartRel:], dataEndTag)
				if d2End != -1 {
					absContentEnd := absContentStart + d2End
					themesStr := string(data[absContentStart:absContentEnd])
					m.Themes = ParseThemes(themesStr)
					m.OriginalThemes = make([]string, len(m.Themes))
					copy(m.OriginalThemes, m.Themes)
					m.ThemesDataStart = absContentStart
					m.ThemesDataEnd = absContentEnd
					m.HasThemesCell = true
				}
			} else {
				// Check if there's an empty second cell (Cell with StyleID but no Data)
				secondCellStart := bytes.Index(afterFirstCell, []byte("<Cell"))
				if secondCellStart != -1 {
					// There's a cell but with no Data — find its position for replacement
					secondCellData := afterFirstCell[secondCellStart:]
					secondCellEndIdx := bytes.Index(secondCellData, cellEndTag)
					if secondCellEndIdx != -1 {
						// We'll treat this as having no themes cell for simplicity
						// and use insert point
						m.HasThemesCell = false
					}
				}
			}

			// Try to find comment (3rd cell)
			if m.HasThemesCell {
				secondCellEnd := bytes.Index(afterFirstCell[d2Start:], cellEndTag)
				if secondCellEnd != -1 {
					afterSecondCell := afterFirstCell[d2Start+secondCellEnd+len(cellEndTag):]
					d3Start := bytes.Index(afterSecondCell, dataStartTag)
					if d3Start != -1 {
						d3ContentStart := d3Start + len(dataStartTag)
						d3End := bytes.Index(afterSecondCell[d3ContentStart:], dataEndTag)
						if d3End != -1 {
							m.Comment = string(afterSecondCell[d3ContentStart : d3ContentStart+d3End])
						}
					}
				}
			}
		}

		missions = append(missions, m)
		pos = rowEndIdx
	}

	if len(missions) == 0 {
		return nil, fmt.Errorf("no missions found in XML")
	}
	return missions, nil
}

var missionNumRe = regexp.MustCompile(`^mission_(\d+)`)

// missionSortKey returns (group, number, name) for sorting.
// Group 0: special (global_map, main_menu, credits, etc.)
// Group 1: numbered missions (mission_1st, mission_2nd, etc.)
// Group 2: named missions without number (mission_electric, mission_action, etc.)
// Group 3: skirmish/rnd/multiplayer
func missionSortKey(name string) (int, int, string) {
	if strings.Contains(name, "multiplayer") {
		return 3, 0, name
	}
	if !strings.HasPrefix(name, "mission_") {
		return 0, 0, name
	}

	// Check for skirmish, rnd_enc
	rest := name[len("mission_"):]
	if strings.HasPrefix(rest, "skirmish") || strings.HasPrefix(rest, "rnd_") {
		return 3, 0, name
	}

	// Extract leading number: mission_1st, mission_2nd, mission_10th, mission_14_middle
	match := missionNumRe.FindStringSubmatch(name)
	if match != nil {
		num, _ := strconv.Atoi(match[1])
		// Suffix after the number part for sub-sorting.
		// Strip "mission_N" prefix and optional ordinal suffix (st/nd/rd/th)
		suffix := name[len(match[0]):]
		suffix = strings.TrimLeft(suffix, "stndrh")
		if suffix == "_mission" || suffix == "" {
			suffix = "" // parent mission sorts first
		}
		return 1, num, suffix
	}

	// Named missions without number (mission_electric, mission_action, mission_surviviist, etc.)
	return 2, 0, name
}

func sortMissions(missions []Mission) {
	sort.SliceStable(missions, func(i, j int) bool {
		gi, ni, si := missionSortKey(missions[i].Name)
		gj, nj, sj := missionSortKey(missions[j].Name)
		if gi != gj {
			return gi < gj
		}
		if ni != nj {
			return ni < nj
		}
		return si < sj
	})
}

func sortByOrigOrder(missions []Mission) {
	sort.SliceStable(missions, func(i, j int) bool {
		return missions[i].OrigOrder < missions[j].OrigOrder
	})
}

// edit represents a byte-level edit operation on the XML.
type edit struct {
	start   int
	end     int
	content []byte
	isInsert bool
}

// SerializeThemesXML applies theme changes to the original XML bytes.
func SerializeThemesXML(original []byte, missions []Mission) []byte {
	var edits []edit

	for _, m := range missions {
		newThemes := ThemesString(m.Themes)

		if m.HasThemesCell {
			// Replace the content between ThemesDataStart and ThemesDataEnd
			edits = append(edits, edit{
				start:   m.ThemesDataStart,
				end:     m.ThemesDataEnd,
				content: []byte(newThemes),
			})
		} else if len(m.Themes) > 0 && m.InsertPoint > 0 {
			// Insert a new cell after the first cell
			newCell := fmt.Sprintf(`<Cell><Data ss:Type="String">%s</Data></Cell>`, newThemes)
			edits = append(edits, edit{
				start:    m.InsertPoint,
				end:      m.InsertPoint,
				content:  []byte(newCell),
				isInsert: true,
			})
		}
	}

	// Sort edits by start offset descending (apply from end to start to preserve offsets)
	for i := 0; i < len(edits); i++ {
		for j := i + 1; j < len(edits); j++ {
			if edits[j].start > edits[i].start {
				edits[i], edits[j] = edits[j], edits[i]
			}
		}
	}

	result := make([]byte, len(original))
	copy(result, original)

	for _, e := range edits {
		var newResult []byte
		newResult = append(newResult, result[:e.start]...)
		newResult = append(newResult, e.content...)
		newResult = append(newResult, result[e.end:]...)
		result = newResult
	}

	return result
}
