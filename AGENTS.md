# AGENTS.md

## Project overview

`mdt` (Markdown Tools) is a Go module containing two terminal utilities:

- `mdv`: a CLI Markdown viewer built with Glamour.
- `mde`: a minimal Nano-style Markdown editor built with tcell.

Both commands must remain cross-platform and cross-shell. They accept absolute or relative paths, must not invoke shell commands, and should work on Linux, macOS, and Windows without CGO. Do not add behavior tied to Bash, Fish, PowerShell, a particular terminal graphics protocol, or a single terminal emulator.

The module path is `mdt`. Use the Go version declared in `go.mod`.

## Common commands

Run from the repository root:

```sh
make fmt
make test
make race
make vet
make build
make check
```

The default `make`/`make build` target builds both utilities into `bin/`. Use `make mdv` or `make mde` for an individual binary, and `make help` to list all targets. The equivalent direct Go commands are `gofmt -w cmd internal`, `go test ./...`, `go test -race ./...`, `go vet ./...`, and `go build ./cmd/...`.

Run `make tidy` (or `go mod tidy`) after changing imports or dependencies. Format all modified Go files and keep `go.mod` and `go.sum` synchronized.

Check cross-platform compilation after platform-sensitive changes:

```sh
GOOS=linux GOARCH=amd64 go build ./cmd/...
GOOS=windows GOARCH=amd64 go build ./cmd/...
GOOS=darwin GOARCH=amd64 go build ./cmd/...
```

Useful smoke tests:

```sh
go run ./cmd/mdv README.md
go run ./cmd/mde README.md
```

Do not launch an interactive editor as part of automated tests. Use tcell's simulation screen for TUI tests. A real pseudo-terminal smoke test is useful after changing startup, terminal initialization, or key handling.

## Project layout

```text
cmd/mdv/             Thin viewer CLI entry point
cmd/mde/             Thin editor CLI entry point
internal/viewer/     Glamour configuration and terminal rendering
internal/editor/     Text buffer, inline highlighting, and tcell UI
internal/tablefmt/   Unicode-aware Markdown pipe-table formatter
.github/workflows/   CI and tag-triggered release automation
Makefile             Build, test, and release-package targets
README.md            User-facing build, usage, and shortcut documentation
LICENSE              The Unlicense
AGENTS.md            Development conventions and project invariants
```

Keep reusable logic in `internal`, not in the command packages. Prefer focused, table-driven tests alongside the package being changed.

## Design notes and invariants

### Viewer

- Use Glamour v2 APIs from `charm.land/glamour/v2`; older examples using `WithAutoStyle` or `WithColorProfile` do not apply.
- Detect terminal dimensions in the CLI and pass an explicit width to the renderer. Redirected output uses an 80-column fallback.
- `GLAMOUR_STYLE` accepts a built-in style or JSON style path. Interactive output auto-selects dark/light styling when no style is set.
- Redirected `notty` output must not contain ANSI or OSC escape sequences. `internal/viewer` strips them and removes trailing padding.
- Images are text placeholders with hyperlink destinations. Do not fetch remote images or require Kitty, Sixel, iTerm, or similar graphics protocols.
- Portable terminals have no standard application-controlled font sizing. Preserve the current heading emphasis approach rather than adding terminal-specific font-size escapes.
- Tables intentionally use Glamour's full-width layout and wrapping behavior.

### Editor

- Keep Markdown source visible, including formatting markers. Highlighting adds WYSIWYG hints; it must not hide or rewrite source text.
- `Buffer` stores lines without newline terminators. Cursor columns are rune offsets, not byte offsets or terminal cells. Use display-width helpers only for rendering and mouse positioning.
- Preserve dirty-state behavior and unsaved-exit confirmation. A nonexistent input file is an empty buffer until the first save.
- Input CRLF/CR line endings are normalized to LF in memory and when saved.
- Tables stay raw except when the user explicitly invokes `F6` or `Alt+T` for the selected rows or table under the cursor.
- Nano's `Ctrl+^` is `tcell.KeyCtrlCarat`; `Ctrl+Space` and `Alt+A` are also selection shortcuts. Keep `F6` as a fallback for terminals that do not transmit `Alt+T` reliably.
- Use `tcell.NewSimulationScreen` for drawing, keyboard, save/exit, and formatting tests. Test UTF-8 editing and cursor movement without requiring a real terminal.

### Table formatter

- `tablefmt.Format` accepts lines without newline characters and returns a new slice without modifying its input.
- Column widths use terminal display width (`go-runewidth`), not byte or rune counts.
- Preserve cell alignment markers, escaped pipes, and pipes inside backtick code spans. Normalize rows to outer pipes and one-space cell padding.
- Keep the guard that distinguishes one-column pipe tables from Setext headings.
- Add tests for malformed input, Unicode width, optional outer pipes, missing cells, and input immutability when changing parsing behavior.

### Build and release automation

- Keep `make mdv`, `make mde`, and `make build` working independently. `GOOS` and `GOARCH` select the target, and Windows builds must gain an `.exe` suffix.
- `make release VERSION=x.y.z` runs native checks and packages both binaries, `README.md`, and `LICENSE` for each entry in `RELEASE_PLATFORMS`. Unix uses `.tar.gz`, Windows uses `.zip`, and `SHA256SUMS` covers every archive. Packaging requires a POSIX shell, `tar`, `zip`, and `sha256sum`.
- The default platforms are Linux, macOS, and Windows, each on AMD64 and ARM64. When changing platform support, update `RELEASE_PLATFORMS`, the CI matrix, and the README together.
- `.github/workflows/ci.yml` handles pushes, pull requests, and manual runs. It performs native checks and race tests before cross-building and uploading artifacts.
- `.github/workflows/release.yml` runs only for `v*` tags. It uses the Makefile release target and GitHub CLI to create a release with generated notes. Keep `contents: write` limited to this release workflow.
- Test Makefile changes with individual builds, a combined build, and at least one limited-platform release. Use a temporary `BIN_DIR`/`DIST_DIR` when validating packaging to avoid deleting a developer's existing output.

## Development guidelines

- Make minimal, idiomatic changes and follow existing package boundaries.
- Prefer established open-source libraries over new handwritten parsers or terminal backends. Keep the dependency set small.
- Add or update tests for behavior changes. Run the full test suite, vet, and relevant builds before finishing.
- Update `README.md` whenever user-facing behavior, installation, options, or shortcuts change.
- Update this file whenever architecture, commands, dependencies, or development invariants change.
- The project is released under The Unlicense; see `LICENSE`.
