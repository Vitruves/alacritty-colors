package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vitruves/alacritty-colors/internal/config"
	"github.com/vitruves/alacritty-colors/internal/theme"
	"github.com/vitruves/alacritty-colors/internal/tui"
	"github.com/vitruves/alacritty-colors/internal/ui"
)

const version = "1.0.0"

var (
	configFile string
	themesDir  string
	backupDir  string
	verbose    bool
)

func main() {
	// Custom help template with enhanced colors and structure
	cobra.AddTemplateFunc("colorize", func(s string) string {
		return ui.ColorizeHeader(s)
	})

	// Create custom help template with colors
	helpTemplate := `{{colorize "Alacritty Colors v1.0.0"}}
Advanced Alacritty theme manager with 500+ themes, smart font pairing, and visual effects.

{{colorize "USAGE"}}
  {{.UseLine}}

{{if .HasAvailableSubCommands}}{{colorize "COMMANDS"}}
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}  {{printf "\033[36m%-12s\033[0m" .Name}} {{.Short}}
{{end}}{{end}}{{end}}
{{if .HasAvailableLocalFlags}}{{colorize "OPTIONS"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}
{{if .HasAvailableInheritedFlags}}{{colorize "GLOBAL OPTIONS"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}
{{colorize "EXAMPLES"}}
  alacritty-colors init                # Initialize configuration
  alacritty-colors apply dracula       # Apply specific theme
  alacritty-colors random --dark       # Random dark theme
  alacritty-colors generate -s neon    # Generate neon theme

{{colorize "MORE INFO"}}
  Use "alacritty-colors [command] --help" for detailed information.
`

	var rootCmd = &cobra.Command{
		Use:     "alacritty-colors",
		Version: version,
	}

	// Set custom help template
	rootCmd.SetUsageTemplate(helpTemplate)

	// Global flags with better organization
	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&configFile, "config", "c", "", "Alacritty config file path")
	flags.StringVar(&themesDir, "themes-dir", "", "Custom themes directory")
	flags.StringVar(&backupDir, "backup-dir", "", "Custom backup directory")
	flags.BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Commands with improved structure
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(applyCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(randomCmd())
	rootCmd.AddCommand(generateCmd())
	rootCmd.AddCommand(searchCmd())
	rootCmd.AddCommand(previewCmd())
	rootCmd.AddCommand(slideshowCmd())
	rootCmd.AddCommand(interactiveCmd())
	rootCmd.AddCommand(backupCmd())
	rootCmd.AddCommand(restoreCmd())
	rootCmd.AddCommand(updateCmd())
	rootCmd.AddCommand(configCmd())

	if err := rootCmd.Execute(); err != nil {
		ui.PrintError("Error: %v", err)
		os.Exit(1)
	}
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration and download themes",
		Long: `Initialize alacritty-colors configuration:

• Create necessary directories (themes/, backups/)
• Download official Alacritty theme collection  
• Set up configuration file with import statements
• Verify Alacritty installation and config location

This command is safe to run multiple times and will not
overwrite existing configurations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				ui.PrintInfo("Initializing with verbose output enabled")
			}

			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			tm := theme.NewManager(cfg)
			tm.SetVerbose(verbose)
			return tm.Initialize()
		},
	}
}

func applyCmd() *cobra.Command {
	var (
		withFont   bool
		opacity    float64
		blur       float64
		fontSize   float64
		fontFamily string
	)

	cmd := &cobra.Command{
		Use:   "apply <theme-name>",
		Short: "Apply a specific theme",
		Long: `Apply a specific theme to your Alacritty configuration:

The theme will be safely applied using the import system, preserving
your existing configuration. Optionally modify font and visual effects.

Examples:

  alacritty-colors apply dracula
  alacritty-colors apply nord --font --font-size 16
  alacritty-colors apply gruvbox --opacity 0.9 --blur 10`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			tm.SetVerbose(verbose)

			opts := &theme.ApplyOptions{
				WithFont:   withFont,
				Opacity:    opacity,
				Blur:       blur,
				FontSize:   fontSize,
				FontFamily: fontFamily,
			}

			return tm.ApplyThemeWithOptions(args[0], opts)
		},
	}

	cmd.Flags().BoolVar(&withFont, "font", false, "Also change font to match theme")
	cmd.Flags().Float64Var(&opacity, "opacity", 0, "Set window opacity (0.0-1.0)")
	cmd.Flags().Float64Var(&blur, "blur", 0, "Set background blur radius")
	cmd.Flags().Float64Var(&fontSize, "font-size", 0, "Set font size")
	cmd.Flags().StringVar(&fontFamily, "font-family", "", "Set font family")

	return cmd
}

func listCmd() *cobra.Command {
	var (
		format     string
		showColors bool
		darkOnly   bool
		lightOnly  bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available themes",
		Long: `List all available themes in various formats:

Formats:
  • grid    - Compact grid layout (default)
  • list    - Detailed list with descriptions
  • json    - JSON output for scripting
  • colors  - Show color preview for each theme

Filters:
  • --dark   - Show only dark themes
  • --light  - Show only light themes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			tm.SetVerbose(verbose)

			opts := &theme.ListOptions{
				Format:     format,
				ShowColors: showColors,
				DarkOnly:   darkOnly,
				LightOnly:  lightOnly,
			}

			return tm.ListThemesWithOptions(opts)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "grid", "Output format (grid|list|json|colors)")
	cmd.Flags().BoolVar(&showColors, "colors", false, "Show color preview")
	cmd.Flags().BoolVar(&darkOnly, "dark", false, "Show only dark themes")
	cmd.Flags().BoolVar(&lightOnly, "light", false, "Show only light themes")

	return cmd
}

