package viewer

import (
	"strings"
	"testing"

	"charm.land/glamour/v2/styles"
	"github.com/charmbracelet/x/ansi"
)

const sample = `# Heading

This is **bold**, _italic_, and ` + "`code`" + `.

| Name | Value |
| --- | ---: |
| A | 42 |

![Architecture diagram](images/diagram.png)
`

func TestRenderMarkdownFeatures(t *testing.T) {
	output, err := Render([]byte(sample), 100, styles.DarkStyle)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	visible := ansi.Strip(output)
	for _, want := range []string{
		"Heading",
		"bold",
		"italic",
		"code",
		"Name",
		"Value",
		"42",
		"[Image: Architecture diagram]",
		"images/diagram.png",
	} {
		if !strings.Contains(visible, want) {
			t.Errorf("output does not contain %q:\n%s", want, visible)
		}
	}
	if !strings.Contains(output, "\x1b[") {
		t.Errorf("dark output has no ANSI styling: %q", output)
	}
}

func TestRenderNoTTY(t *testing.T) {
	output, err := Render([]byte(sample), 100, styles.NoTTYStyle)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if strings.Contains(output, "\x1b") {
		t.Errorf("notty output contains escape sequences: %q", output)
	}
	if !strings.Contains(output, "Name") || !strings.Contains(output, "A") || !strings.Contains(output, "|") {
		t.Errorf("notty output did not preserve a readable table:\n%s", output)
	}
	if !strings.Contains(output, "images/diagram.png") {
		t.Errorf("notty output omitted the image destination:\n%s", output)
	}
}

func TestRenderDefaultsInvalidWidth(t *testing.T) {
	if _, err := Render([]byte("hello"), 0, styles.NoTTYStyle); err != nil {
		t.Fatalf("Render() with default width error = %v", err)
	}
}

func TestRenderRejectsUnknownStyle(t *testing.T) {
	if _, err := Render([]byte("hello"), 80, "style-that-does-not-exist"); err == nil {
		t.Fatal("Render() with unknown style returned no error")
	}
}
