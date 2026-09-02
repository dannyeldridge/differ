package main

import (
	"strings"

	"github.com/dannyeldridge/differ/internal/diffparse"
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
	lines := diffparse.Parse(diff)
	rendered := make([]string, len(lines))
	for i, l := range lines {
		switch l.Type {
		case diffparse.LineHeader:
			rendered[i] = headerStyle.Render(l.Content)
		case diffparse.LineHunk:
			rendered[i] = hunkStyle.Render(l.Content)
		case diffparse.LineAdd:
			rendered[i] = addStyle.Render("+" + l.Content)
		case diffparse.LineDel:
			rendered[i] = delStyle.Render("-" + l.Content)
		default:
			rendered[i] = contextStyle.Render(l.Content)
		}
	}
	return strings.Join(rendered, "\n")
}
