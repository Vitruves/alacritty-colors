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


<img width="918" alt="image" src="https://github.com/user-attachments/assets/e8a82abd-bf34-408e-acf0-955d72a61fb1" />


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

## Usage

Simply run `alacritty-colors` to launch the interactive theme editor:

```bash
alacritty-colors
```

### Interactive TUI Keybindings

The interactive terminal interface (TUI) provides the following keybindings for navigation and theme management:

- **Tab**: Switch between panels
- **↑↓**: Navigate through lists
- **←→**: Adjust brightness of selected color
- **Shift+←→**: Adjust hue of selected color
- **Enter/a**: Apply the selected theme
- **e**: Edit a copy of the current theme
- **d**: Delete the current theme
- **c**: Cycle through color modes
- **q**: Quit the application
- **s**: Save the current theme
- **r**: Reset changes to the current theme

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
The application automatically handles paths.

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

The interactive TUI provides a real-time preview of the selected theme as you navigate and make changes.

## Contributing

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
- Reinitialize configuration (re-run `alacritty-colors` and ensure themes are downloaded).
- Check current theme status within the TUI.
- Verify import line exists in config (manual check of `~/.config/alacritty/alacritty.toml`).

**Permission errors:**
- Fix directory permissions:
  ```bash
  chmod 755 ~/.config/alacritty/
  chmod 644 ~/.config/alacritty/alacritty.toml
  ```
- Ensure themes directory is writable:
  ```bash
  chmod 755 ~/.config/alacritty/themes/
  ```

**Missing themes after update:**
- The application automatically handles theme updates during startup.
- Check themes directory: `ls -la ~/.config/alacritty/themes/`

**Configuration backup and restore:**
- Backups are automatically created by the application. You can manually manage them in `~/.config/alacritty/backups/`.
- To restore, you would typically replace your `alacritty.toml` with a backup.

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