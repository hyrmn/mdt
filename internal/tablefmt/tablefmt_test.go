package tablefmt

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestFormatAlignsAndNormalizes(t *testing.T) {
	input := []string{
		"| h | left | right | center |",
		"| --- | :--- | ---: | :---: |",
		"| abcde | x | 42 | z |",
	}
	want := []string{
		"| h     | left | right | center |",
		"| ----- | :--- | ----: | :----: |",
		"| abcde | x    | 42    | z      |",
	}

	got, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	assertLinesEqual(t, got, want)
}

func TestFormatMinimumDelimiterWidths(t *testing.T) {
	input := []string{
		"a|b|c|d",
		"---|:---|---:|:---: |",
	}
	want := []string{
		"| a   | b    | c    | d     |",
		"| --- | :--- | ---: | :---: |",
	}

	got, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	assertLinesEqual(t, got, want)
}

func TestFormatOptionalOuterPipesAndWhitespace(t *testing.T) {
	input := []string{
		"  Header A  | Header B  ",
		" --- | ---: ",
		"  a | b  ",
	}
	want := []string{
		"| Header A | Header B |",
		"| -------- | -------: |",
		"| a        | b        |",
	}

	got, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	assertLinesEqual(t, got, want)
}

func TestFormatPreservesEscapedPipesAndCodeSpans(t *testing.T) {
	input := []string{
		"| Literal \\| pipe | Code | More |",
		"| --- | --- | --- |",
		"| a \\| b | `x|y` | plain |",
	}
	want := []string{
		"| Literal \\| pipe | Code  | More  |",
		"| " + strings.Repeat("-", 15) + " | ----- | ----- |",
		"| a \\| b          | `x|y` | plain |",
	}

	got, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	assertLinesEqual(t, got, want)
}

func TestFormatBacktickRuns(t *testing.T) {
	input := []string{
		"| ``a|`b`` | c |",
		"| --- | --- |",
		"| `x` | y |",
	}
	want := []string{
		"| ``a|`b`` | c   |",
		"| -------- | --- |",
		"| `x`      | y   |",
	}

	got, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	assertLinesEqual(t, got, want)
}

func TestFormatUnclosedBacktickDoesNotHidePipe(t *testing.T) {
	input := []string{
		"a `literal | b",
		"--- | ---",
	}
	want := []string{
		"| a `literal | b   |",
		"| ---------- | --- |",
	}

	got, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	assertLinesEqual(t, got, want)
}

func TestFormatUnicodeDisplayWidth(t *testing.T) {
	input := []string{
		"| 汉字 | e\u0301 | 🙂 |",
		"| --- | --- | --- |",
		"| x | y | z |",
	}
	want := []string{
		"| 汉字 | e\u0301   | 🙂  |",
		"| ---- | --- | --- |",
		"| x    | y   | z   |",
	}

	got, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	assertLinesEqual(t, got, want)
	for _, line := range got {
		if width := runewidth.StringWidth(line); width != runewidth.StringWidth(got[0]) {
			t.Errorf("line display width = %d, want %d: %q", width, runewidth.StringWidth(got[0]), line)
		}
	}
}

func TestFormatPadsMissingCellsAndRetainsExtraCells(t *testing.T) {
	input := []string{
		"| a | b |",
		"| --- | --- |",
		"| only |",
		"| 1 | 2 | 3 | 4 |",
	}
	want := []string{
		"| a    | b   |     |     |",
		"| ---- | --- | --- | --- |",
		"| only |     |     |     |",
		"| 1    | 2   | 3   | 4   |",
	}

	got, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	assertLinesEqual(t, got, want)
}

func TestFormatPreservesSurroundingLines(t *testing.T) {
	input := []string{
		"",
		"preface",
		"  ",
		"| a | b |",
		"| --- | --- |",
		"| c | d |",
		"\t",
		"after",
		"",
	}
	want := []string{
		"",
		"preface",
		"  ",
		"| a   | b   |",
		"| --- | --- |",
		"| c   | d   |",
		"\t",
		"after",
		"",
	}

	got, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	assertLinesEqual(t, got, want)
}

func TestFormatTrimsOnlyOuterCellWhitespace(t *testing.T) {
	input := []string{
		"|  a   b  |  c  d  |",
		"| --- | --- |",
		"| x | y |",
	}
	want := []string{
		"| a   b | c  d |",
		"| ----- | ---- |",
		"| x     | y    |",
	}

	got, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	assertLinesEqual(t, got, want)
}

func TestFormatDoesNotModifyInput(t *testing.T) {
	input := []string{"| a | b |", "| --- | --- |", "| c | d |"}
	original := append([]string(nil), input...)

	if _, err := Format(input); err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	assertLinesEqual(t, input, original)
}

func TestFormatRejectsInvalidTables(t *testing.T) {
	tests := map[string][]string{
		"nil":                     nil,
		"empty":                   {},
		"blank lines":             {"", "  ", "\t"},
		"header only":             {"| a | b |"},
		"setext heading":          {"Title", "---"},
		"missing delimiter":       {"| a | b |", "| c | d |"},
		"too few hyphens":         {"| a | b |", "| -- | --- |"},
		"non-hyphen delimiter":    {"| a | b |", "| --- | -x- |"},
		"empty delimiter cell":    {"| a | b |", "| --- | |"},
		"internal delimiter text": {"| a | b |", "| --- | --- x |"},
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Format(input)
			if err == nil {
				t.Fatalf("Format() = %q, want an error", got)
			}
			if !strings.Contains(err.Error(), "header") || !strings.Contains(err.Error(), "delimiter") {
				t.Errorf("Format() error = %q, want a descriptive table error", err)
			}
		})
	}
}

func assertLinesEqual(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("lines mismatch\n got: %q\nwant: %q", got, want)
	}
}
