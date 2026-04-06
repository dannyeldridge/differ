package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	addStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))             // green
	delStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))             // red
	hunkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))             // cyan
	headerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true) // yellow bold
	contextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))           // dim
)

// ColorizeDiff applies terminal colors to a unified diff string.
// Lines starting with + are green, - are red, @@ are cyan, file headers are bold yellow.
func ColorizeDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	var sb strings.Builder
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			sb.WriteString(headerStyle.Render(line))
		case strings.HasPrefix(line, "+"):
			sb.WriteString(addStyle.Render(line))
		case strings.HasPrefix(line, "-"):
			sb.WriteString(delStyle.Render(line))
		case strings.HasPrefix(line, "@@"):
			sb.WriteString(hunkStyle.Render(line))
		default:
			sb.WriteString(contextStyle.Render(line))
		}
		if i < len(lines)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
