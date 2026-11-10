package editor

import (
	"strings"
	"unicode/utf8"
)

// Position identifies a cursor location between runes in the buffer.
type Position struct {
	Line   int
	Column int
}

func (p Position) before(other Position) bool {
	return p.Line < other.Line || (p.Line == other.Line && p.Column < other.Column)
}

// Buffer is a UTF-8 text buffer. Columns are rune offsets, not byte offsets.
type Buffer struct {
	lines      []string
	cursor     Position
	mark       *Position
	goalColumn int
	goalValid  bool
	dirty      bool
}

func NewBuffer(text string) *Buffer {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	return &Buffer{lines: lines}
}

func (b *Buffer) Lines() []string {
	return b.lines
}

func (b *Buffer) Text() string {
	return strings.Join(b.lines, "\n")
}

func (b *Buffer) Cursor() Position {
	return b.cursor
}

func (b *Buffer) Dirty() bool {
	return b.dirty
}

func (b *Buffer) SetDirty(dirty bool) {
	b.dirty = dirty
}

func (b *Buffer) Mark() (Position, bool) {
	if b.mark == nil {
		return Position{}, false
	}
	return *b.mark, true
}

func (b *Buffer) ToggleMark() {
	if b.mark != nil {
		b.mark = nil
		return
	}
	mark := b.cursor
	b.mark = &mark
}

func (b *Buffer) ClearMark() {
	b.mark = nil
}

// Selection returns the selected range in reading order. The end is exclusive.
func (b *Buffer) Selection() (Position, Position, bool) {
	if b.mark == nil {
		return Position{}, Position{}, false
	}
	start, end := *b.mark, b.cursor
	if end.before(start) {
		start, end = end, start
	}
	return start, end, start != end
}

func (b *Buffer) SetCursor(position Position) {
	b.cursor = b.clamp(position)
	b.goalValid = false
}

func (b *Buffer) clamp(position Position) Position {
	if position.Line < 0 {
		position.Line = 0
	}
	if position.Line >= len(b.lines) {
		position.Line = len(b.lines) - 1
	}
	if position.Column < 0 {
		position.Column = 0
	}
	lineLength := utf8.RuneCountInString(b.lines[position.Line])
	if position.Column > lineLength {
		position.Column = lineLength
	}
	return position
}

func (b *Buffer) MoveHorizontal(delta int) {
	if delta < 0 && b.cursor.Column == 0 && b.cursor.Line > 0 {
		b.cursor.Line--
		b.cursor.Column = utf8.RuneCountInString(b.lines[b.cursor.Line])
	} else if delta > 0 && b.cursor.Column == utf8.RuneCountInString(b.lines[b.cursor.Line]) && b.cursor.Line < len(b.lines)-1 {
		b.cursor.Line++
		b.cursor.Column = 0
	} else {
		b.cursor.Column += delta
	}
	b.cursor = b.clamp(b.cursor)
	b.goalValid = false
}

func (b *Buffer) MoveVertical(delta int) {
	if !b.goalValid {
		b.goalColumn = b.cursor.Column
		b.goalValid = true
	}
	b.cursor = b.clamp(Position{Line: b.cursor.Line + delta, Column: b.goalColumn})
	b.goalValid = true
}

func (b *Buffer) MoveLineStart() {
	b.cursor.Column = 0
	b.goalValid = false
}

func (b *Buffer) MoveLineEnd() {
	b.cursor.Column = utf8.RuneCountInString(b.lines[b.cursor.Line])
	b.goalValid = false
}

func (b *Buffer) MoveDocumentStart() {
	b.SetCursor(Position{})
}

func (b *Buffer) MoveDocumentEnd() {
	b.SetCursor(Position{Line: len(b.lines) - 1, Column: utf8.RuneCountInString(b.lines[len(b.lines)-1])})
}

func runeSlice(line string) []rune {
	return []rune(line)
}

func (b *Buffer) deleteSelection() bool {
	start, end, ok := b.Selection()
	if !ok {
		return false
	}

	startRunes := runeSlice(b.lines[start.Line])
	endRunes := runeSlice(b.lines[end.Line])
	if start.Line == end.Line {
		b.lines[start.Line] = string(startRunes[:start.Column]) + string(startRunes[end.Column:])
	} else {
		b.lines[start.Line] = string(startRunes[:start.Column]) + string(endRunes[end.Column:])
		b.lines = append(b.lines[:start.Line+1], b.lines[end.Line+1:]...)
	}
	b.cursor = start
	b.mark = nil
	b.goalValid = false
	b.dirty = true
	return true
}

func (b *Buffer) InsertRune(r rune) {
	b.deleteSelection()
	runes := runeSlice(b.lines[b.cursor.Line])
	column := b.cursor.Column
	runes = append(runes, 0)
	copy(runes[column+1:], runes[column:])
	runes[column] = r
	b.lines[b.cursor.Line] = string(runes)
	b.cursor.Column++
	b.goalValid = false
	b.dirty = true
}

func (b *Buffer) InsertNewline() {
	b.deleteSelection()
	runes := runeSlice(b.lines[b.cursor.Line])
	column := b.cursor.Column
	before, after := string(runes[:column]), string(runes[column:])
	b.lines[b.cursor.Line] = before
	b.lines = append(b.lines, "")
	copy(b.lines[b.cursor.Line+2:], b.lines[b.cursor.Line+1:])
	b.lines[b.cursor.Line+1] = after
	b.cursor.Line++
	b.cursor.Column = 0
	b.goalValid = false
	b.dirty = true
}

