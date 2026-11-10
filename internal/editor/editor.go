package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"

	"mdt/internal/tablefmt"
)

const tabWidth = 4

type clipboard struct {
	lines      []string
	wholeLines bool
}

// Editor owns the terminal user interface and its text buffer.
type Editor struct {
	screen      tcell.Screen
	path        string
	buffer      *Buffer
	top         int
	left        int
	status      string
	clipboard   clipboard
	showHelp    bool
	confirmQuit bool
	quit        bool
}

// Run opens path and starts the interactive editor. A missing file is created on save.
func Run(path string) error {
	cleanPath := filepath.Clean(path)
	contents, err := os.ReadFile(cleanPath)
	status := "Ctrl+G for help"
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("open %s: %w", cleanPath, err)
		}
		contents = nil
		status = "New file (save to create it)"
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("initialize terminal: %w", err)
	}
	if err := screen.Init(); err != nil {
		return fmt.Errorf("initialize terminal: %w", err)
	}
	defer screen.Fini()

	screen.SetTitle("mde - " + cleanPath)
	screen.EnablePaste()
	screen.EnableMouse()
	editor := &Editor{
		screen: screen,
		path:   cleanPath,
		buffer: NewBuffer(string(contents)),
		status: status,
	}
	for !editor.quit {
		editor.draw()
		event := screen.PollEvent()
		if event == nil {
			break
		}
		editor.handleEvent(event)
	}
	return nil
}

func (e *Editor) handleEvent(event tcell.Event) {
	switch event := event.(type) {
	case *tcell.EventResize:
		e.screen.Sync()
	case *tcell.EventKey:
		e.handleKey(event)
	case *tcell.EventMouse:
		e.handleMouse(event)
	}
}

