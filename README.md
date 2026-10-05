<img width="1367" height="915" alt="image" src="https://github.com/user-attachments/assets/cb7952b7-7e90-45fb-b770-bb7282574bd1" />

# differ

A native macOS and terminal UI for browsing git history and working-tree changes. Navigate commits, files, and diffs from the keyboard. The macOS app keeps every repo you open in one window, one tab each.

## Install

Requires Go 1.21+.

```sh
go install -tags production github.com/dannyeldridge/differ@latest
```

The `-tags production` flag is required for the GUI to build correctly.

The binary is placed in `$GOPATH/bin` (usually `~/go/bin`). Make sure that's on your `$PATH`.

## Usage

Run `differ` from inside any git repository:

```sh
cd your-repo
differ
```

This opens a native macOS window and returns your shell. Running `differ` in another repo adds a tab to that same window, or focuses the tab if that repo is already open, instead of opening a second window. The tab bar appears once more than one repo is open. Click a tab to switch, or its `×` to close it.

The GUI runs detached from the terminal, so closing the terminal doesn't close the app. Its output goes to `~/Library/Caches/differ/differ.log`. Pass `--foreground` to keep it attached to your terminal instead, which is useful for debugging.

To use the terminal UI instead:

```sh
differ --tui
```

## Key bindings

The GUI and the terminal UI share these keys. The tab keys at the bottom are GUI-only.

| Key | Action |
|-----|--------|
| `c` | Switch between the Changes and History views |
| `h` / `←` / `shift+tab` | Focus previous pane |
| `l` / `→` / `tab` | Focus next pane |
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` | Go to top |
| `G` | Go to bottom |
| `q` / `ctrl+c` | Quit (terminal UI) |
| `q` / `Cmd+W` | Close the current tab; quits when it is the last one (GUI) |
| `Cmd+Q` | Quit the app and all tabs (GUI) |
| `Cmd+1`–`Cmd+9` | Jump to tab by position (GUI) |

## Development

If you edit any GUI source under `internal/gui/frontend/src`, rebuild the embedded frontend bundle before building the Go binary:

```sh
cd internal/gui/frontend && npm install && npm run build
```

This regenerates `internal/gui/frontend/dist/`, which is committed to git and embedded via `go:embed` — commit the updated `dist/` alongside your Go changes. Note that `go build`/`go run` also need `-tags production` for a working GUI locally (not just `go install` — see the Install section above).

Because the GUI detaches by default, use `--foreground` when running from source so you can see its output and stop it with Ctrl+C:

```sh
go run -tags production . --foreground
```

The Dock and Cmd+Tab icon is `internal/gui/icon.png` (1024×1024), embedded in the binary and applied at startup. Replace the file and rebuild to change it. It only applies while the app is running, since `differ` is a bare binary and not an `.app` bundle.

## Requirements

- Go 1.21+
- `git` must be on your `$PATH`
- Xcode Command Line Tools (required for the GUI's cgo/WebKit bindings)

## License

MIT
