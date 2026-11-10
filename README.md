# Markdown Tools (`mdt`)

Two cross-platform Go utilities for working with Markdown:

- **`mdv` — Markdown Viewer:** renders Markdown as styled text in a terminal.
- **`mde` — Markdown Editor:** a minimal Nano-style TUI with inline Markdown styling and table formatting.

Both commands accept absolute paths and paths relative to the current working directory. They are pure Go and do not invoke a shell, so the same binaries work from Bash, Fish, Zsh, Command Prompt, and PowerShell.

## Build

Go 1.27.3 or newer is required by this module.

```sh
go install ./cmd/mdv ./cmd/mde
```

This installs both commands in Go's binary directory (`GOBIN`, or `GOPATH/bin` by default). Make sure that directory is on your `PATH`.

To build local binaries instead on Linux or macOS:

```sh
go build -o mdv ./cmd/mdv
go build -o mde ./cmd/mde
```

On Windows PowerShell:

```powershell
go build -o mdv.exe .\cmd\mdv
go build -o mde.exe .\cmd\mde
```

### Make targets

The included GNU Makefile can build each utility separately or together. Build output goes into `bin/` by default:

```sh
make          # Build both mdv and mde
make build    # Same as make
make mdv      # Build only bin/mdv
make mde      # Build only bin/mde
```

Useful development and packaging targets:

| Target | Purpose |
| --- | --- |
| `make help` | List the available targets |
| `make install` | Install both commands with `go install` |
| `make test` | Run all unit and simulated-TUI tests |
| `make race` | Run tests with the race detector |
| `make vet` | Run `go vet` |
| `make fmt` | Format all Go source files |
| `make fmt-check` | Fail if any Go source is unformatted |
| `make tidy` | Run `go mod tidy` |
| `make check` | Check formatting, vet, test, and build natively |
| `make release VERSION=0.1.0` | Build all release archives and checksums |
| `make clean` | Remove `bin/` and `dist/` |

`GO`, `BIN_DIR`, `DIST_DIR`, `GOFLAGS`, `LDFLAGS`, and `RELEASE_PLATFORMS` can be overridden as Make variables. `GOOS` and `GOARCH` select the target platform. For example:

```sh
make build BIN_DIR=output
make build GOOS=windows GOARCH=amd64
make mdv GOOS=darwin GOARCH=arm64
```

`make release VERSION=0.1.0` runs the native checks and creates archives in `dist/` for Linux, macOS, and Windows on AMD64 and ARM64. Unix archives use `.tar.gz`; Windows archives use `.zip`. Each archive contains both utilities, this README, and the license. `dist/SHA256SUMS` contains checksums for every archive. Override `RELEASE_PLATFORMS` with a space-separated list such as `linux/amd64 windows/amd64` to limit the targets.

The Makefile uses a POSIX-compatible shell. The direct Go build commands above are the simplest native Windows option. Release packaging additionally requires `tar`, `zip`, and `sha256sum`; these are available on GitHub's Ubuntu runners.

## `mdv`: view Markdown

```sh
mdv README.md
mdv ./docs/guide.md
mdv /absolute/path/to/guide.md
```

PowerShell examples:

```powershell
.\mdv.exe .\README.md
.\mdv.exe 'C:\docs\guide.md'
```

`mdv` uses [Glamour](https://github.com/charmbracelet/glamour) and supports common Markdown and GitHub-flavored Markdown constructs, including headings, emphasis, lists, links, fenced code blocks, block quotes, task lists, and tables. Output wraps to the current terminal width, with an 80-column fallback for redirected output. Styles adapt to a dark or light terminal background.

Set `GLAMOUR_STYLE` to a built-in Glamour style (`dark`, `light`, `ascii`, `notty`, `pink`, `dracula`, or `tokyo-night`) or to a JSON style file:

```sh
GLAMOUR_STYLE=light mdv README.md
```

```powershell
$env:GLAMOUR_STYLE = 'light'
.\mdv.exe .\README.md
```

### Tables and images

Tables are laid out with aligned columns and wrapped when necessary. Images are represented by a readable `[Image: alt-text]` placeholder followed by the destination; destinations can be terminal hyperlinks when supported. `mdv` deliberately does not fetch remote images or require Kitty, iTerm, Sixel, or another platform-specific graphics protocol.

Portable terminals do not provide a standard way for applications to change font size. Headings therefore use **bold**, color, backgrounds, prefixes, and vertical spacing to create a larger visual hierarchy. Actual font size remains a terminal emulator setting.

## `mde`: edit Markdown

```sh
mde README.md
mde ./docs/new-file.md
```

A missing file starts as an empty buffer and is created when saved. `mde` displays the original Markdown, including its markers, with WYSIWYG hints: headings and bold text are bold, emphasis is italic, inline code is colored, links are underlined, and strikethrough is styled where the terminal supports it. Tables remain in their raw Markdown form.

### Keyboard shortcuts

| Shortcut | Action |
| --- | --- |
| `Ctrl+X` or `Ctrl+C` | Exit; prompts before discarding unsaved changes |
| `Ctrl+O` or `Ctrl+S` | Save |
| `Ctrl+K` | Cut the selection, or the current line if there is no selection |
| `Ctrl+U` | Paste the last cut text |
| `Ctrl+^` or `Alt+A` | Start or clear a selection |
| `Alt+T` or `F6` | Format the selected table, or the table under the cursor |
| `Ctrl+G` | Show help; press any key to return |
| Arrow keys | Move the cursor |
| `Page Up` / `Page Down` | Move by one page |
| `Home` / `End` | Move to line start/end |
| `Ctrl+Home` / `Ctrl+End` | Move to document start/end |
| `Tab` | Insert four spaces |

To format a table, place the cursor on any of its lines and press `F6` or `Alt+T`. To select it explicitly, press `Ctrl+^` (often entered as `Ctrl+6`, depending on the keyboard layout) or `Alt+A`, move to the last table row, then press `F6` or `Alt+T`. The formatter aligns columns using Unicode display widths, preserves alignment markers, and understands escaped pipes and pipes inside inline code.

`F6` is provided because some terminals or keyboard layouts do not transmit `Alt+T` consistently. Windows Terminal or another modern terminal is recommended on Windows for the best color and keyboard support.

## Development

```sh
go test ./...
go build ./cmd/mdv ./cmd/mde
```

Project layout:

```text
cmd/mdv/             Viewer CLI
cmd/mde/             Editor CLI
internal/viewer/     Glamour renderer configuration
internal/editor/     Editor buffer, highlighting, and TUI
internal/tablefmt/   Shared Markdown table formatter
.github/workflows/   CI and release automation
Makefile             Build, test, and packaging targets
```

Key dependencies are Glamour for Markdown parsing/rendering, tcell for portable terminal input/output, and go-runewidth for Unicode-aware editing and table layout.

### GitHub Actions

The CI workflow runs on pushes, pull requests, and manual dispatches. It checks formatting, runs standard and race-enabled tests, vets the code, and builds both utilities for every supported operating system and architecture. Cross-build binaries are retained as workflow artifacts for seven days.

Once a GitHub remote is configured, pushing a version tag triggers the release workflow:

```sh
git tag v0.1.0
git push origin v0.1.0
```

The workflow verifies the tag, runs the checks, builds all six release archives and their SHA-256 checksums, and creates a GitHub release with generated notes. Tag names must begin with `v`; the leading `v` is omitted from archive filenames.

## License

This project is released under [The Unlicense](LICENSE), dedicating it to the public domain. See the license file for the full text.
