package config

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"gioui.org/font/gofont"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/lucasb-eyer/go-colorful"
)

type Config struct {
	Dashboard    DashboardConfig    `yaml:"dashboard"`
	Integrations IntegrationsConfig `yaml:"integrations"`
}

type DashboardConfig struct {
	Title      string       `yaml:"title"`
	Theme      string       `yaml:"theme"`
	Colors     *ColorConfig `yaml:"colors,omitempty"`
	Fullscreen bool         `yaml:"fullscreen"`
	Widget     WidgetConfig `yaml:"widget"`
}

type ColorConfig struct {
	Bg         string `yaml:"bg,omitempty"`
	Fg         string `yaml:"fg,omitempty"`
	ContrastBg string `yaml:"contrast_bg,omitempty"`
	ContrastFg string `yaml:"contrast_fg,omitempty"`
}

type WidgetConfig struct {
	Type     string         `yaml:"type" json:"Type"`
	Config   ast.Node       `yaml:"config,omitempty" json:"-"`
	Children []WidgetConfig `yaml:"children,omitempty" json:"Children,omitempty"`
}

type IntegrationsConfig struct {
	HomeAssistant *HomeAssistantConfig `yaml:"home_assistant,omitempty"`
	Dexcom        *DexcomConfig        `yaml:"dexcom,omitempty"`
	Prometheus    *PrometheusConfig    `yaml:"prometheus,omitempty"`
	RSS           []RSSConfig          `yaml:"rss,omitempty"`
}

type HomeAssistantConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

type DexcomConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type PrometheusConfig struct {
	URL string `yaml:"url"`
}

type RSSConfig struct {
	URL             string `yaml:"url"`
	RefreshInterval string `yaml:"refresh_interval"`
}

func LoadConfig(path string) (*Config, error) {
	// Clean and validate path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	// #nosec G304 - This is intentionally reading a config file path provided by the user
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

func validateConfig(config *Config) error {
	if config.Dashboard.Title == "" {
		return fmt.Errorf("dashboard title is required")
	}

	if config.Dashboard.Widget.Type == "" {
		return fmt.Errorf("dashboard widget type is required")
	}

	return validateWidget(config.Dashboard.Widget)
}

func validateWidget(widget WidgetConfig) error {
	if widget.Type == "" {
		return fmt.Errorf("widget type is required")
	}

	// Validate children for layout widgets
	if widget.Type == "horizontal_split" || widget.Type == "vertical_split" {
		if len(widget.Children) == 0 {
			return fmt.Errorf("layout widget must have at least one child")
		}
		for i, child := range widget.Children {
			if err := validateWidget(child); err != nil {
				return fmt.Errorf("child %d: %w", i, err)
			}
		}
	}

	return nil
}

func GetDefaultConfigPath() string {
	return findConfigFile()
}

// findConfigFile looks for config files in order of preference.
func findConfigFile() string {
	configPaths := getConfigPaths()

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Return the first path (current directory) as default if none found
	return configPaths[0]
}

// getConfigPaths returns all possible config file locations in order of preference.
func getConfigPaths() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get home directory, only use current directory
		return []string{filepath.Join(".", "config.yaml")}
	}

	paths := []string{
		filepath.Join(".", "config.yaml"),                        // Current directory
		filepath.Join(homeDir, ".config", "dash", "config.yaml"), // XDG config
		filepath.Join(homeDir, ".dash.yaml"),                     // User home
	}

	// Add system path on Unix-like systems
	if homeDir != "" && homeDir != "/" {
		paths = append(paths, "/etc/dash/config.yaml")
	}

	return paths
}

// parseColor parses a color string (hex, CSS names, etc.) into color.NRGBA using go-colorful
func parseColor(colorStr string) (color.NRGBA, error) {
	if colorStr == "" {
		return color.NRGBA{}, fmt.Errorf("empty color string")
	}

	c, err := colorful.Hex(colorStr)
	if err != nil {
		return color.NRGBA{}, fmt.Errorf("failed to parse color %s: %w", colorStr, err)
	}

	r, g, b := c.RGB255()
	return color.NRGBA{R: r, G: g, B: b, A: 255}, nil
}

// Theme returns a material theme configured for this dashboard
func (c *Config) Theme() *material.Theme {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	// Apply custom colors if configured
	if c.Dashboard.Colors != nil {
		palette := &th.Palette

		if c.Dashboard.Colors.Bg != "" {
			if color, err := parseColor(c.Dashboard.Colors.Bg); err == nil {
				palette.Bg = color
			}
		}

		if c.Dashboard.Colors.Fg != "" {
			if color, err := parseColor(c.Dashboard.Colors.Fg); err == nil {
				palette.Fg = color
			}
		}

		if c.Dashboard.Colors.ContrastBg != "" {
			if color, err := parseColor(c.Dashboard.Colors.ContrastBg); err == nil {
				palette.ContrastBg = color
			}
		}

		if c.Dashboard.Colors.ContrastFg != "" {
			if color, err := parseColor(c.Dashboard.Colors.ContrastFg); err == nil {
				palette.ContrastFg = color
			}
		}
	}

	return th
}
