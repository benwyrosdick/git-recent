# git-recent

A simple CLI tool to quickly checkout recently active git branches using an interactive menu.

## Installation

```bash
go build -o git-recent
```

Move the binary to somewhere in your PATH:

```bash
mv git-recent /usr/local/bin/
```

Or use `go install`:

```bash
go install
```

## Usage

### Checkout local branches

```bash
git-recent
```

Shows a list of your local branches sorted by most recent commit activity.

### Checkout remote branches

```bash
git-recent -r
# or
git-recent --remote
```

Shows a list of remote branches. If a local tracking branch already exists, it will checkout that branch. Otherwise, it creates a new tracking branch.

## Controls

- `↑`/`k` - Move up
- `↓`/`j` - Move down
- `Enter` - Checkout selected branch
- `q`/`Esc`/`Ctrl+C` - Quit without checking out
