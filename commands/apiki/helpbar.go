package apiki

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewHelpBar() string {
	keyStyle := lipgloss.NewStyle().
		Background(ColorCyan).
		Foreground(ColorBlack).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Foreground(ColorWhite)

	var items []string

	switch m.mode {
	case modeList, modeImport:
		if m.filtering {
			items = []string{
				keyStyle.Render("Enter") + labelStyle.Render("Apply"),
				keyStyle.Render("Esc") + labelStyle.Render("Cancel"),
			}
		} else {
			// Check if current entry is from .env file
			isDotEnvEntry := false
			if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
				actualIndex := m.filteredIndices[m.cursor]
				if actualIndex < len(m.entries) {
					isDotEnvEntry = m.entries[actualIndex].SourceFile != ""
				}
			}

			baseItems := []string{
				keyStyle.Render("/") + labelStyle.Render("Filter"),
			}
			if m.filterInput.Value() != "" {
				baseItems = append(baseItems,
					keyStyle.Render("Esc")+labelStyle.Render("Clear"),
				)
			}
			baseItems = append(baseItems,
				keyStyle.Render("↑↓")+labelStyle.Render("Move"),
				keyStyle.Render("Space")+labelStyle.Render("Toggle"),
			)

			// Only show edit/delete/create options in list mode, not import
			// mode
			if m.mode == modeList {
				baseItems = append(baseItems,
					keyStyle.Render("+")+labelStyle.Render("Create"),
				)

				if isDotEnvEntry {
					baseItems = append(baseItems,
						keyStyle.Render("=")+labelStyle.Render("Add&Edit"),
					)
				} else {
					baseItems = append(baseItems,
						keyStyle.Render("=")+labelStyle.Render("Edit"),
						keyStyle.Render("-")+labelStyle.Render("Delete"),
					)
				}
			}

			if m.mode == modeImport {
				baseItems = append(baseItems,
					keyStyle.Render("Enter")+labelStyle.Render("Import"),
					keyStyle.Render("Esc")+labelStyle.Render("Cancel"),
				)
			} else {
				baseItems = append(baseItems,
					keyStyle.Render("i")+labelStyle.Render("Import"),
					keyStyle.Render("Enter")+labelStyle.Render("Apply"),
					keyStyle.Render("q")+labelStyle.Render("Cancel"),
				)
			}

			items = baseItems
		}
	case modeAdd, modeEdit:
		items = []string{
			keyStyle.Render("Tab/↓") + labelStyle.Render("Next"),
			keyStyle.Render("Shift+Tab/↑") + labelStyle.Render("Prev"),
			keyStyle.Render("Enter") + labelStyle.Render("Save"),
			keyStyle.Render("Esc") + labelStyle.Render("Cancel"),
		}
	case modeConfirmDelete:
		items = []string{
			keyStyle.Render("y/Enter") + labelStyle.Render("Yes"),
			keyStyle.Render("n/Esc") + labelStyle.Render("No"),
		}
	case modeConfirmPromote:
		items = []string{
			keyStyle.Render("y/Enter") + labelStyle.Render("Yes"),
			keyStyle.Render("n/Esc") + labelStyle.Render("No"),
		}
	case modeConfirmImport:
		items = []string{
			keyStyle.Render("y/Enter") + labelStyle.Render("Yes"),
			keyStyle.Render("n/Esc") + labelStyle.Render("No"),
		}
	case modeError:
		items = []string{
			keyStyle.Render("Enter") + labelStyle.Render("Continue"),
			keyStyle.Render("q") + labelStyle.Render("Quit"),
		}
	}

	return strings.Join(items, "")
}