func (e *Editor) handleKey(event *tcell.EventKey) {
	if e.showHelp {
		e.showHelp = false
		return
	}

	if e.confirmQuit {
		switch event.Key() {
		case tcell.KeyCtrlS:
			if e.save() {
				e.quit = true
			}
		case tcell.KeyCtrlX, tcell.KeyCtrlC:
			e.quit = true
		case tcell.KeyEscape:
			e.confirmQuit = false
			e.status = "Exit cancelled"
		default:
			e.status = "Unsaved: Ctrl+S save & exit, Ctrl+X discard, Esc cancel"
		}
		return
	}

	if event.Key() == tcell.KeyRune && event.Modifiers()&tcell.ModAlt != 0 {
		switch event.Rune() {
		case 't', 'T':
			e.formatTable()
			return
		case 'a', 'A':
			e.buffer.ToggleMark()
			e.selectionStatus()
			return
		}
	}

	switch event.Key() {
	case tcell.KeyCtrlX, tcell.KeyCtrlC:
		if e.buffer.Dirty() {
			e.confirmQuit = true
			e.status = "Unsaved: Ctrl+S save & exit, Ctrl+X discard, Esc cancel"
		} else {
			e.quit = true
		}
	case tcell.KeyCtrlO, tcell.KeyCtrlS:
		e.save()
	case tcell.KeyCtrlG:
		e.showHelp = true
	case tcell.KeyCtrlK:
		lines, wholeLines := e.buffer.Cut()
		e.clipboard = clipboard{lines: lines, wholeLines: wholeLines}
		e.screen.SetClipboard([]byte(strings.Join(lines, "\n")))
		e.status = "Cut"
	case tcell.KeyCtrlU:
		e.buffer.Paste(e.clipboard.lines, e.clipboard.wholeLines)
		e.status = "Pasted"
	case tcell.KeyCtrlCarat, tcell.KeyCtrlSpace:
		e.buffer.ToggleMark()
		e.selectionStatus()
	case tcell.KeyF6:
		e.formatTable()
	case tcell.KeyUp:
		e.buffer.MoveVertical(-1)
	case tcell.KeyDown:
		e.buffer.MoveVertical(1)
	case tcell.KeyLeft:
		e.buffer.MoveHorizontal(-1)
	case tcell.KeyRight:
		e.buffer.MoveHorizontal(1)
	case tcell.KeyPgUp:
		e.buffer.MoveVertical(-e.pageSize())
	case tcell.KeyPgDn:
		e.buffer.MoveVertical(e.pageSize())
	case tcell.KeyHome:
		if event.Modifiers()&tcell.ModCtrl != 0 {
			e.buffer.MoveDocumentStart()
		} else {
			e.buffer.MoveLineStart()
		}
	case tcell.KeyEnd:
		if event.Modifiers()&tcell.ModCtrl != 0 {
			e.buffer.MoveDocumentEnd()
		} else {
			e.buffer.MoveLineEnd()
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		e.buffer.Backspace()
	case tcell.KeyDelete:
		e.buffer.Delete()
	case tcell.KeyEnter:
		e.buffer.InsertNewline()
	case tcell.KeyTab:
		for i := 0; i < tabWidth; i++ {
			e.buffer.InsertRune(' ')
		}
	case tcell.KeyRune:
		if event.Rune() != 0 {
			e.buffer.InsertRune(event.Rune())
		}
	}
}

func (e *Editor) handleMouse(event *tcell.EventMouse) {
	x, y := event.Position()
	if event.Buttons()&tcell.Button1 == 0 || y < 1 || y > e.contentHeight() {
		return
	}
	line := e.top + y - 1
	if line >= len(e.buffer.Lines()) {
		return
	}
	column := e.columnAtX(e.buffer.Lines()[line], e.left, x)
	e.buffer.SetCursor(Position{Line: line, Column: column})
}

func (e *Editor) selectionStatus() {
	if _, ok := e.buffer.Mark(); ok {
		e.status = "Selection started; move the cursor, then Alt+T/F6 to format a table"
	} else {
		e.status = "Selection cleared"
	}
}

func (e *Editor) save() bool {
	if err := os.WriteFile(e.path, []byte(e.buffer.Text()), 0o644); err != nil {
		e.status = "Save failed: " + err.Error()
		return false
	}
	e.buffer.SetDirty(false)
	e.confirmQuit = false
	e.status = fmt.Sprintf("Wrote %d lines to %s", len(e.buffer.Lines()), e.path)
	return true
}

func (e *Editor) formatTable() {
	lines := e.buffer.Lines()
	cursor := e.buffer.Cursor()
	start, end := cursor.Line, cursor.Line
	if mark, ok := e.buffer.Mark(); ok {
		start, end = mark.Line, cursor.Line
		if end < start {
			start, end = end, start
		}
	} else {
		for start > 0 && looksLikeTableLine(lines[start-1]) {
			start--
		}
		for end+1 < len(lines) && looksLikeTableLine(lines[end+1]) {
			end++
		}
	}

	selected := append([]string(nil), lines[start:end+1]...)
	formatted, err := tablefmt.Format(selected)
	if err != nil {
		e.status = "Table format: " + err.Error()
		return
	}
	e.buffer.ReplaceLines(start, end, formatted)
	e.status = "Table formatted"
}

func looksLikeTableLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed != "" && strings.Contains(trimmed, "|")
}

func (e *Editor) pageSize() int {
	if height := e.contentHeight(); height > 1 {
		return height - 1
	}
	return 1
}

func (e *Editor) contentHeight() int {
	_, height := e.screen.Size()
	height -= 4
	if height < 1 {
		return 1
	}
	return height
}

func (e *Editor) ensureCursorVisible() {
	width, _ := e.screen.Size()
	height := e.contentHeight()
	cursor := e.buffer.Cursor()
	if cursor.Line < e.top {
		e.top = cursor.Line
	}
	if cursor.Line >= e.top+height {
		e.top = cursor.Line - height + 1
	}
	if e.top < 0 {
		e.top = 0
	}
	if cursor.Column < e.left {
		e.left = cursor.Column
	}
	if width < 2 {
		e.left = cursor.Column
		return
	}
	for e.left < cursor.Column && e.prefixWidth(e.buffer.Lines()[cursor.Line], e.left, cursor.Column) > width-1 {
		e.left++
	}
}

