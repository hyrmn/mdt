package editor

import (
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var (
	headingPattern = regexp.MustCompile(`^ {0,3}#{1,6}(?: |$)`)
	listPattern    = regexp.MustCompile(`^(\s*)([-+*]|\d+[.)])(\s+)`)
)

// inlineStyles adds WYSIWYG hints while leaving every Markdown marker visible.
func inlineStyles(line string) []tcell.Style {
	base := tcell.StyleDefault
	styles := make([]tcell.Style, len([]rune(line)))
	for i := range styles {
		styles[i] = base
	}

	if headingPattern.MatchString(line) {
		for i := range styles {
			styles[i] = styles[i].Bold(true).Foreground(tcell.ColorAqua)
		}
		return styles
	}
	if strings.HasPrefix(strings.TrimLeft(line, " "), ">") {
		for i := range styles {
			styles[i] = styles[i].Foreground(tcell.ColorGreen).Italic(true)
		}
	}
	if match := listPattern.FindStringSubmatchIndex(line); match != nil {
		for i := match[4]; i < match[5]; i++ {
			styles[i] = styles[i].Bold(true).Foreground(tcell.ColorYellow)
		}
	}
	if strings.HasPrefix(strings.TrimSpace(line), "```") || strings.HasPrefix(strings.TrimSpace(line), "~~~") {
		for i := range styles {
			styles[i] = styles[i].Foreground(tcell.ColorYellow)
		}
		return styles
	}

	runes := []rune(line)
	for i := 0; i < len(runes); {
		if isEscaped(runes, i) {
			i++
			continue
		}

		if runes[i] == '`' {
			runLength := delimiterLength(runes, i, '`')
			if end := findDelimiter(runes, i+runLength, "`", runLength); end >= 0 {
				style := tcell.StyleDefault.Foreground(tcell.ColorYellow)
				applyStyle(styles, i, end+runLength, style)
				i = end + runLength
				continue
			}
		}

		matched := false
		for _, delimiter := range []struct {
			token string
			attr  tcell.AttrMask
		}{
			{token: "**", attr: tcell.AttrBold},
			{token: "__", attr: tcell.AttrBold},
			{token: "~~", attr: tcell.AttrStrikeThrough},
			{token: "*", attr: tcell.AttrItalic},
			{token: "_", attr: tcell.AttrItalic},
		} {
			token := []rune(delimiter.token)
			if !hasRunes(runes, i, token) {
				continue
			}
			end := findToken(runes, i+len(token), token)
			if end < 0 || end == i+len(token) {
				continue
			}
			for column := i; column < end+len(token); column++ {
				styles[column] = styles[column].Attributes(delimiter.attr)
			}
			for column := i; column < i+len(token); column++ {
				styles[column] = styles[column].Dim(true)
			}
			for column := end; column < end+len(token); column++ {
				styles[column] = styles[column].Dim(true)
			}
			i = end + len(token)
			matched = true
			break
		}
		if matched {
			continue
		}

		if runes[i] == '[' {
			if closeText := findByteRune(runes, i+1, ']'); closeText >= 0 && closeText+1 < len(runes) && runes[closeText+1] == '(' {
				if closeTarget := findByteRune(runes, closeText+2, ')'); closeTarget >= 0 {
					for column := i + 1; column < closeText; column++ {
						styles[column] = styles[column].Foreground(tcell.ColorAqua).Underline(true)
					}
					for column := closeText + 1; column <= closeTarget; column++ {
						styles[column] = styles[column].Dim(true)
					}
					i = closeTarget + 1
					continue
				}
			}
		}
		i++
	}
	return styles
}

func applyStyle(styles []tcell.Style, start, end int, style tcell.Style) {
	for i := start; i < end && i < len(styles); i++ {
		styles[i] = style
	}
}

func delimiterLength(runes []rune, start int, delimiter rune) int {
	end := start
	for end < len(runes) && runes[end] == delimiter {
		end++
	}
	return end - start
}

func findDelimiter(runes []rune, start int, delimiter string, length int) int {
	token := []rune(delimiter)
	for i := start; i+length <= len(runes); i++ {
		if isEscaped(runes, i) {
			continue
		}
		matched := true
		for offset := 0; offset < length; offset++ {
			if runes[i+offset] != token[0] {
				matched = false
				break
			}
		}
		if matched {
			return i
		}
	}
	return -1
}

func findToken(runes []rune, start int, token []rune) int {
	for i := start; i+len(token) <= len(runes); i++ {
		if isEscaped(runes, i) {
			continue
		}
		if hasRunes(runes, i, token) {
			return i
		}
	}
	return -1
}

func hasRunes(runes []rune, start int, expected []rune) bool {
	if start < 0 || start+len(expected) > len(runes) {
		return false
	}
	for i, r := range expected {
		if runes[start+i] != r {
			return false
		}
	}
	return true
}

func isEscaped(runes []rune, position int) bool {
	backslashes := 0
	for position--; position >= 0 && runes[position] == '\\'; position-- {
		backslashes++
	}
	return backslashes%2 == 1
}

func findByteRune(runes []rune, start int, target rune) int {
	for i := start; i < len(runes); i++ {
		if runes[i] == target && !isEscaped(runes, i) {
			return i
		}
	}
	return -1
}
