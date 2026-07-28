# Alacritty Colors

**An interactive Alacritty theme manager with comprehensive theme editing, font management, and seamless configuration capabilities.**

A command-line tool for managing Alacritty terminal themes, offering automatic downloads, custom theme generation, font configuration, and configuration management features.

![The editor: theme list, palette with live contrast ratios, and a preview of real terminal output](images/tui_1.png)

*Browsing the library. Every theme applies to the terminal as the cursor rests on it, and each palette row carries its WCAG contrast ratio against the background.*


## Features

- **326 themes out of the box** - 176 fetched automatically from the official Alacritty repository, plus 150 exclusive to this tool.
- **150 exclusive themes** - Designed here and found nowhere else: fifty grouped by mood, a hundred built on named accounts of colour — Goethe, Itten, Hering, Kobayashi, Birren, scotopic vision, circadian light, Munsell. Every one contrast-checked. See [the collection](#the-curated-collection).
- **Interactive TUI** - A terminal user interface for theme and font management.
- **Theme Management** - Apply, preview, and switch between available themes.
- **Theme Editing** - Edit theme colors directly within the TUI, with options for brightness and hue adjustments.
- **Theme Generation** - Create custom themes with color science-based harmonies (complementary, analogous, triadic, split-complementary, tetradic, monochromatic).
- **Random Theme Generator** - Generate random themes with WCAG-compliant contrast ratios.
- **Font Management** - Family, style and size, previewed live. Only fonts the terminal can actually resolve are offered, and sixteen popular monospace families can be downloaded and installed from inside the editor.
- **Auto-Download** - The 176 official themes are fetched and installed on first run.
- **Search & Favorites** - Find themes by name and mark favorites for quick access.
- **Backup & Restore** - Configuration management with automatic backups.
- **Cross-Platform** - Compatible with macOS, Linux, and Windows.
- **Efficient & Safe** - Designed to preserve your personal Alacritty settings.


## Quick Start

### Installation

**Using Go (recommended)**

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

Simply run `alacritty-colors` to launch the interactive TUI:

```bash
alacritty-colors
```

### Command Line Options

```bash
alacritty-colors [options]

Options:
  -config string    Path to alacritty.toml config file
  -themes string    Path to themes directory
  -export-collection dir
                    Write the 150 curated themes to a directory and exit
  -version          Show version information
  -help             Show help message

Examples:
  alacritty-colors                                    # Use default paths
  alacritty-colors -config /path/to/alacritty.toml   # Custom config
  alacritty-colors -themes /path/to/themes           # Custom themes dir
```

### Interactive TUI Keybindings

**Navigation**
- **Tab / Shift+Tab**: Cycle between the theme list, the palette and the preview
- **↑↓** or **j/k**: Move within a panel
- **PgUp / PgDn**: Jump ten rows — **Home / End**: first / last row
- **/**: Search themes as you type — fuzzy, so `ctpmoc` finds `catppuccin-mocha`
- **Esc**: Clear the search

**Theme management**
- **a**: Apply what you see immediately (browsing already applies as you go, and quitting keeps the one you landed on)
- **\***: Toggle favorite — **F**: show favorites only
- **e**: Fork the theme into an editable copy
- **d**: Delete a theme you created (downloaded themes are protected)
- **n**: Open the theme creator — **g**: generate a harmonious theme instantly
- **r**: Revert the theme to its saved version

**Color editing**
- **←→**: Brighten / darken the selected color
- **Shift+←→**: Rotate its hue
- **-** / **+**: Saturation — **[** / **]**: lightness
- **Enter** or **#**: Type an exact hex value
- **u**: Undo the selected color
- **s**: Save — **S**: save as…
- **c**: Cycle the preview color mode

**Panels & settings**
- **f**: Font settings — **p**: parameters (backup, reset, redownload)
- **?**: Full keyboard reference — **q**: quit

**Inside the font browser**
- **Tab**: Cycle family → style → size
- **↑↓**: Move within a column — **←→**: nudge the size from anywhere
- **d**: Download and install a monospace family (see below)
- **Esc**: Back to the editor

### Fonts

The browser lists families and styles taken from the same font database the terminal itself will use — Core Text on macOS, fontconfig elsewhere. This matters more than it sounds: `fc-list` is not what Alacritty consults on a Mac, and offering a name it reports but Core Text does not recognise is precisely what produces *unable to load font*. Nothing is listed that cannot be loaded.

Family, style and size all apply live, and writes are debounced, so scrolling a long list of families costs one write rather than one per row.

Press **d** for sixteen well-known monospace families — JetBrains Mono, Fira Code, Hack, Meslo, Iosevka, Victor Mono and others. They are fetched from the [Nerd Fonts](https://github.com/ryanoasis/nerd-fonts) release archive, so the powerline and icon glyphs a prompt expects are already patched in. Everything lands in your user font directory (`~/Library/Fonts`, or `~/.local/share/fonts` on Linux) — no administrator rights, nothing installed system-wide, and the browser reopens on the new list.

![The font browser: families, styles, sizes and a live sample](images/tui_2.png)

*The font browser. Families and styles come from the same database the terminal resolves against, so nothing listed can fail to load.*


### How editing treats your files

Edits are pushed to `themes/current.toml` only, so Alacritty repaints instantly and you judge colors at their true values in your own terminal. The theme file they came from is never touched until you save, and downloaded themes are never overwritten at all — saving one writes a copy under a new name. Writes are debounced, so holding an arrow key or scrolling through a long theme list does not hammer the disk.

Quitting commits: whatever is on screen stays applied, and any unsaved edits are written out on the way out rather than discarded. **a** does the same on demand, without the debounce.

Every palette row shows its WCAG contrast ratio against the theme background, and the preview grades the weakest color in the palette, which makes an unreadable theme obvious before you commit to it.

## The curated collection

**150 themes exclusive to this tool** ship with it and install on first run, alongside the 176 fetched from the official repository — press **p** → *Install the curated collection* to add or refresh them at any time. Each is built from a chosen background, text colour and accent; the sixteen ANSI slots are derived from that identity, so red stays red and blue stays distinct from cyan.

The first fifty are grouped by mood — Focus, Calm, Warmth, Energy, Imagination, Neutral, Contrast. The hundred that follow are grouped by the account of colour their structure comes from: Goethe, Itten, Hering, Kobayashi, Valdez & Mehrabian, Birren, scotopic vision, circadian light, Luscher, and the colour order systems of Munsell and Chevreul — ten palettes each. Naming a theory says where a palette's logic comes from — it is not an endorsement. Goethe was wrong about Newton and Luscher's test does not measure what it claimed to; what each left behind is a coherent way of arranging colour, which is what a sixteen-slot palette needs.

**On "colour psychology":** colour–mood links are mostly cultural convention, and the literature is thin. Two effects are better supported and are the ones leaned on here: short-wavelength light suppresses melatonin, which is why the evening palettes pull blue out; and strong red raises arousal, which is why it is an accent and never a field. The rest is stated design intent, not a claim about your brain.

What is *not* a matter of taste: every palette is contrast-checked by the test suite — AAA for body text, AA for every ANSI colour against its background.

### Download the collection separately

The 150 `.toml` files are in **[`themes/collection/`](https://github.com/vitruves/alacritty-colors/tree/main/themes/collection)** — browse them there, or take the lot without installing anything:

```bash
# The whole collection, straight into Alacritty
curl -L https://github.com/vitruves/alacritty-colors/archive/refs/heads/main.tar.gz \
  | tar -xz --strip-components=3 -C ~/.config/alacritty/themes \
        alacritty-colors-main/themes/collection

# Or a single theme
curl -O --output-dir ~/.config/alacritty/themes \
  https://raw.githubusercontent.com/vitruves/alacritty-colors/main/themes/collection/purkinje-shift.toml
```

If you already have the binary, it will write them anywhere you like without going near your Alacritty config:

```bash
alacritty-colors -export-collection ./somewhere
```

Each file carries a `# alacritty-colors collection ·` header. That marker is how the tool knows a file is its own: it will refresh those on reinstall, and never overwrite one you have edited.

Each `.toml` names its family and its intent in the header comment, so the file explains itself.

## How it works

The tool writes an import into your main config and then only ever touches the file it points at:

```toml
[general]
import = ["themes/current.toml"]
```

Applying a theme copies it over `themes/current.toml`; Alacritty reloads and repaints. Your own settings are never modified, the whole thing is disabled by removing that one line, and the import goes under `[general]` because Alacritty moved it there in 0.13 — a root-level `import` is ignored by every release since.

Config lives in `~/.config/alacritty/` (`%APPDATA%/alacritty/` on Windows):

```
alacritty.toml           # yours, plus the import line
themes/current.toml      # what the terminal is showing right now
themes/*.toml            # downloaded, curated and generated themes
backups/                 # timestamped copies of alacritty.toml
alacritty-colors.json    # this tool's own settings
```

Both the theme manager and the font browser edit `alacritty.toml`, so every write goes through one lock and lands atomically — a temporary file renamed over the target. A reader can never catch it half-written.

## Theme Generation

Press **n** for the creator, or **g** for an instant random palette.

Both build from a harmony rule — complementary, analogous, triadic, split-complementary, tetradic or monochromatic — in dark or light, and both hold WCAG contrast: at least 4.5:1 for every ANSI colour, targeting 7:1 for body text.

A terminal palette is semantic before it is decorative, so each ANSI slot starts at its canonical hue and is only pulled *towards* the harmony, never past it. Random means wide, not wrong: `red` never comes out green.

## Troubleshooting

**A theme does not apply.** Check that `alacritty.toml` has the import under `[general]` — Alacritty moved it there in 0.13 and a root-level `import` is silently ignored:

```toml
[general]
import = ["themes/current.toml"]
```

**Colours on screen do not match the editor.** Two usual causes, neither of them the theme. `window.opacity` below 1.0 blends the background with whatever is behind the window, so it can never match the swatch — the editor says so in the preview when it detects this. And many programs (Claude Code, `bat`, `delta`, fish with a frozen theme) emit 24-bit colour of their own and bypass the palette entirely.

**The font list is empty, or a font will not load.** Only families the terminal can actually resolve are listed — Core Text on macOS, fontconfig on Linux. Font browsing is not implemented on Windows yet; set `font.normal.family` by hand there.

**Something went wrong with the config.** Every write is atomic and timestamped backups live in `~/.config/alacritty/backups/`.

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
