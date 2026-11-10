package editor

import (
	"reflect"
	"testing"
)

func TestBufferEditingUTF8(t *testing.T) {
	buffer := NewBuffer("hello 世界")
	buffer.MoveDocumentEnd()
	buffer.InsertRune('!')
	buffer.MoveHorizontal(-1)
	buffer.Backspace()

	if got, want := buffer.Text(), "hello 世!"; got != want {
		t.Fatalf("Text() = %q, want %q", got, want)
	}
	if got, want := buffer.Cursor(), (Position{Line: 0, Column: 7}); got != want {
		t.Fatalf("Cursor() = %+v, want %+v", got, want)
	}
	if !buffer.Dirty() {
		t.Fatal("edited buffer is not dirty")
	}
}

func TestBufferNewlineBackspaceAndDelete(t *testing.T) {
	buffer := NewBuffer("one\r\ntwo")
	buffer.SetCursor(Position{Line: 0, Column: 3})
	buffer.InsertNewline()
	buffer.InsertRune('x')
	if got, want := buffer.Text(), "one\nx\ntwo"; got != want {
		t.Fatalf("after insert = %q, want %q", got, want)
	}

	buffer.Backspace()
	buffer.Backspace()
	if got, want := buffer.Text(), "one\ntwo"; got != want {
		t.Fatalf("after backspace = %q, want %q", got, want)
	}
	if got, want := buffer.Cursor(), (Position{Line: 0, Column: 3}); got != want {
		t.Fatalf("Cursor() = %+v, want %+v", got, want)
	}

	buffer.Delete()
	if got, want := buffer.Text(), "onetwo"; got != want {
		t.Fatalf("after delete = %q, want %q", got, want)
	}
}

func TestBufferVerticalMovementKeepsGoalColumn(t *testing.T) {
	buffer := NewBuffer("abcdef\nx\n123456")
	buffer.SetCursor(Position{Line: 0, Column: 5})
	buffer.MoveVertical(1)
	if got := buffer.Cursor().Column; got != 1 {
		t.Fatalf("short line column = %d, want 1", got)
	}
	buffer.MoveVertical(1)
	if got := buffer.Cursor().Column; got != 5 {
		t.Fatalf("restored column = %d, want 5", got)
	}
}

func TestBufferSelectionAndPaste(t *testing.T) {
	buffer := NewBuffer("alpha\nbeta\ngamma")
	buffer.SetCursor(Position{Line: 0, Column: 2})
	buffer.ToggleMark()
	buffer.SetCursor(Position{Line: 2, Column: 3})

	selected, ok := buffer.SelectedText()
	if !ok || !reflect.DeepEqual(selected, []string{"pha", "beta", "gam"}) {
		t.Fatalf("SelectedText() = %q, %v", selected, ok)
	}
	clip, wholeLines := buffer.Cut()
	if wholeLines || !reflect.DeepEqual(clip, selected) {
		t.Fatalf("Cut() = %q, %v", clip, wholeLines)
	}
	if got, want := buffer.Text(), "alma"; got != want {
		t.Fatalf("after cut = %q, want %q", got, want)
	}

	buffer.Paste(clip, wholeLines)
	if got, want := buffer.Text(), "alpha\nbeta\ngamma"; got != want {
		t.Fatalf("after paste = %q, want %q", got, want)
	}
	if got, want := buffer.Cursor(), (Position{Line: 2, Column: 3}); got != want {
		t.Fatalf("Cursor() = %+v, want %+v", got, want)
	}
}

func TestBufferWholeLineCutAndPaste(t *testing.T) {
	buffer := NewBuffer("first\nsecond\nthird")
	buffer.SetCursor(Position{Line: 1, Column: 4})
	clip, wholeLines := buffer.Cut()
	if !wholeLines || !reflect.DeepEqual(clip, []string{"second"}) {
		t.Fatalf("Cut() = %q, %v", clip, wholeLines)
	}
	if got, want := buffer.Text(), "first\nthird"; got != want {
		t.Fatalf("after cut = %q, want %q", got, want)
	}
	buffer.Paste(clip, wholeLines)
	if got, want := buffer.Text(), "first\nsecond\nthird"; got != want {
		t.Fatalf("after paste = %q, want %q", got, want)
	}
	if got, want := buffer.Cursor(), (Position{Line: 1, Column: 6}); got != want {
		t.Fatalf("Cursor() = %+v, want %+v", got, want)
	}
}

func TestBufferMultilinePasteAtCursor(t *testing.T) {
	buffer := NewBuffer("ab")
	buffer.SetCursor(Position{Line: 0, Column: 1})
	buffer.Paste([]string{"12", "34"}, false)
	if got, want := buffer.Text(), "a12\n34b"; got != want {
		t.Fatalf("Text() = %q, want %q", got, want)
	}
	if got, want := buffer.Cursor(), (Position{Line: 1, Column: 2}); got != want {
		t.Fatalf("Cursor() = %+v, want %+v", got, want)
	}
}

func TestBufferReplaceLines(t *testing.T) {
	buffer := NewBuffer("before\na|b\n-| -\nafter")
	buffer.ReplaceLines(1, 2, []string{"| a | b |", "| --- | --- |"})
	if got, want := buffer.Text(), "before\n| a | b |\n| --- | --- |\nafter"; got != want {
		t.Fatalf("Text() = %q, want %q", got, want)
	}
	if got, want := buffer.Cursor(), (Position{Line: 1}); got != want {
		t.Fatalf("Cursor() = %+v, want %+v", got, want)
	}
}
