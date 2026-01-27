# hn

A terminal UI for [Hacker News](https://news.ycombinator.com).

## Installation

### Quick Install (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/JonathanWThom/hn/main/install.sh | bash
```

### Download Binary

Download the latest release for your platform from the [releases page](https://github.com/JonathanWThom/hn/releases).

**macOS:**
```bash
# Apple Silicon
curl -L https://github.com/JonathanWThom/hn/releases/latest/download/hn_darwin_arm64.tar.gz | tar xz
mv hn ~/.local/bin/

# Intel
curl -L https://github.com/JonathanWThom/hn/releases/latest/download/hn_darwin_amd64.tar.gz | tar xz
mv hn ~/.local/bin/
```

**Linux:**
```bash
# amd64
curl -L https://github.com/JonathanWThom/hn/releases/latest/download/hn_linux_amd64.tar.gz | tar xz
mv hn ~/.local/bin/

# arm64
curl -L https://github.com/JonathanWThom/hn/releases/latest/download/hn_linux_arm64.tar.gz | tar xz
mv hn ~/.local/bin/
```

**Windows (PowerShell):**
```powershell
# Download and extract
Invoke-WebRequest -Uri "https://github.com/JonathanWThom/hn/releases/latest/download/hn_windows_amd64.zip" -OutFile "$env:TEMP\hn.zip"
Expand-Archive -Path "$env:TEMP\hn.zip" -DestinationPath "$env:TEMP\hn" -Force

# Move to a directory in your PATH (e.g., create ~/bin)
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin"
Move-Item -Path "$env:TEMP\hn\hn.exe" -Destination "$env:USERPROFILE\bin\hn.exe" -Force

# Add to PATH (run once)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\bin", "User")
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
mv hn ~/.local/bin/
```

> **Note:** Ensure `~/.local/bin` is in your PATH. Add `export PATH="$HOME/.local/bin:$PATH"` to your shell config if needed.

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
