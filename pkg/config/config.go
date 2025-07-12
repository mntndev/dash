package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Dashboard     DashboardConfig     `yaml:"dashboard"`
	Integrations  IntegrationsConfig  `yaml:"integrations"`
}

type DashboardConfig struct {
	Title      string       `yaml:"title"`
	Theme      string       `yaml:"theme"`
	Fullscreen bool         `yaml:"fullscreen"`
	Widget     WidgetConfig `yaml:"widget"`
}


type WidgetConfig struct {
	Type     string                 `yaml:"type"`
	Config   map[string]interface{} `yaml:"config,omitempty"`
	Children []WidgetConfig         `yaml:"children,omitempty"`
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
	data, err := os.ReadFile(path)
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

// findConfigFile looks for config files in order of preference
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

// getConfigPaths returns all possible config file locations in order of preference
func getConfigPaths() []string {
	homeDir, _ := os.UserHomeDir()
	
	paths := []string{
		filepath.Join(".", "config.yaml"),                    // Current directory
		filepath.Join(homeDir, ".config", "dash", "config.yaml"), // XDG config
		filepath.Join(homeDir, ".dash.yaml"),                // User home
	}
	
	// Add system path on Unix-like systems
	if homeDir != "" && homeDir != "/" {
		paths = append(paths, "/etc/dash/config.yaml")
	}
	
	return paths
}