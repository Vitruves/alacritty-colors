# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Alacritty Colors is a command-line tool for managing Alacritty terminal themes, written in Go. It provides theme generation, preview, and seamless management capabilities.

## Build and Development Commands

```bash
# Build the binary
make build                    # Builds to build/alacritty-colors

# Install
make install                  # System-wide install to /usr/local/bin (requires sudo)
make local-install           # User install to ~/.local/bin

# Development
make run                     # Build and run immediately
make test                    # Run Go tests (currently no tests exist)
make deps                    # Tidy Go module dependencies
make clean                   # Clean build artifacts
```

## Architecture and Code Structure

### Core Architecture
The application follows a command-line tool pattern using Cobra for CLI framework with subcommands for different operations (apply, generate, list, etc.).

### Package Structure
- **Entry Point**: `cmd/alacritty-colors/main.go` - Initializes Cobra commands
- **Internal Packages** (private to this module):
  - `internal/config/` - Manages tool configuration and Alacritty config file operations
  - `internal/downloader/` - Downloads themes from the official Alacritty themes repository
  - `internal/theme/` - Core theme logic including generation, parsing, and management
  - `internal/tui/` - Terminal UI for interactive theme creation using tview
  - `internal/ui/` - UI utilities for formatting and color display
- **Public Package**: `pkg/alacritty/` - Alacritty configuration parsing utilities

### Key Implementation Details

1. **Theme Management Strategy**: 
   - Themes are stored in `~/.config/alacritty/themes/`
   - Current theme is always copied to `themes/current.toml`
   - Main config imports this file via `import = ["themes/current.toml"]`
   - This preserves user settings while allowing easy theme switching

2. **Theme Generation**: 
   - Located in `internal/theme/generator.go`
   - Supports 12+ color schemes (pastel, neon, cyberpunk, etc.)
   - Uses HSL color space for harmonious color generation

3. **Configuration Safety**:
   - Automatic backups before any config modifications
   - Non-destructive approach - only adds import line to main config
   - All theme operations work through file copying, not parsing/modifying

4. **Interactive TUI**:
   - Built with rivo/tview library
   - Allows real-time color picking and theme creation
   - Located in `internal/tui/`

## Dependencies

Key Go dependencies:
- `github.com/spf13/cobra` - CLI framework
- `github.com/rivo/tview` - Terminal UI
- `github.com/gdamore/tcell/v2` - Terminal handling
- `github.com/fatih/color` - Colored terminal output

## Development Guidelines

1. **No Tests Currently**: The project has no test files despite having a test target
2. **File Operations**: All config modifications should create backups first
3. **Theme Format**: Themes use TOML format compatible with Alacritty
4. **Cross-Platform**: Code should work on macOS, Linux, and Windows