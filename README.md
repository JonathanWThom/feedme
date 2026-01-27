# hn

A terminal UI for [Hacker News](https://news.ycombinator.com).

## Installation

### Download Binary (Recommended)

Download the latest release for your platform from the [releases page](https://github.com/JonathanWThom/hn/releases).

```bash
# macOS (Apple Silicon)
curl -L https://github.com/JonathanWThom/hn/releases/latest/download/hn_darwin_arm64.tar.gz | tar xz
sudo mv hn /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/JonathanWThom/hn/releases/latest/download/hn_darwin_amd64.tar.gz | tar xz
sudo mv hn /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/JonathanWThom/hn/releases/latest/download/hn_linux_amd64.tar.gz | tar xz
sudo mv hn /usr/local/bin/

# Linux (arm64)
curl -L https://github.com/JonathanWThom/hn/releases/latest/download/hn_linux_arm64.tar.gz | tar xz
sudo mv hn /usr/local/bin/
```

### With Go

```bash
go install github.com/JonathanWThom/hn@latest
```

### From Source

```bash
git clone https://github.com/JonathanWThom/hn.git
cd hn
go build -o hn .
sudo mv hn /usr/local/bin/
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