func (e *Editor) draw() {
	e.screen.Clear()
	if e.showHelp {
		e.drawHelp()
		e.screen.Show()
		return
	}

	e.ensureCursorVisible()
	width, height := e.screen.Size()
	title := " mde  " + e.path
	if e.buffer.Dirty() {
		title += "  [Modified]"
	}
	e.drawText(0, 0, width, title, tcell.StyleDefault.Reverse(true).Bold(true))
	for x := runewidth.StringWidth(title); x < width; x++ {
		e.screen.SetContent(x, 0, ' ', nil, tcell.StyleDefault.Reverse(true))
	}

	contentHeight := e.contentHeight()
	selectionStart, selectionEnd, hasSelection := e.buffer.Selection()
	for row := 0; row < contentHeight && e.top+row < len(e.buffer.Lines()); row++ {
		lineIndex := e.top + row
		line := e.buffer.Lines()[lineIndex]
		e.drawLine(line, lineIndex, row+1, width, selectionStart, selectionEnd, hasSelection)
	}

	statusStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	if height >= 3 {
		e.drawText(0, height-3, width, " "+e.status, statusStyle)
	}
	if height >= 2 {
		e.drawText(0, height-2, width, "^X Exit   ^O Save   ^K Cut    ^U Paste  ^^/Alt+A Select", tcell.StyleDefault.Reverse(true))
	}
	if height >= 1 {
		e.drawText(0, height-1, width, "^G Help   F6/Alt+T Format table   Arrows Move   ^S Save", tcell.StyleDefault.Reverse(true))
	}

	cursor := e.buffer.Cursor()
	cursorX := e.prefixWidth(e.buffer.Lines()[cursor.Line], e.left, cursor.Column)
	e.screen.ShowCursor(cursorX, cursor.Line-e.top+1)
	e.screen.Show()
}

func (e *Editor) drawLine(line string, lineIndex, y, width int, selectionStart, selectionEnd Position, hasSelection bool) {
	runes := []rune(line)
	styles := inlineStyles(line)
	x := 0
	for column := e.left; column < len(runes); column++ {
		r := runes[column]
		style := styles[column]
		position := Position{Line: lineIndex, Column: column}
		if hasSelection && !position.before(selectionStart) && position.before(selectionEnd) {
			style = style.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
		}
		if r == '\t' {
			spaces := tabWidth - x%tabWidth
			for i := 0; i < spaces && x+i < width; i++ {
				e.screen.SetContent(x+i, y, ' ', nil, style)
			}
			x += spaces
		} else {
			if x >= width {
				break
			}
			e.screen.SetContent(x, y, r, nil, style)
			x += runewidth.RuneWidth(r)
		}
		if x >= width {
			break
		}
	}
}

func (e *Editor) drawText(x, y, width int, text string, style tcell.Style) {
	for _, r := range text {
		if x >= width {
			return
		}
		if r == '\n' || r == '\r' {
			continue
		}
		e.screen.SetContent(x, y, r, nil, style)
		x += runewidth.RuneWidth(r)
	}
}

func (e *Editor) drawHelp() {
	e.screen.HideCursor()
	width, height := e.screen.Size()
	lines := []string{
		"mde — Markdown Editor",
		"",
		"Ctrl+X / Ctrl+C   Exit (prompts for unsaved changes)",
		"Ctrl+O / Ctrl+S   Save",
		"Ctrl+K            Cut selection or current line",
		"Ctrl+U            Paste last cut",
		"Ctrl+^ / Alt+A    Start or clear a selection",
		"Alt+T / F6        Format selected table or table at cursor",
		"Ctrl+G            Open this help",
		"Arrows            Move; Page Up/Down moves by a page",
		"Home/End          Line start/end; add Ctrl for document",
		"Tab               Insert four spaces",
		"",
		"Markdown markers remain visible with inline styling.",
		"Press any key to return.",
	}
	startY := 0
	if height > len(lines) {
		startY = (height - len(lines)) / 2
	}
	for i, line := range lines {
		x := 0
		if width > len([]rune(line))+4 {
			x = (width - len([]rune(line))) / 2
		}
		style := tcell.StyleDefault
		if i == 0 {
			style = style.Bold(true).Foreground(tcell.ColorAqua)
		}
		e.drawText(x, startY+i, width, line, style)
	}
}

func (e *Editor) prefixWidth(line string, start, end int) int {
	runes := []rune(line)
	if start > len(runes) {
		start = len(runes)
	}
	if end > len(runes) {
		end = len(runes)
	}
	width := 0
	for _, r := range runes[start:end] {
		width += runeWidth(r, width)
	}
	return width
}

func (e *Editor) columnAtX(line string, left, target int) int {
	runes := []rune(line)
	x := 0
	for column := left; column < len(runes); column++ {
		width := runeWidth(runes[column], x)
		if target < x+(width+1)/2 {
			return column
		}
		x += width
		if x > target {
			return column + 1
		}
	}
	return len(runes)
}

func runeWidth(r rune, column int) int {
	if r == '\t' {
		return tabWidth - column%tabWidth
	}
	return runewidth.RuneWidth(r)
}
