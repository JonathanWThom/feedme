# hn

A terminal UI for [Hacker News](https://news.ycombinator.com).

## Installation

### With Go

```bash
go install github.com/JonathanWThom/hn@latest
```

### From Source

```bash
git clone https://github.com/JonathanWThom/hn.git
cd hn
make install
```

## Usage

```bash
hn
```

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Open link in browser |
| `c` | View comments |
| `b` / `Esc` | Back to stories |
| `Tab` / `l` | Next feed |
| `Shift+Tab` / `h` | Previous feed |
| `r` | Refresh |
| `?` | Toggle help |
| `q` | Quit |

## Feeds

- **Top** - Top stories
- **New** - Newest stories
- **Best** - Best stories
- **Ask** - Ask HN posts
- **Show** - Show HN posts

## License

MIT
