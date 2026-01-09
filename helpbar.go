package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewHelpBar() string {
	keyStyle := lipgloss.NewStyle().
		Background(colorCyan).
		Foreground(colorBlack).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Foreground(colorWhite).
		MarginRight(2)

	var items []string

	switch m.mode {
	case modeList:
		if m.filtering {
			items = []string{
				keyStyle.Render("Enter") + labelStyle.Render("Apply"),
				keyStyle.Render("Esc") + labelStyle.Render("Cancel"),
			}
		} else {
			items = []string{
				keyStyle.Render("/") + labelStyle.Render("Filter"),
				keyStyle.Render("↑↓") + labelStyle.Render("Move"),
				keyStyle.Render("Space") + labelStyle.Render("Toggle"),
				keyStyle.Render("+") + labelStyle.Render("Add"),
				keyStyle.Render("Enter") + labelStyle.Render("Edit"),
				keyStyle.Render("-") + labelStyle.Render("Delete"),
				keyStyle.Render("q") + labelStyle.Render("Quit"),
			}
			if m.filterInput.Value() != "" {
				// Insert "Esc Clear" after "/" when filter is active
				items = []string{
					keyStyle.Render("/") + labelStyle.Render("Filter"),
					keyStyle.Render("Esc") + labelStyle.Render("Clear"),
					keyStyle.Render("↑↓") + labelStyle.Render("Move"),
					keyStyle.Render("Space") + labelStyle.Render("Toggle"),
					keyStyle.Render("+") + labelStyle.Render("Add"),
					keyStyle.Render("Enter") + labelStyle.Render("Edit"),
					keyStyle.Render("-") + labelStyle.Render("Delete"),
					keyStyle.Render("q") + labelStyle.Render("Quit"),
				}
			}
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
	}

	return strings.Join(items, "")
}