func (b *Buffer) Backspace() {
	if b.deleteSelection() {
		return
	}
	if b.cursor.Column == 0 {
		if b.cursor.Line == 0 {
			return
		}
		previousLength := utf8.RuneCountInString(b.lines[b.cursor.Line-1])
		b.lines[b.cursor.Line-1] += b.lines[b.cursor.Line]
		b.lines = append(b.lines[:b.cursor.Line], b.lines[b.cursor.Line+1:]...)
		b.cursor.Line--
		b.cursor.Column = previousLength
	} else {
		runes := runeSlice(b.lines[b.cursor.Line])
		column := b.cursor.Column
		b.lines[b.cursor.Line] = string(runes[:column-1]) + string(runes[column:])
		b.cursor.Column--
	}
	b.goalValid = false
	b.dirty = true
}

func (b *Buffer) Delete() {
	if b.deleteSelection() {
		return
	}
	runes := runeSlice(b.lines[b.cursor.Line])
	if b.cursor.Column < len(runes) {
		column := b.cursor.Column
		b.lines[b.cursor.Line] = string(runes[:column]) + string(runes[column+1:])
	} else if b.cursor.Line < len(b.lines)-1 {
		b.lines[b.cursor.Line] += b.lines[b.cursor.Line+1]
		b.lines = append(b.lines[:b.cursor.Line+1], b.lines[b.cursor.Line+2:]...)
	} else {
		return
	}
	b.goalValid = false
	b.dirty = true
}

// SelectedText returns the selected text without changing the buffer.
func (b *Buffer) SelectedText() ([]string, bool) {
	start, end, ok := b.Selection()
	if !ok {
		return nil, false
	}
	if start.Line == end.Line {
		runes := runeSlice(b.lines[start.Line])
		return []string{string(runes[start.Column:end.Column])}, true
	}

	selected := make([]string, 0, end.Line-start.Line+1)
	first := runeSlice(b.lines[start.Line])
	selected = append(selected, string(first[start.Column:]))
	selected = append(selected, b.lines[start.Line+1:end.Line]...)
	last := runeSlice(b.lines[end.Line])
	selected = append(selected, string(last[:end.Column]))
	return selected, true
}

// Cut removes the selection, or the whole current line when there is no selection.
// The boolean result reports whether the clipboard contains whole lines.
func (b *Buffer) Cut() ([]string, bool) {
	if selected, ok := b.SelectedText(); ok {
		b.deleteSelection()
		return selected, false
	}

	line := b.lines[b.cursor.Line]
	if len(b.lines) == 1 {
		b.lines[0] = ""
	} else {
		b.lines = append(b.lines[:b.cursor.Line], b.lines[b.cursor.Line+1:]...)
		if b.cursor.Line >= len(b.lines) {
			b.cursor.Line = len(b.lines) - 1
		}
	}
	b.cursor.Column = 0
	b.goalValid = false
	b.dirty = true
	return []string{line}, true
}

// Paste inserts clipboard text. Whole-line clips are inserted above the cursor.
func (b *Buffer) Paste(lines []string, wholeLines bool) {
	if len(lines) == 0 {
		return
	}
	b.deleteSelection()
	if wholeLines {
		line := b.cursor.Line
		inserted := append([]string(nil), lines...)
		b.lines = append(b.lines, make([]string, len(inserted))...)
		copy(b.lines[line+len(inserted):], b.lines[line:len(b.lines)-len(inserted)])
		copy(b.lines[line:], inserted)
		b.cursor = Position{Line: line + len(inserted) - 1, Column: utf8.RuneCountInString(inserted[len(inserted)-1])}
	} else {
		before := runeSlice(b.lines[b.cursor.Line])
		column := b.cursor.Column
		prefix, suffix := string(before[:column]), string(before[column:])
		if len(lines) == 1 {
			b.lines[b.cursor.Line] = prefix + lines[0] + suffix
			b.cursor.Column += utf8.RuneCountInString(lines[0])
		} else {
			replacement := make([]string, len(lines))
			replacement[0] = prefix + lines[0]
			copy(replacement[1:], lines[1:])
			replacement[len(replacement)-1] += suffix
			b.lines = append(b.lines[:b.cursor.Line], append(replacement, b.lines[b.cursor.Line+1:]...)...)
			b.cursor.Line += len(lines) - 1
			b.cursor.Column = utf8.RuneCountInString(lines[len(lines)-1])
		}
	}
	b.goalValid = false
	b.dirty = true
}

// ReplaceLines replaces an inclusive line range and places the cursor at its start.
func (b *Buffer) ReplaceLines(start, end int, replacement []string) {
	if start < 0 || end < start || end >= len(b.lines) {
		return
	}
	updated := make([]string, 0, len(b.lines)-(end-start+1)+len(replacement))
	updated = append(updated, b.lines[:start]...)
	updated = append(updated, replacement...)
	updated = append(updated, b.lines[end+1:]...)
	b.lines = updated
	if len(b.lines) == 0 {
		b.lines = []string{""}
	}
	b.cursor = b.clamp(Position{Line: start})
	b.mark = nil
	b.goalValid = false
	b.dirty = true
}
