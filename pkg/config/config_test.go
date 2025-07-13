package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
		validate    func(t *testing.T, config *Config)
	}{
		{
			name:        "valid complete config",
			filename:    "valid_complete.yaml",
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "Complete Test Dashboard", config.Dashboard.Title)
				assert.Equal(t, "dark", config.Dashboard.Theme)
				assert.True(t, config.Dashboard.Fullscreen)
				assert.Equal(t, "vertical_split", config.Dashboard.Widget.Type)
				assert.Len(t, config.Dashboard.Widget.Children, 2)

				// Check integrations
				assert.NotNil(t, config.Integrations.HomeAssistant)
				assert.Equal(t, "ws://localhost:8123/api/websocket", config.Integrations.HomeAssistant.URL)
				assert.Equal(t, "test_token_123", config.Integrations.HomeAssistant.Token)

				assert.NotNil(t, config.Integrations.Dexcom)
				assert.Equal(t, "test_user", config.Integrations.Dexcom.Username)
				assert.Equal(t, "test_pass", config.Integrations.Dexcom.Password)

				assert.NotNil(t, config.Integrations.Prometheus)
				assert.Equal(t, "http://localhost:9090", config.Integrations.Prometheus.URL)

				assert.Len(t, config.Integrations.RSS, 1)
				assert.Equal(t, "https://example.com/feed.xml", config.Integrations.RSS[0].URL)
				assert.Equal(t, "1h", config.Integrations.RSS[0].RefreshInterval)
			},
		},
		{
			name:        "valid minimal config",
			filename:    "valid_minimal.yaml",
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "Minimal Test Dashboard", config.Dashboard.Title)
				assert.Equal(t, "dark", config.Dashboard.Theme)
				assert.False(t, config.Dashboard.Fullscreen) // Should default to false
				assert.Equal(t, "clock", config.Dashboard.Widget.Type)
				assert.Empty(t, config.Dashboard.Widget.Children)

				// Integrations should be nil/empty
				assert.Nil(t, config.Integrations.HomeAssistant)
				assert.Nil(t, config.Integrations.Dexcom)
				assert.Nil(t, config.Integrations.Prometheus)
				assert.Empty(t, config.Integrations.RSS)
			},
		},
		{
			name:        "file not found",
			filename:    "nonexistent.yaml",
			expectError: true,
		},
		{
			name:        "invalid YAML syntax",
			filename:    "invalid_syntax.yaml",
			expectError: true,
		},
		{
			name:        "empty file",
			filename:    "empty.yaml",
			expectError: true, // Should fail validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", tt.filename)
			config, err := LoadConfig(path)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				if tt.validate != nil {
					tt.validate(t, config)
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &Config{
				Dashboard: DashboardConfig{
					Title: "Test Dashboard",
					Widget: WidgetConfig{
						Type: "clock",
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing title",
			config: &Config{
				Dashboard: DashboardConfig{
					Widget: WidgetConfig{
						Type: "clock",
					},
				},
			},
			expectError: true,
			errorMsg:    "dashboard title is required",
		},
		{
			name: "missing widget type",
			config: &Config{
				Dashboard: DashboardConfig{
					Title:  "Test Dashboard",
					Widget: WidgetConfig{},
				},
			},
			expectError: true,
			errorMsg:    "dashboard widget type is required",
		},
		{
			name: "empty dashboard",
			config: &Config{
				Dashboard: DashboardConfig{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateWidget(t *testing.T) {
	tests := []struct {
		name        string
		widget      WidgetConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid clock widget",
			widget: WidgetConfig{
				Type: "clock",
				Config: map[string]interface{}{
					"format": "15:04:05",
				},
			},
			expectError: false,
		},
		{
			name: "valid layout with children",
			widget: WidgetConfig{
				Type: "horizontal_split",
				Children: []WidgetConfig{
					{Type: "clock"},
					{Type: "home_assistant.entity"},
				},
			},
			expectError: false,
		},
		{
			name:        "empty widget type",
			widget:      WidgetConfig{},
			expectError: true,
			errorMsg:    "widget type is required",
		},
		{
			name: "layout widget without children",
			widget: WidgetConfig{
				Type: "horizontal_split",
			},
			expectError: true,
			errorMsg:    "layout widget must have at least one child",
		},
		{
			name: "vertical split without children",
			widget: WidgetConfig{
				Type: "vertical_split",
			},
			expectError: true,
			errorMsg:    "layout widget must have at least one child",
		},
		{
			name: "layout with invalid child",
			widget: WidgetConfig{
				Type: "horizontal_split",
				Children: []WidgetConfig{
					{Type: "clock"},
					{}, // Invalid child with empty type
				},
			},
			expectError: true,
			errorMsg:    "child 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWidget(tt.widget)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigPathFunctions(t *testing.T) {
	t.Run("get config paths", func(t *testing.T) {
		paths := getConfigPaths()
		assert.Contains(t, paths, filepath.Join(".", "config.yaml"))

		homeDir, err := os.UserHomeDir()
		if err == nil && homeDir != "" {
			expectedHomePath := filepath.Join(homeDir, ".config", "dash", "config.yaml")
			assert.Contains(t, paths, expectedHomePath)
		}
	})

	t.Run("get default config path", func(t *testing.T) {
		path := GetDefaultConfigPath()
		assert.NotEmpty(t, path)
	})

	t.Run("find config file", func(t *testing.T) {
		path := findConfigFile()
		assert.NotEmpty(t, path)
	})
}

// Benchmark tests.
func BenchmarkLoadValidConfig(b *testing.B) {
	path := filepath.Join("testdata", "valid_complete.yaml")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadConfig(path)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateConfig(b *testing.B) {
	config := &Config{
		Dashboard: DashboardConfig{
			Title: "Benchmark Dashboard",
			Widget: WidgetConfig{
				Type: "vertical_split",
				Children: []WidgetConfig{
					{Type: "clock"},
					{Type: "home_assistant.entity"},
					{Type: "dexcom"},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validateConfig(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}
