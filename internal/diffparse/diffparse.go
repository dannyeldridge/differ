package diffparse

import "strings"

// LineType classifies a single line of a unified diff.
type LineType string

const (
	LineHeader  LineType = "header"
	LineHunk    LineType = "hunk"
	LineAdd     LineType = "add"
	LineDel     LineType = "del"
	LineContext LineType = "context"
)

// Line is one classified line of a parsed diff.
type Line struct {
	Type LineType
	// Content is the line's text. For LineAdd and LineDel the leading
	// +/- marker is stripped (callers render it separately). For every
	// other type, Content is the full, unmodified raw line.
	Content string
}

// Parse splits a raw unified diff into classified lines.
func Parse(diff string) []Line {
	if diff == "" {
		return nil
	}
	raw := strings.Split(diff, "\n")
	lines := make([]Line, 0, len(raw))
	for _, l := range raw {
		switch {
		case strings.HasPrefix(l, "+++") || strings.HasPrefix(l, "---"):
			lines = append(lines, Line{Type: LineHeader, Content: l})
		case strings.HasPrefix(l, "@@"):
			lines = append(lines, Line{Type: LineHunk, Content: l})
		case strings.HasPrefix(l, "+"):
			lines = append(lines, Line{Type: LineAdd, Content: l[1:]})
		case strings.HasPrefix(l, "-"):
			lines = append(lines, Line{Type: LineDel, Content: l[1:]})
		default:
			lines = append(lines, Line{Type: LineContext, Content: l})
		}
	}
	return lines
}
