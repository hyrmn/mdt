// Package viewer renders Markdown for ANSI-compatible terminals.
package viewer

import (
	"fmt"
	"strings"

	"charm.land/glamour/v2"
	"github.com/charmbracelet/x/ansi"
)

const defaultWidth = 80

// Images stay readable in every terminal: no graphics protocol is required.
// Explicit heading weights also prevent lower-level styles from dropping bold.
const styleOverrides = `{
  "heading": {"bold": true},
  "h1": {"bold": true},
  "h2": {"bold": true},
  "h3": {"bold": true},
  "h4": {"bold": true},
  "h5": {"bold": true},
  "h6": {"bold": true},
  "image_text": {"format": "[Image: {{.text}}]"},
  "image": {"prefix": "(", "suffix": ")"}
}`

// Render converts Markdown to styled terminal text. style may be a Glamour
// built-in style name or a JSON style path; an empty style uses GLAMOUR_STYLE.
func Render(source []byte, width int, style string) (string, error) {
	if width <= 0 {
		width = defaultWidth
	}
	styleOption := glamour.WithEnvironmentConfig()
	if style != "" {
		styleOption = glamour.WithStylePath(style)
	}
	renderer, err := glamour.NewTermRenderer(
		styleOption,
		glamour.WithStylesFromJSONBytes([]byte(styleOverrides)),
		glamour.WithWordWrap(width),
		glamour.WithTableWrap(true),
		glamour.WithInlineTableLinks(true),
		glamour.WithEmoji(),
	)
	if err != nil {
		return "", fmt.Errorf("configure renderer: %w", err)
	}
	output, err := renderer.Render(string(source))
	if err != nil {
		return "", fmt.Errorf("render Markdown: %w", err)
	}
	if style == "notty" {
		output = cleanPlainOutput(output)
	}
	return output, nil
}

func cleanPlainOutput(output string) string {
	lines := strings.Split(ansi.Strip(output), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}