func randomCmd() *cobra.Command {
	var (
		darkTheme  bool
		lightTheme bool
		withFont   bool
		opacity    float64
		blur       float64
		scheme     string
	)

	cmd := &cobra.Command{
		Use:   "random",
		Short: "Apply a random theme",
		Long: `Apply a random theme with optional constraints:

Theme Selection:

  • Default: Any random theme from collection
  • --dark:  Only dark themes
  • --light: Only light themes  
  • --scheme: Generate new theme with specific scheme

Visual Options:

  • --font:    Auto-select matching font
  • --opacity: Set window transparency
  • --blur:    Add background blur effect

Examples:

  alacritty-colors random --dark
  alacritty-colors random --light --font
  alacritty-colors random --scheme cyberpunk --opacity 0.85`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			tm.SetVerbose(verbose)

			opts := &theme.RandomOptions{
				DarkOnly:  darkTheme,
				LightOnly: lightTheme,
				WithFont:  withFont,
				Opacity:   opacity,
				Blur:      blur,
				Scheme:    scheme,
			}

			return tm.RandomThemeWithOptions(opts)
		},
	}

	cmd.Flags().BoolVar(&darkTheme, "dark", false, "Only select dark themes")
	cmd.Flags().BoolVar(&lightTheme, "light", false, "Only select light themes")
	cmd.Flags().BoolVar(&withFont, "font", false, "Also change font to match theme")
	cmd.Flags().Float64Var(&opacity, "opacity", 0, "Set window opacity (0.0-1.0)")
	cmd.Flags().Float64Var(&blur, "blur", 0, "Set background blur radius")
	cmd.Flags().StringVarP(&scheme, "scheme", "s", "", "Generate new theme with scheme (random|pastel|neon|mono|warm|cool|nature|cyberpunk|dracula|nord|solarized|gruvbox)")

	return cmd
}

