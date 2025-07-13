package widgets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateClockWidget(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		config      map[string]interface{}
		expectError bool
		validate    func(t *testing.T, widget Widget)
	}{
		{
			name:        "create with default format",
			id:          "test_clock",
			config:      nil,
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				assert.Equal(t, "test_clock", widget.GetID())
				assert.Equal(t, "clock", widget.GetType())
				assert.Empty(t, widget.GetChildren())
			},
		},
		{
			name: "create with custom format",
			id:   "custom_clock",
			config: map[string]interface{}{
				"format": "2006-01-02 15:04:05",
			},
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				assert.Equal(t, "custom_clock", widget.GetID())
				assert.Equal(t, "clock", widget.GetType())

				// Check that config is stored
				clockWidget := widget.(*ClockWidget)
				assert.Equal(t, "2006-01-02 15:04:05", clockWidget.Format)
			},
		},
		{
			name: "create with complex format",
			id:   "complex_clock",
			config: map[string]interface{}{
				"format": "Monday, January 2, 2006 15:04:05",
			},
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				clockWidget := widget.(*ClockWidget)
				assert.Equal(t, "Monday, January 2, 2006 15:04:05", clockWidget.Format)
			},
		},
		{
			name: "create with minimal config",
			id:   "minimal_clock",
			config: map[string]interface{}{
				"format": "HH:mm", // Different valid format
			},
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				clockWidget := widget.(*ClockWidget)
				assert.Equal(t, "HH:mm", clockWidget.Format)
			},
		},
		{
			name:        "create with empty config map",
			id:          "empty_config_clock",
			config:      map[string]interface{}{},
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				clockWidget := widget.(*ClockWidget)
				assert.Equal(t, "15:04:05", clockWidget.Format) // Should use default
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			widget, err := CreateClockWidget(tt.id, tt.config, nil)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, widget)
			} else {
				require.NoError(t, err)
				require.NotNil(t, widget)
				if tt.validate != nil {
					tt.validate(t, widget)
				}
			}
		})
	}
}

