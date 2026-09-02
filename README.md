<img width="1367" height="915" alt="image" src="https://github.com/user-attachments/assets/cb7952b7-7e90-45fb-b770-bb7282574bd1" />

# differ

A native macOS and terminal UI for browsing git history and working-tree changes. Navigate commits, files, and diffs from the keyboard.

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

This opens a native macOS window. To use the terminal UI instead:

```sh
differ --tui
```

## Key bindings

Both the GUI and the terminal UI use the same keys:

| Key | Action |
|-----|--------|
| `c` | Switch between Changes and History tabs |
| `h` / `←` / `shift+tab` | Focus previous pane |
| `l` / `→` / `tab` | Focus next pane |
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` | Go to top |
| `G` | Go to bottom |
| `q` / `ctrl+c` | Quit (terminal UI) |
| `q` / `Cmd+Q` / `Cmd+W` | Quit (GUI) |

## Development

If you edit any GUI source under `internal/gui/frontend/src`, rebuild the embedded frontend bundle before building the Go binary:

```sh
cd internal/gui/frontend && npm install && npm run build
```

This regenerates `internal/gui/frontend/dist/`, which is committed to git and embedded via `go:embed` — commit the updated `dist/` alongside your Go changes. Note that `go build`/`go run` also need `-tags production` for a working GUI locally (not just `go install` — see the Install section above).

## Requirements

- Go 1.21+
- `git` must be on your `$PATH`
- Xcode Command Line Tools (required for the GUI's cgo/WebKit bindings)

## License

MIT