func generateCmd() *cobra.Command {
	var (
		scheme     string
		name       string
		save       bool
		darkTheme  bool
		lightTheme bool
		withFont   bool
		opacity    float64
		blur       float64
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a custom theme",
		Long: `Generate a custom theme using various color schemes:

Color Schemes:

  • random     - Completely random colors
  • pastel     - Soft, muted tones perfect for long coding sessions
  • neon       - Bright, vibrant colors for high contrast
  • mono       - Monochromatic grayscale for minimalist setups
  • warm       - Reds, oranges, yellows for cozy environments  
  • cool       - Blues, greens, purples for clean professional look
  • nature     - Earth tones and forest colors
  • cyberpunk  - Neon greens and magentas for futuristic aesthetic
  • dracula    - Dracula-inspired dark theme variations
  • nord       - Nord-inspired cool tones and minimalism
  • solarized  - Solarized variations with scientific precision
  • gruvbox    - Warm retro computing feel

Theme Types:

  • --dark     - Generate dark variant
  • --light    - Generate light variant
  • Default: Auto-determine based on scheme

Examples:

  alacritty-colors generate --scheme cyberpunk --dark
  alacritty-colors generate --scheme nature --light --name forest
  alacritty-colors generate --scheme warm --font --opacity 0.9`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if darkTheme && lightTheme {
				return fmt.Errorf("cannot specify both --dark and --light")
			}

			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			tm.SetVerbose(verbose)

			opts := &theme.GenerateOptions{
				Scheme:     scheme,
				Name:       name,
				Save:       save,
				DarkTheme:  darkTheme,
				LightTheme: lightTheme,
				WithFont:   withFont,
				Opacity:    opacity,
				Blur:       blur,
			}

			return tm.GenerateThemeWithOptions(opts)
		},
	}

	cmd.Flags().StringVarP(&scheme, "scheme", "s", "random", "Color scheme")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Custom theme name")
	cmd.Flags().BoolVar(&save, "save", true, "Save generated theme")
	cmd.Flags().BoolVar(&darkTheme, "dark", false, "Generate dark variant")
	cmd.Flags().BoolVar(&lightTheme, "light", false, "Generate light variant")
	cmd.Flags().BoolVar(&withFont, "font", false, "Auto-select matching font")
	cmd.Flags().Float64Var(&opacity, "opacity", 0, "Set window opacity (0.0-1.0)")
	cmd.Flags().Float64Var(&blur, "blur", 0, "Set background blur radius")

	return cmd
}

func searchCmd() *cobra.Command {
	var (
		format     string
		showColors bool
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search themes by name or tags",
		Long: `Search available themes by name, description, or tags:

The search is case-insensitive and matches partial strings.
Use quotes for exact phrases.

Examples:
  alacritty-colors search dark
  alacritty-colors search "solarized"
  alacritty-colors search nord --colors`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			tm.SetVerbose(verbose)

			opts := &theme.SearchOptions{
				Format:     format,
				ShowColors: showColors,
			}

			return tm.SearchThemesWithOptions(args[0], opts)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "list", "Output format (list|grid|colors)")
	cmd.Flags().BoolVar(&showColors, "colors", false, "Show color preview")

	return cmd
}

func previewCmd() *cobra.Command {
	var (
		apply     bool
		showHex   bool
		slideshow bool
		interval  int
		darkOnly  bool
		lightOnly bool
		randomize bool
		loop      bool
	)

	cmd := &cobra.Command{
		Use:   "preview [theme-name]",
		Short: "Preview themes with live terminal updates",
		Long: `Preview themes in your actual terminal with live updates:

SINGLE THEME MODE (with theme name):
• Temporarily apply a specific theme
• Show theme information and color palette
• Offer to keep or restore previous theme

SLIDESHOW MODE (no theme name or --slideshow flag):
• Automatically cycle through all available themes
• Live preview with configurable intervals
• Interactive controls for navigation and selection
• Alacritty auto-reloads each theme in real-time

Controls during slideshow:
• SPACE/ENTER: Select current theme and exit
• n/RIGHT: Next theme immediately
• p/LEFT: Previous theme
• r: Restart slideshow from beginning
• q/ESC: Quit and restore original theme
• +/-: Increase/decrease cycling speed

Examples:
  alacritty-colors preview                      # Start slideshow mode
  alacritty-colors preview --slideshow         # Explicit slideshow mode
  alacritty-colors preview --interval 5        # 5-second intervals
  alacritty-colors preview --dark --random     # Random dark themes only
  alacritty-colors preview dracula             # Preview specific theme
  alacritty-colors preview nord --apply        # Preview and auto-apply`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			tm.SetVerbose(verbose)

			// If no theme name provided or slideshow flag is set, start slideshow
			if len(args) == 0 || slideshow {
				opts := &theme.SlideshowOptions{
					Interval:   time.Duration(interval) * time.Second,
					DarkOnly:   darkOnly,
					LightOnly:  lightOnly,
					Randomize:  randomize,
					Loop:       loop,
					Categories: nil,
				}
				return tm.ThemeSlideshow(opts)
			}

			// Single theme preview mode
			opts := &theme.PreviewOptions{
				AutoApply: apply,
				ShowHex:   showHex,
			}

			return tm.PreviewThemeWithOptions(args[0], opts)
		},
	}

	cmd.Flags().BoolVarP(&apply, "apply", "a", false, "Apply theme after preview (single theme mode)")
	cmd.Flags().BoolVar(&showHex, "hex", false, "Show hex color values (single theme mode)")
	cmd.Flags().BoolVarP(&slideshow, "slideshow", "s", false, "Force slideshow mode")
	cmd.Flags().IntVarP(&interval, "interval", "i", 3, "Seconds between theme changes (slideshow mode)")
	cmd.Flags().BoolVar(&darkOnly, "dark", false, "Show only dark themes (slideshow mode)")
	cmd.Flags().BoolVar(&lightOnly, "light", false, "Show only light themes (slideshow mode)")
	cmd.Flags().BoolVar(&randomize, "random", false, "Randomize theme order (slideshow mode)")
	cmd.Flags().BoolVar(&loop, "loop", true, "Loop indefinitely (slideshow mode)")

	return cmd
}

