// Package tablefmt formats GitHub-flavored Markdown pipe tables.
package tablefmt

import (
	"errors"
	"strings"

	"github.com/mattn/go-runewidth"
)

// Format formats the first pipe table in lines. Lines before and after the
// table, including surrounding blank lines, are returned unchanged.
func Format(lines []string) ([]string, error) {
	header, delimiter, ok := findTable(lines)
	if !ok {
		return nil, errors.New("tablefmt: no valid table (expected a header followed by a delimiter row)")
	}

	headerCells := splitRow(lines[header])
	delimiters := splitRow(lines[delimiter])
	alignments := make([]alignment, len(delimiters))
	for i, cell := range delimiters {
		alignments[i], _ = parseDelimiter(cell)
	}

	end := delimiter + 1
	for end < len(lines) && !isBlank(lines[end]) {
		end++
	}

	rows := make([][]string, 0, end-delimiter)
	rows = append(rows, headerCells)
	for i := delimiter + 1; i < end; i++ {
		rows = append(rows, splitRow(lines[i]))
	}

	columnCount := len(alignments)
	for _, row := range rows {
		if len(row) > columnCount {
			columnCount = len(row)
		}
	}
	for len(alignments) < columnCount {
		alignments = append(alignments, alignNone)
	}

	widths := make([]int, columnCount)
	for i, alignment := range alignments {
		widths[i] = 3 + alignment.colonCount()
	}
	for _, row := range rows {
		for i, cell := range row {
			if width := runewidth.StringWidth(cell); width > widths[i] {
				widths[i] = width
			}
		}
	}

	formatted := make([]string, 0, len(lines))
	formatted = append(formatted, lines[:header]...)
	formatted = append(formatted, formatRow(headerCells, widths))
	formatted = append(formatted, formatDelimiterRow(alignments, widths))
	for _, row := range rows[1:] {
		formatted = append(formatted, formatRow(row, widths))
	}
	formatted = append(formatted, lines[end:]...)
	return formatted, nil
}

type alignment uint8

const (
	alignNone alignment = iota
	alignLeft
	alignRight
	alignCenter
)

func (a alignment) colonCount() int {
	switch a {
	case alignLeft, alignRight:
		return 1
	case alignCenter:
		return 2
	default:
		return 0
	}
}

func findTable(lines []string) (header, delimiter int, ok bool) {
	for i := 0; i+1 < len(lines); i++ {
		if isBlank(lines[i]) || isBlank(lines[i+1]) {
			continue
		}
		// A pipe is required to distinguish a one-column table from a Setext heading.
		if !strings.Contains(lines[i], "|") && !strings.Contains(lines[i+1], "|") {
			continue
		}
		headerCells := splitRow(lines[i])
		delimiterCells := splitRow(lines[i+1])
		if len(headerCells) == 0 || len(delimiterCells) == 0 {
			continue
		}
		valid := true
		for _, cell := range delimiterCells {
			if _, ok := parseDelimiter(cell); !ok {
				valid = false
				break
			}
		}
		if valid {
			return i, i + 1, true
		}
	}
	return 0, 0, false
}

func parseDelimiter(cell string) (alignment, bool) {
	cell = strings.TrimSpace(cell)
	if cell == "" {
		return alignNone, false
	}

	left := cell[0] == ':'
	if left {
		cell = cell[1:]
	}
	right := strings.HasSuffix(cell, ":")
	if right {
		cell = cell[:len(cell)-1]
	}
	if len(cell) < 3 || strings.Trim(cell, "-") != "" {
		return alignNone, false
	}

	switch {
	case left && right:
		return alignCenter, true
	case left:
		return alignLeft, true
	case right:
		return alignRight, true
	default:
		return alignNone, true
	}
}

func formatRow(cells []string, widths []int) string {
	var b strings.Builder
	b.WriteString("|")
	for i, width := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		b.WriteByte(' ')
		b.WriteString(cell)
		b.WriteString(strings.Repeat(" ", width-runewidth.StringWidth(cell)))
		b.WriteString(" |")
	}
	return b.String()
}

func formatDelimiterRow(alignments []alignment, widths []int) string {
	cells := make([]string, len(widths))
	for i, alignment := range alignments {
		switch alignment {
		case alignLeft:
			cells[i] = ":" + strings.Repeat("-", widths[i]-1)
		case alignRight:
			cells[i] = strings.Repeat("-", widths[i]-1) + ":"
		case alignCenter:
			cells[i] = ":" + strings.Repeat("-", widths[i]-2) + ":"
		default:
			cells[i] = strings.Repeat("-", widths[i])
		}
	}
	return "| " + strings.Join(cells, " | ") + " |"
}

func splitRow(line string) []string {
	line = strings.TrimSpace(line)
	var cells []string
	start := 0
	codeTicks := 0

	for i := 0; i < len(line); {
		switch line[i] {
		case '\\':
			if codeTicks == 0 && i+1 < len(line) {
				i += 2
			} else {
				i++
			}
		case '`':
			runEnd := i + 1
			for runEnd < len(line) && line[runEnd] == '`' {
				runEnd++
			}
			runLength := runEnd - i
			if codeTicks == 0 {
				if hasClosingBackticks(line, runEnd, runLength) {
					codeTicks = runLength
				}
			} else if runLength == codeTicks {
				codeTicks = 0
			}
			i = runEnd
		case '|':
			if codeTicks == 0 {
				cells = append(cells, strings.TrimSpace(line[start:i]))
				start = i + 1
			}
			i++
		default:
			i++
		}
	}
	cells = append(cells, strings.TrimSpace(line[start:]))

	if len(cells) > 1 && cells[0] == "" && strings.HasPrefix(line, "|") {
		cells = cells[1:]
	}
	if len(cells) > 1 && cells[len(cells)-1] == "" && strings.HasSuffix(line, "|") {
		cells = cells[:len(cells)-1]
	}
	return cells
}

func hasClosingBackticks(s string, start, wanted int) bool {
	for i := start; i < len(s); {
		if s[i] != '`' {
			i++
			continue
		}
		end := i + 1
		for end < len(s) && s[end] == '`' {
			end++
		}
		if end-i == wanted {
			return true
		}
		i = end
	}
	return false
}

func isBlank(line string) bool {
	return strings.TrimSpace(line) == ""
}
