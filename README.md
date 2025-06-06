# Alacritty Colors

**Advanced Alacritty theme manager with generation, preview, and seamless management capabilities**

A powerful command-line tool for managing Alacritty terminal themes with automatic downloads, custom theme generation, and safe configuration management.

![image-10](https://github.com/user-attachments/assets/40163b06-ba60-4bb0-a961-074a55d32d3f)


## Features

- **Theme Management** - Apply, preview, and switch between hundreds of themes
- **Random Themes** - Instantly apply random themes for variety
- **Theme Generation** - Create custom themes with 12+ color schemes
- **Auto-Download** - Automatically download official Alacritty themes
- **Search & Preview** - Find themes by name and preview before applying
- **Backup & Restore** - Safe configuration management with automatic backups
- **Cross-Platform** - Works on macOS, Linux, and Windows
- **Fast & Safe** - Preserves your personal Alacritty settings
- **User Friendly interactive terminal interface** - Theme creation made easy!


<img width="918" alt="image" src="https://github.com/user-attachments/assets/e8a82abd-bf34-408e-acf0-955e72a61fb1" />


## Quick Start

### Installation


**Using Go (recommanded) **
```bash
go install github.com/vitruves/alacritty-colors/cmd/alacritty-colors@latest
```

**From Source:**
```bash
git clone https://github.com/vitruves/alacritty-colors.git
cd alacritty-colors
make build && make install
```

### First Time Setup

```bash
# Initialize configuration and download themes
alacritty-colors init

# List available themes
alacritty-colors list

# Apply your first theme
alacritty-colors apply dracula
```

## Usage

### Basic Commands

```bash
# Theme Management
alacritty-colors list                    # List all themes
alacritty-colors apply <theme>           # Apply a specific theme
alacritty-colors random                  # Apply random theme
alacritty-colors current                 # Show current theme

# Search and Preview
alacritty-colors search nord             # Search themes
alacritty-colors preview dracula         # Preview theme colors

# Theme Generation
alacritty-colors generate --scheme random
alacritty-colors generate --scheme pastel --name "my-theme"

# Interactive theme editor
alacritty-colors interactive

# Backup Management
alacritty-colors backup                  # Create backup
alacritty-colors restore                 # Restore from backup

# Updates
alacritty-colors update                  # Update theme database
```

### Theme Generation Schemes

Generate custom themes with various color palettes:

| Scheme      | Description                | Best For                        |
| ----------- | -------------------------- | ------------------------------- |
| `random`    | Completely random colors   | Experimentation                 |
| `pastel`    | Soft, muted tones          | Extended coding sessions        |
| `neon`      | Bright, vibrant colors     | High contrast, retro aesthetics |
| `mono`      | Monochromatic grayscale    | Minimalist setups               |
| `warm`      | Reds, oranges, yellows     | Cozy, comfortable environments  |
| `cool`      | Blues, greens, purples     | Clean, professional look        |
| `nature`    | Earth tones, forest colors | Natural, organic feel           |
| `cyberpunk` | Neon greens, magentas      | Futuristic, hacker aesthetic    |
| `dracula`   | Dracula-inspired variants  | Popular dark theme variations   |
| `nord`      | Nord-inspired cool tones   | Scandinavian minimalism         |
| `solarized` | Solarized variations       | Scientific color precision      |
| `gruvbox`   | Gruvbox retro variants     | Warm retro computing feel       |

### Command Examples

```bash
# Apply popular themes
alacritty-colors apply dracula
alacritty-colors apply nord
alacritty-colors apply gruvbox-dark

# Generate custom themes
alacritty-colors generate --scheme cyberpunk --name "my-cyberpunk"
alacritty-colors generate --scheme warm --name "sunset-terminal"

# Search for specific themes
alacritty-colors search dark
alacritty-colors search solarized

# Preview before applying
alacritty-colors preview nord
# Shows color palette and prompts to apply

# Different output formats
alacritty-colors list --format grid
alacritty-colors list --format json
alacritty-colors list --format list
```

## Configuration

Alacritty Colors automatically detects your configuration location:

**Default Locations:**
- **macOS/Linux**: `~/.config/alacritty/`
- **Windows**: `%APPDATA%/alacritty/`

**Directory Structure:**
```
~/.config/alacritty/
├── alacritty.toml           # Main config (with import line)
├── themes/
│   ├── current.toml         # Currently applied theme
│   ├── dracula.toml         # Downloaded/generated themes
│   ├── nord.toml
│   └── my-custom.toml
├── backups/                 # Automatic backups
│   ├── alacritty_2024-01-15_10-30-45.toml
│   └── alacritty_2024-01-16_14-22-10.toml
└── alacritty-colors.json    # Tool configuration
```

**Custom Paths:**
```bash
alacritty-colors --config /path/to/alacritty.toml \
                 --themes-dir /path/to/themes \
                 apply dracula
```

## How It Works

Alacritty Colors uses a simple and safe approach:

1. **Import Line**: Adds `import = ["themes/current.toml"]` to your main config
2. **Theme Directory**: Stores all themes in a `themes/` subdirectory  
3. **Current Theme**: Copies selected theme to `themes/current.toml`
4. **Preservation**: Your personal settings remain untouched in the main config

**Benefits:**
- Your personal settings are never modified
- Easy to disable (just remove the import line)
- Themes are portable and shareable
- No complex configuration file parsing required
- Clean separation between themes and personal config

## Theme Preview

When previewing themes, you'll see:
- **Color Palette**: Visual representation of all theme colors
- **Theme Information**: Author, description, and metadata when available
- **Interactive Apply**: Option to apply the theme immediately

Example preview output:
```bash
$ alacritty-colors preview nord

Theme Preview: Nord
Description: An arctic, north-bluish clean and elegant color palette
Author: Arctic Ice Studio

Primary Colors:
  background    #2e3440
  foreground    #d8dee9

Normal Colors:
  black         #3b4252
  red           #bf616a
  green         #a3be8c
  yellow        #ebcb8b
  blue          #81a1c1
  magenta       #b48ead
  cyan          #88c0d0
  white         #e5e9f0

Bright Colors:
  black         #4c566a
  red           #bf616a
  green         #a3be8c
  ...

Apply this theme? [y/N]:
```

## Advanced Usage

### Batch Operations

```bash
# Apply random themes from a specific set
alacritty-colors search dark | grep -v "^Search" | head -5 | while read theme; do
    echo "Trying: $theme"
    alacritty-colors apply "$theme"
    sleep 2
done

# Generate multiple themed variants
for scheme in warm cool nature cyberpunk; do
    alacritty-colors generate --scheme "$scheme" --name "auto-$scheme"
done
```

### Configuration Management

```bash
# Create named backup
alacritty-colors backup
cp ~/.config/alacritty/backups/alacritty_*.toml my-config-backup.toml

# Restore specific configuration
alacritty-colors restore my-config-backup.toml

# Reset to default theme
echo '# Default theme' > ~/.config/alacritty/themes/current.toml
```

### Integration with Other Tools

```bash
# Use with system theme switching
if [ "$(uname)" = "Darwin" ]; then
    if defaults read -g AppleInterfaceStyle 2>/dev/null | grep -q Dark; then
        alacritty-colors apply dracula
    else
        alacritty-colors apply solarized-light
    fi
fi

# Random theme on new shell
if [ "$RANDOM_THEME" = "1" ]; then
    alacritty-colors random
fi
```

### Contributing

Contributions are welcome! Here's how to get started:

1. **Fork** the repository on GitHub
2. **Clone** your fork locally
3. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
4. **Make** your changes with tests
5. **Commit** your changes (`git commit -m 'Add amazing feature'`)
6. **Push** to your branch (`git push origin feature/amazing-feature`)
7. **Open** a Pull Request

## Troubleshooting

### Common Issues

**Themes not applying properly:**
```bash
# Reinitialize configuration
alacritty-colors init

# Check current theme status
alacritty-colors current

# Verify import line exists in config
grep -n "import.*themes/current.toml" ~/.config/alacritty/alacritty.toml
```

**Permission errors:**
```bash
# Fix directory permissions
chmod 755 ~/.config/alacritty/
chmod 644 ~/.config/alacritty/alacritty.toml

# Ensure themes directory is writable
chmod 755 ~/.config/alacritty/themes/
```

**Missing themes after update:**
```bash
# Re-download theme database
alacritty-colors update

# Check themes directory
ls -la ~/.config/alacritty/themes/
```

**Configuration backup and restore:**
```bash
# List available backups
alacritty-colors restore

# Restore from specific backup
alacritty-colors restore alacritty_2024-01-15_10-30-45.toml

# Manual backup
cp ~/.config/alacritty/alacritty.toml ~/alacritty-backup.toml
```

### Getting Help

If you encounter issues:

1. Check this troubleshooting section
2. Look at existing [GitHub Issues](https://github.com/vitruves/alacritty-colors/issues)
3. Create a new issue with:
   - Your operating system and version
   - Alacritty version (`alacritty --version`)
   - Steps to reproduce the problem
   - Expected vs actual behavior

## Performance

Alacritty Colors is designed for speed:

- **Fast theme switching**: Themes apply instantly by copying files
- **Minimal overhead**: No complex parsing or processing during application
- **Efficient downloads**: Themes are downloaded once and cached locally
- **Small footprint**: Written in Go for fast startup and low memory usage

Typical performance on modern systems:
- Theme application: < 100ms
- Theme listing: < 50ms
- Theme generation: < 200ms
- Initial setup: < 5 seconds (including downloads)

## License

MIT License - see [LICENSE](LICENSE) file for complete details.

## Acknowledgments

This project builds upon the excellent work of:

- **[Alacritty](https://github.com/alacritty/alacritty)** - The fast, cross-platform, OpenGL terminal emulator
- **[Alacritty Themes](https://github.com/alacritty/alacritty-theme)** - Official community theme collection
- **Color Theory Research** - HSL color space implementation for harmonious theme generation
- **Go Community** - For excellent tooling and libraries

## Related Projects

- **[Alacritty](https://github.com/alacritty/alacritty)** - The terminal emulator itself
- **[Alacritty Themes](https://github.com/alacritty/alacritty-theme)** - Official theme repository
- **[Base16](https://github.com/chriskempson/base16)** - Color scheme framework
- **[Pywal](https://github.com/dylanaraps/pywal)** - System-wide color scheme generation

---

**Made for terminal enthusiasts.**
