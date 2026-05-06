# Topai

AI-powered terminal process monitor built in Go with a beautiful TUI interface.

![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)

## Features

- **Real-time Process Monitoring** - Watch CPU and memory usage with live updates
- **AI-Powered Analysis** - AI providers analyze high-CPU processes to determine if they're stuck or legitimately busy
- **Interactive TUI** - Beautiful terminal interface built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Multiple AI Providers** - Support for Anthropic Claude, OpenAI GPT-4, and Google Gemini
- **Process Management** - Kill processes directly from the interface
- **Smart Sorting** - Sort by CPU or memory usage

## Installation

```bash
go install github.com/Brownei/topai@latest
```

Or clone and build manually:

```bash
git clone https://github.com/Brownei/topai.git
cd topai
go build -o topai
./topai
```

## Usage

Simply run:

```bash
topai
```

On first run, you'll be prompted to select an AI provider and enter your API key.

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate processes |
| `c` | Sort by CPU usage |
| `m` | Sort by memory usage |
| `K` | Kill selected process |
| `t` | Open chat interface |
| `q` or `Ctrl+C` | Quit |

## AI Providers

topai supports multiple AI providers for process analysis:

| Provider | Model | Best For |
|----------|-------|----------|
| 🔮 Anthropic Claude | Claude Haiku 4.5 | Fast, efficient analysis (Recommended) |
| 🤖 OpenAI GPT-4 | GPT-4 | Powerful general-purpose analysis |
| ✨ Google Gemini | Gemini 1.5 Flash | Google's latest with fast responses |

## Configuration

Configuration is stored at `~/.config/topai/config.json`:

```json
{
  "provider": "anthropic",
  "api_key": "your-api-key-here"
}
```

## Architecture

```
topai/
├── main.go                 # Entry point
├── providers/
│   ├── provider.go         # Provider management & configuration
│   ├── selection.go        # Interactive provider selection UI
│   ├── anthropic.go        # Anthropic Claude implementation
│   ├── openai.go           # OpenAI GPT-4 implementation
│   └── gemini.go           # Google Gemini implementation
├── process/
│   ├── process.go          # Process monitoring logic
│   ├── tui.go              # Main TUI implementation
│   └── system-stats.go     # System statistics collection
└── utils/
    └── utils.go            # AI prompt generation utilities
```

## Technologies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - Pre-built TUI components
- [gopsutil](https://github.com/shirou/gopsutil) - System and process monitoring
- [Anthropic SDK](https://github.com/anthropics/anthropic-sdk-go) - Claude integration
- [OpenAI SDK](https://github.com/sashabaranov/go-openai) - GPT-4 integration
- [Google GenAI](https://google.golang.org/genai) - Gemini integration

