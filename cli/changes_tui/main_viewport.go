package changes_tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m changesUIModel) renderMainView() string {
	mainViewHeader := m.renderMainViewHeader()

	if m.selectedFullFile() {
		fileView := m.fileViewport.View()

		fileViews := []string{fileView}

		if m.fileScrollable() {
			fileViews = append(fileViews, m.renderScrollFooter())
		}

		fileContainer := lipgloss.JoinVertical(lipgloss.Left, fileViews...)

		fileContainerStyle := lipgloss.NewStyle().Width(m.fileViewport.Width)
		fileContainer = fileContainerStyle.Render(fileContainer)

		return lipgloss.JoinVertical(lipgloss.Left, mainViewHeader, fileContainer)
	} else {
		oldView := m.changeOldViewport.View()
		newView := m.changeNewViewport.View()

		oldViews := []string{oldView}
		newViews := []string{newView}

		if m.oldScrollable() && m.selectedViewport == 0 {
			oldViews = append(oldViews, m.renderScrollFooter())
		} else if m.newScrollable() {
			newViews = append(newViews, m.renderScrollFooter())
		}

		oldContainer := lipgloss.JoinVertical(lipgloss.Left, oldViews...)
		newContainer := lipgloss.JoinVertical(lipgloss.Left, newViews...)

		oldContainerStyle := lipgloss.NewStyle().Width(m.changeOldViewport.Width)
		oldContainer = oldContainerStyle.Render(oldContainer)

		newContainerStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(lipgloss.Color(borderColor)).
			Width(m.changeNewViewport.Width)

		newContainer = newContainerStyle.Render(newContainer)

		return lipgloss.JoinVertical(lipgloss.Left,
			mainViewHeader,
			lipgloss.JoinHorizontal(lipgloss.Top, oldContainer, newContainer),
			m.renderMainViewFooter(),
		)
	}

}

func (m changesUIModel) renderMainViewHeader() string {
	if m.selectionInfo == nil {
		return "\n"
	}

	sidebarWidth := lipgloss.Width(m.renderSidebar())
	style := lipgloss.NewStyle().
		Width(m.width - sidebarWidth).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(borderColor)

	var header string
	if m.selectedFullFile() {
		numChanges := m.selectionInfo.currentPlanBeforeReplacement.NumPendingForPath(m.selectionInfo.currentPath)

		if numChanges > 0 {
			suffix := "s"
			if numChanges == 1 {
				suffix = ""
			}
			header = fmt.Sprintf(" ✅ Final state of %s including %d change%s", m.selectionInfo.currentPath, numChanges, suffix)
		} else {
			header = fmt.Sprintf(" ✅ New file: %s", m.selectionInfo.currentPath)
		}
	} else {
		header = " 👉 " + m.selectionInfo.currentRep.Summary
	}

	return style.Render(header)
}

func (m changesUIModel) renderMainViewFooter() string {
	if m.selectedFullFile() {
		return ""
	}

	sidebarWidth := lipgloss.Width(m.renderSidebar())
	style := lipgloss.NewStyle().Width(m.width - sidebarWidth).Inherit(topBorderStyle).Foreground(lipgloss.Color(helpTextColor))
	footer := ` (d)rop selected change • (c)opy to clipboard`
	return style.Render(footer)
}

func (m changesUIModel) renderScrollFooter() string {
	if m.selectionInfo == nil {
		return ""
	}

	width, _ := m.getMainViewDims()

	if m.selectionInfo.currentRes != nil {
		width = width / 2
	}

	var footer string

	if m.selectedFullFile() {
		footer = `(j/k) scroll down/up • (J/K) page down/up`
	} else {
		footer = ` (j/k) scroll`
		if m.oldScrollable() && m.newScrollable() {
			footer += ` • (tab) switch view`
		} else {
			footer += ` down/up`
		}
	}

	style := lipgloss.NewStyle().Width(width).Inherit(topBorderStyle).Foreground(lipgloss.Color(helpTextColor))

	return style.Render(footer)
}