func slideshowCmd() *cobra.Command {
	var (
		interval   int
		darkOnly   bool
		lightOnly  bool
		randomize  bool
		loop       bool
		categories []string
	)

	cmd := &cobra.Command{
		Use:   "slideshow",
		Short: "Cycle through themes automatically with live preview",
		Long: `Automatically cycle through themes with live preview in your terminal:

This command will apply themes successively with configurable intervals,
allowing you to see each theme in action in your actual Alacritty terminal.
Alacritty will auto-reload each theme as it's applied.

Features:
• Auto-cycle through themes with customizable intervals
• Live preview in your actual terminal (not just color swatches)
• Interactive controls for navigation and selection
• Filter by dark/light themes or categories
• Randomization option for discovery
• Loop or single-pass modes

Controls during slideshow:
• SPACE/ENTER: Select current theme and exit
• n/RIGHT: Next theme immediately
• p/LEFT: Previous theme
• r: Restart slideshow from beginning
• q/ESC: Quit and restore original theme
• +/-: Increase/decrease cycling speed

Examples:
  alacritty-colors slideshow                    # Default 3-second intervals
  alacritty-colors slideshow --interval 5      # 5-second intervals
  alacritty-colors slideshow --dark --random   # Random dark themes only
  alacritty-colors slideshow --interval 2 --loop  # Loop indefinitely`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			tm.SetVerbose(verbose)

			opts := &theme.SlideshowOptions{
				Interval:   time.Duration(interval) * time.Second,
				DarkOnly:   darkOnly,
				LightOnly:  lightOnly,
				Randomize:  randomize,
				Loop:       loop,
				Categories: categories,
			}

			return tm.ThemeSlideshow(opts)
		},
	}

	cmd.Flags().IntVarP(&interval, "interval", "i", 3, "Seconds between theme changes (1-10)")
	cmd.Flags().BoolVar(&darkOnly, "dark", false, "Show only dark themes")
	cmd.Flags().BoolVar(&lightOnly, "light", false, "Show only light themes")
	cmd.Flags().BoolVar(&randomize, "random", false, "Randomize theme order")
	cmd.Flags().BoolVar(&loop, "loop", true, "Loop indefinitely (default true)")
	cmd.Flags().StringSliceVar(&categories, "categories", nil, "Filter by theme categories")

	return cmd
}

func interactiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "interactive",
		Short: "Interactive color editor TUI",
		Long: `Launch an interactive terminal user interface for editing themes:

Features:
• Browse and select themes from the collection
• Live color preview with terminal output examples
• Real-time color editing with hex and RGB inputs
• Visual color palette display
• Save changes back to theme files
• Keyboard navigation and shortcuts

Controls:
• Arrow keys: Navigate colors
• Enter: Edit selected color
• 's': Save theme
• 'r': Reset to original
• 'q': Quit (with unsaved changes prompt)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			ui.PrintInfo("Launching interactive color editor...")
			ui.PrintInfo("Use 'q' to quit, 's' to save changes")

			return tui.StartInteractive(cfg)
		},
	}
}

func backupCmd() *cobra.Command {
	var (
		name        string
		description string
	)

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Create a configuration backup",
		Long: `Create a backup of your current Alacritty configuration:

