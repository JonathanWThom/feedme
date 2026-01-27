# feedme

A terminal UI for browsing [Hacker News](https://news.ycombinator.com) and [Lobste.rs](https://lobste.rs).

## Installation

### Quick Install (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/JonathanWThom/feedme/main/install.sh | bash
```

### Download Binary

Download the latest release for your platform from the [releases page](https://github.com/JonathanWThom/feedme/releases).

**macOS:**
```bash
# Apple Silicon
curl -L https://github.com/JonathanWThom/feedme/releases/latest/download/fm_darwin_arm64.tar.gz | tar xz
mv fm ~/.local/bin/

# Intel
curl -L https://github.com/JonathanWThom/feedme/releases/latest/download/fm_darwin_amd64.tar.gz | tar xz
mv fm ~/.local/bin/
```

**Linux:**
```bash
# amd64
curl -L https://github.com/JonathanWThom/feedme/releases/latest/download/fm_linux_amd64.tar.gz | tar xz
mv fm ~/.local/bin/

# arm64
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

## Usage

```bash
# Browse Hacker News (default)
fm

# Browse Lobste.rs
fm -s lobsters
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
| `m` | Toggle select mode (to copy text) |
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

## License

MIT