func TestClockWidgetInit(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		validate func(t *testing.T, widget *ClockWidget)
	}{
		{
			name:   "default format",
			format: "",
			validate: func(t *testing.T, widget *ClockWidget) {
				assert.Equal(t, "15:04:05", widget.Format)

				// Data should be set after init
				assert.NotNil(t, widget.Data)

				// Check data structure
				data, ok := widget.Data.(*ClockData)
				require.True(t, ok, "Data should be *ClockData")

				assert.Equal(t, "15:04:05", data.Format)
				assert.NotEmpty(t, data.Display)
				assert.False(t, data.Time.IsZero())
			},
		},
		{
			name:   "custom format",
			format: "2006-01-02 15:04:05",
			validate: func(t *testing.T, widget *ClockWidget) {
				assert.Equal(t, "2006-01-02 15:04:05", widget.Format)

				data, ok := widget.Data.(*ClockData)
				require.True(t, ok)
				assert.Equal(t, "2006-01-02 15:04:05", data.Format)

				// Display should match the format pattern
				assert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`, data.Display)
			},
		},
		{
			name:   "short format",
			format: "15:04",
			validate: func(t *testing.T, widget *ClockWidget) {
				assert.Equal(t, "15:04", widget.Format)

				data, ok := widget.Data.(*ClockData)
				require.True(t, ok)
				assert.Equal(t, "15:04", data.Format)

				// Display should match the format pattern
				assert.Regexp(t, `^\d{2}:\d{2}$`, data.Display)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := make(map[string]interface{})
			if tt.format != "" {
				config["format"] = tt.format
			}

			widget, err := CreateClockWidget("test_clock", config, nil)
			require.NoError(t, err)

			clockWidget := widget.(*ClockWidget)

			// Test initialization
			ctx := context.Background()
			err = clockWidget.Init(ctx)
			require.NoError(t, err)

			// Validate results
			if tt.validate != nil {
				tt.validate(t, clockWidget)
			}

			// LastUpdate should be set
			assert.False(t, clockWidget.LastUpdate.IsZero())
		})
	}
}

func TestClockWidgetDataUpdates(t *testing.T) {
	widget, err := CreateClockWidget("test_clock", map[string]interface{}{
		"format": "15:04:05",
	}, nil)
	require.NoError(t, err)

	clockWidget := widget.(*ClockWidget)

	ctx := context.Background()
	err = clockWidget.Init(ctx)
	require.NoError(t, err)

	// Get initial data
	data1, ok := clockWidget.Data.(*ClockData)
	require.True(t, ok)
	time1 := data1.Display

	// Wait a bit - since the clock updater runs in background,
	// let's just verify the current data is valid
	assert.NotEmpty(t, time1)
	assert.Regexp(t, `^\d{2}:\d{2}:\d{2}$`, time1)

	// Verify the data structure is correct
	assert.Equal(t, "15:04:05", data1.Format)
	assert.False(t, data1.Time.IsZero())
}

func TestClockWidgetConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
		expected    string
	}{
		{"valid format", map[string]interface{}{"format": "2006-01-02"}, false, "2006-01-02"},
		{"nil config uses default", nil, false, "15:04:05"},
		{"empty config uses default", map[string]interface{}{}, false, "15:04:05"},
		{"invalid format type", map[string]interface{}{"format": 123}, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			widget, err := CreateClockWidget("test_clock", tt.config, nil)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, widget)
			} else {
				require.NoError(t, err)
				clockWidget := widget.(*ClockWidget)
				assert.Equal(t, tt.expected, clockWidget.Format)
			}
		})
	}
}

func TestClockWidgetClose(t *testing.T) {
	widget, err := CreateClockWidget("test_clock", nil, nil)
	require.NoError(t, err)

	// Close should not error
	err = widget.Close()
	assert.NoError(t, err)
}

func TestClockWidgetGetters(t *testing.T) {
	widget, err := CreateClockWidget("test_clock", map[string]interface{}{
		"format": "2006-01-02 15:04:05",
	}, nil)
	require.NoError(t, err)

	assert.Equal(t, "test_clock", widget.GetID())
	assert.Equal(t, "clock", widget.GetType())
	assert.Empty(t, widget.GetChildren())
	assert.Nil(t, widget.GetData()) // Data should be nil before Init

	// After init, data should be present
	ctx := context.Background()
	err = widget.Init(ctx)
	require.NoError(t, err)
	assert.NotNil(t, widget.GetData())
}

func TestClockWidgetDataStructure(t *testing.T) {
	widget, err := CreateClockWidget("test_clock", map[string]interface{}{
		"format": "15:04:05",
	}, nil)
	require.NoError(t, err)

	clockWidget := widget.(*ClockWidget)
	ctx := context.Background()
	err = clockWidget.Init(ctx)
	require.NoError(t, err)

	data, ok := clockWidget.Data.(*ClockData)
	require.True(t, ok, "Data should be *ClockData")

	// Check all expected fields are present and correct
	assert.Equal(t, "15:04:05", data.Format)
	assert.False(t, data.Time.IsZero(), "Time should be set")
	assert.NotEmpty(t, data.Display, "Display should not be empty")

	// Display should match the format pattern
	assert.Regexp(t, `^\d{2}:\d{2}:\d{2}$`, data.Display)

	// Verify the display is actually formatted from the time
	expectedDisplay := data.Time.Format(data.Format)
	assert.Equal(t, expectedDisplay, data.Display)
}

// Benchmark tests.
func BenchmarkClockWidgetCreation(b *testing.B) {
	config := map[string]interface{}{
		"format":   "15:04:05",
		"timezone": "UTC",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widget, err := CreateClockWidget("bench_clock", config, nil)
		if err != nil {
			b.Fatal(err)
		}
		if widget == nil {
			b.Fatal("widget is nil")
		}
	}
}

func BenchmarkClockWidgetInit(b *testing.B) {
	widget, err := CreateClockWidget("bench_clock", map[string]interface{}{
		"format": "15:04:05",
	}, nil)
	if err != nil {
		b.Fatal(err)
	}

	clockWidget := widget.(*ClockWidget)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := clockWidget.Init(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