Backups are stored with timestamps and can include custom names
and descriptions for easy identification.

Examples:
  alacritty-colors backup
  alacritty-colors backup --name "before-theme-experiment"
  alacritty-colors backup --name "stable" --description "Working config"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			tm.SetVerbose(verbose)

			opts := &theme.BackupOptions{
				Name:        name,
				Description: description,
			}

			return tm.CreateBackupWithOptions(opts)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Custom backup name")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Backup description")

	return cmd
}

func restoreCmd() *cobra.Command {
	var (
		list        bool
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "restore [backup-file]",
		Short: "Restore from configuration backup",
		Long: `Restore your Alacritty configuration from a backup:

Without arguments, shows available backups for interactive selection.
With a backup file argument, restores directly from that backup.

Examples:
  alacritty-colors restore                    # Interactive selection
  alacritty-colors restore --list             # List available backups  
  alacritty-colors restore backup_2024.toml   # Restore specific backup`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			tm.SetVerbose(verbose)

			if list {
				return tm.ListBackups()
			}

			var backupFile string
			if len(args) > 0 {
				backupFile = args[0]
			}

			opts := &theme.RestoreOptions{
				Interactive: interactive || backupFile == "",
			}

			return tm.RestoreBackupWithOptions(backupFile, opts)
		},
	}

	cmd.Flags().BoolVarP(&list, "list", "l", false, "List available backups")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive backup selection")

	return cmd
}

