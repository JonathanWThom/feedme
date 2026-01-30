# feedme

A read-only terminal UI for browsing [Hacker News](https://news.ycombinator.com), [Lobste.rs](https://lobste.rs), and [reddit](https://reddit.com).

## Install (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/JonathanWThom/feedme/main/install.sh | bash
```

[Other installation options](#other-installation-options) (including Windows) available below.

## Usage

```bash
# Browse Hacker News (default)
fm

# Browse Lobste.rs
fm -s lobsters

# Browse any subreddit
fm -s r/golang
fm -s r/bellingham
fm -s r/seinfeld
```

You can also switch sources from within the app by pressing `s`.

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` / `o` | Open link in browser |
| `c` | View comments |
| `b` / `Esc` | Back to stories |
| `Tab` / `l` | Next feed |
| `Shift+Tab` / `h` | Previous feed |
| `s` | Switch source (HN, Lobste.rs, Reddit) |
| `r` | Refresh |
| `v` | Visual mode (in comments) |
| `y` | Yank selection to clipboard |
| `m` | Toggle mouse (for terminal copy) |
| `?` | Toggle help |
| `q` | Quit |

## Sources

### Hacker News
- **Top** - Top stories
- **New** - Newest stories
- **Best** - Best stories
- **Ask** - Ask HN posts
- **Show** - Show HN posts

### Lobste.rs
- **Hot** - Hottest stories
- **New** - Newest stories
- **Recent** - Recently active

### Reddit (`-s r/subreddit`)
- **Hot** - Hot posts (default)
- **New** - Newest posts
- **Top** - Top posts
- **Rising** - Rising posts
- **Best** - Best posts

## Other Installation Options

### Download Binary

Download the latest release for your platform from the [releases page](https://github.com/JonathanWThom/feedme/releases).

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/JonathanWThom/feedme/releases/latest/download/fm_darwin_arm64.tar.gz | tar xz
mv fm ~/.local/bin/
```

**macOS (Intel):**
```bash
curl -L https://github.com/JonathanWThom/feedme/releases/latest/download/fm_darwin_amd64.tar.gz | tar xz
mv fm ~/.local/bin/
```

**Linux (amd64):**
```bash
curl -L https://github.com/JonathanWThom/feedme/releases/latest/download/fm_linux_amd64.tar.gz | tar xz
mv fm ~/.local/bin/
```

**Linux (arm64):**
```bash
curl -L https://github.com/JonathanWThom/feedme/releases/latest/download/fm_linux_arm64.tar.gz | tar xz
mv fm ~/.local/bin/
```

**Windows (PowerShell):**
```powershell
# Download and extract
Invoke-WebRequest -Uri "https://github.com/JonathanWThom/feedme/releases/latest/download/fm_windows_amd64.zip" -OutFile "$env:TEMP\fm.zip"
Expand-Archive -Path "$env:TEMP\fm.zip" -DestinationPath "$env:TEMP\fm" -Force

# Move to a directory in your PATH (e.g., create ~/bin)
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin"
Move-Item -Path "$env:TEMP\fm\fm.exe" -Destination "$env:USERPROFILE\bin\fm.exe" -Force

# Add to PATH (run once)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\bin", "User")
```

### With Go

```bash
go install github.com/JonathanWThom/feedme@latest
```

### From Source

```bash
git clone https://github.com/JonathanWThom/feedme.git
cd feedme
go build -o fm .
mv fm ~/.local/bin/
```

> **Note:** Ensure `~/.local/bin` is in your PATH. Add `export PATH="$HOME/.local/bin:$PATH"` to your shell config if needed.

## Disclaimer

This project was largely "vibe coded" with Claude Code my own use. Did I read
all of it? No. Does it work pretty well? Sure.

## License

MIT
