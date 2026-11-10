package editor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func newTestEditor(t *testing.T, text string) *Editor {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("initialize simulation screen: %v", err)
	}
	screen.SetSize(100, 30)
	t.Cleanup(screen.Fini)
	return &Editor{screen: screen, path: "test.md", buffer: NewBuffer(text)}
}

func TestFormatTableAtCursor(t *testing.T) {
	editor := newTestEditor(t, "before\nName|Value\n---|---:\nA|42\nafter")
	editor.buffer.SetCursor(Position{Line: 2})
	editor.formatTable()

	want := "before\n| Name | Value |\n| ---- | ----: |\n| A    | 42    |\nafter"
	if got := editor.buffer.Text(); got != want {
		t.Fatalf("formatted buffer =\n%s\nwant:\n%s", got, want)
	}
	if !editor.buffer.Dirty() {
		t.Fatal("formatting did not mark the buffer dirty")
	}
}

func TestFormatSelectedTable(t *testing.T) {
	editor := newTestEditor(t, "before\n| a | b |\n| --- | --- |\n| longer | c |\nafter")
	editor.buffer.SetCursor(Position{Line: 1})
	editor.buffer.ToggleMark()
	editor.buffer.SetCursor(Position{Line: 3})
	editor.handleKey(tcell.NewEventKey(tcell.KeyF6, 0, tcell.ModNone))

	want := "before\n| a      | b   |\n| ------ | --- |\n| longer | c   |\nafter"
	if got := editor.buffer.Text(); got != want {
		t.Fatalf("formatted buffer =\n%s\nwant:\n%s", got, want)
	}
	if _, ok := editor.buffer.Mark(); ok {
		t.Fatal("selection mark was not cleared")
	}
}

func TestFormatRejectsNonTable(t *testing.T) {
	editor := newTestEditor(t, "plain text")
	editor.formatTable()
	if got := editor.buffer.Text(); got != "plain text" {
		t.Fatalf("buffer changed to %q", got)
	}
	if editor.buffer.Dirty() {
		t.Fatal("failed formatting marked the buffer dirty")
	}
}

func TestSaveAndUnsavedExit(t *testing.T) {
	editor := newTestEditor(t, "before")
	editor.path = filepath.Join(t.TempDir(), "note.md")
	editor.buffer.InsertRune('!')

	editor.handleKey(tcell.NewEventKey(tcell.KeyCtrlX, 0, tcell.ModNone))
	if !editor.confirmQuit || editor.quit {
		t.Fatal("unsaved exit did not request confirmation")
	}
	editor.handleKey(tcell.NewEventKey(tcell.KeyCtrlS, 0, tcell.ModNone))
	if !editor.quit || editor.confirmQuit {
		t.Fatal("save-and-exit did not exit cleanly")
	}
	contents, err := os.ReadFile(editor.path)
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if got, want := string(contents), "!before"; got != want {
		t.Fatalf("saved contents = %q, want %q", got, want)
	}
}

func TestDrawShowsBoldHintAndKeepsMarkers(t *testing.T) {
	editor := newTestEditor(t, "**bold**")
	editor.draw()

	for x, want := range []rune("**bold**") {
		cell, _, _, _ := editor.screen.GetContent(x, 1)
		if cell != want {
			t.Errorf("cell %d = %q, want %q", x, cell, want)
		}
	}
	_, _, attributes := mustCellStyle(t, editor.screen, 2, 1)
	if attributes&tcell.AttrBold == 0 {
		t.Fatal("text inside ** markers is not bold")
	}
}

func mustCellStyle(t *testing.T, screen tcell.Screen, x, y int) (tcell.Color, tcell.Color, tcell.AttrMask) {
	t.Helper()
	_, style, _ := screen.Get(x, y)
	foreground, background, attributes := style.Decompose()
	return foreground, background, attributes
}