func updateCmd() *cobra.Command {
	var (
		force bool
		check bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update theme database",
		Long: `Update the theme database from official sources:

Downloads the latest themes from the Alacritty themes repository
and updates the local theme collection.

Examples:
  alacritty-colors update           # Update themes
  alacritty-colors update --check   # Check for updates only
  alacritty-colors update --force   # Force re-download all themes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			tm.SetVerbose(verbose)

			opts := &theme.UpdateOptions{
				Force: force,
				Check: check,
			}

			return tm.UpdateThemesWithOptions(opts)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force re-download all themes")
	cmd.Flags().BoolVar(&check, "check", false, "Check for updates only")

	return cmd
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long: `Advanced configuration management commands:

Manage paths, clean up old files, and configure tool behavior.
Use subcommands for specific configuration tasks.`,
	}

	cmd.AddCommand(configCleanBackupsCmd())
	cmd.AddCommand(configCleanThemesCmd())
	cmd.AddCommand(configSetPathCmd())
	cmd.AddCommand(configShowCmd())

	return cmd
}

func configCleanBackupsCmd() *cobra.Command {
	var keepCount int
	cmd := &cobra.Command{
		Use:   "clean-backups",
		Short: "Clean up old backup files",
		Long:  "Remove old backup files, keeping only the most recent ones",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			ui.PrintHeader("Cleaning Backup Files")
			ui.PrintInfo("Keeping %d most recent backups", keepCount)

			// Get list of backup files
			files, err := os.ReadDir(cfg.BackupDir)
			if err != nil {
				return fmt.Errorf("failed to read backup directory: %w", err)
			}

			// Filter only backup files
			backups := []os.DirEntry{}
			for _, file := range files {
				if !file.IsDir() && strings.HasPrefix(file.Name(), "alacritty-") && strings.HasSuffix(file.Name(), ".bak") {
					backups = append(backups, file)
				}
			}

			// Sort by modification time (newest first)
			sort.Slice(backups, func(i, j int) bool {
				fi, _ := backups[i].Info()
				fj, _ := backups[j].Info()
				return fi.ModTime().After(fj.ModTime())
			})

			// Keep only the specified number of backups
			if len(backups) <= keepCount {
				ui.PrintInfo("No backups to clean up (found %d, keeping %d)", len(backups), keepCount)
				return nil
			}

			// Remove older backups
			deleted := 0
			for i := keepCount; i < len(backups); i++ {
				path := filepath.Join(cfg.BackupDir, backups[i].Name())
				if err := os.Remove(path); err != nil {
					ui.PrintWarning("Failed to remove %s: %v", backups[i].Name(), err)
					continue
				}
				deleted++
			}

			ui.PrintSuccess("Cleaned up %d backup files, kept %d most recent", deleted, keepCount)
			return nil
		},
	}

	cmd.Flags().IntVarP(&keepCount, "keep", "k", 5, "number of most recent backups to keep")
	return cmd
}

func configCleanThemesCmd() *cobra.Command {
	var removeGenerated bool
	var removeUnused bool
	cmd := &cobra.Command{
		Use:   "clean-themes",
		Short: "Clean up theme files",
		Long:  "Remove generated or unused theme files",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			ui.PrintHeader("Cleaning Theme Files")

			// Get list of theme files
			files, err := os.ReadDir(cfg.ThemesDir)
			if err != nil {
				return fmt.Errorf("failed to read themes directory: %w", err)
			}

			deleted := 0

			// Process each theme file
			for _, file := range files {
				// Skip directories and current.toml
				if file.IsDir() || file.Name() == "current.toml" {
					continue
				}

				// Check if it's a generated theme
				isGenerated := strings.HasPrefix(file.Name(), "generated-")

				// Skip if it's the current theme
				themeName := strings.TrimSuffix(file.Name(), ".toml")
				isCurrent := themeName == cfg.CurrentTheme

				// Determine if we should delete this file
				shouldDelete := false
				if isGenerated && removeGenerated {
					shouldDelete = true
				}
				if !isGenerated && !isCurrent && removeUnused {
					shouldDelete = true
				}

				// Delete if criteria met
				if shouldDelete {
					path := filepath.Join(cfg.ThemesDir, file.Name())
					if err := os.Remove(path); err != nil {
						ui.PrintWarning("Failed to remove %s: %v", file.Name(), err)
						continue
					}
					deleted++
				}
			}

			ui.PrintSuccess("Cleaned up %d theme files", deleted)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&removeGenerated, "generated", "g", true, "remove generated themes")
	cmd.Flags().BoolVarP(&removeUnused, "unused", "u", false, "remove unused themes (except current)")
	return cmd
}

func configSetPathCmd() *cobra.Command {
	var newConfigPath string
	var newThemesDir string
	var newBackupDir string

	cmd := &cobra.Command{
		Use:   "set-path",
		Short: "Set custom paths for configuration",
		Long:  "Set custom paths for Alacritty config file, themes directory, and backup directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load current config
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			ui.PrintHeader("Setting Custom Paths")

			// Update config file path if specified
			if newConfigPath != "" {
				oldPath := cfg.ConfigFile
				cfg.ConfigFile = newConfigPath
				ui.PrintInfo("Updated config path: %s -> %s", oldPath, newConfigPath)
			}

			// Update themes directory if specified
			if newThemesDir != "" {
				oldPath := cfg.ThemesDir
				cfg.ThemesDir = newThemesDir
				ui.PrintInfo("Updated themes directory: %s -> %s", oldPath, newThemesDir)

				// Create the directory if it doesn't exist
				if err := os.MkdirAll(newThemesDir, 0755); err != nil {
					return fmt.Errorf("failed to create themes directory: %w", err)
				}
			}

			// Update backup directory if specified
			if newBackupDir != "" {
				oldPath := cfg.BackupDir
				cfg.BackupDir = newBackupDir
				ui.PrintInfo("Updated backup directory: %s -> %s", oldPath, newBackupDir)

				// Create the directory if it doesn't exist
				if err := os.MkdirAll(newBackupDir, 0755); err != nil {
					return fmt.Errorf("failed to create backup directory: %w", err)
				}
			}

			// Save the updated config
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			ui.PrintSuccess("Configuration paths updated successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&newConfigPath, "config", "", "new path for Alacritty config file")
	cmd.Flags().StringVar(&newThemesDir, "themes-dir", "", "new path for themes directory")
	cmd.Flags().StringVar(&newBackupDir, "backup-dir", "", "new path for backup directory")
	return cmd
}

func configShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long: `Display current configuration paths and settings:

Shows all configured paths, current theme, and tool status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			return tm.ShowConfig()
		},
	}
}
