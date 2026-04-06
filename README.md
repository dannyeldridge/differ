# differ

A terminal UI for browsing git history. Navigate commits, files, and diffs from the keyboard.

## Install

Requires Go 1.21+.

```sh
go install github.com/dannyeldridge/differ@latest
```

The binary is placed in `$GOPATH/bin` (usually `~/go/bin`). Make sure that's on your `$PATH`.

## Usage

Run `differ` from inside any git repository:

```sh
cd your-repo
differ
```

## Key bindings

| Key | Action |
|-----|--------|
| `h` / `←` / `shift+tab` | Focus previous pane |
| `l` / `→` / `tab` | Focus next pane |
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` | Go to top |
| `G` | Go to bottom |
| `q` / `ctrl+c` | Quit |

## Requirements

- Go 1.21+
- `git` must be on your `$PATH`

## License

MIT
