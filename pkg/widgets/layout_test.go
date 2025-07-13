package widgets

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateHorizontalSplitWidget(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		config      map[string]interface{}
		children    []Widget
		expectError bool
		validate    func(t *testing.T, widget Widget)
	}{
		{
			name:        "create with children",
			id:          "test_hsplit",
			config:      nil,
			children:    createTestChildren(2),
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				assert.Equal(t, "test_hsplit", widget.GetID())
				assert.Equal(t, "horizontal_split", widget.GetType())
				assert.Len(t, widget.GetChildren(), 2)
			},
		},
		{
			name:        "create without children",
			id:          "empty_hsplit",
			config:      nil,
			children:    nil,
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				assert.Equal(t, "empty_hsplit", widget.GetID())
				assert.Equal(t, "horizontal_split", widget.GetType())
				assert.Empty(t, widget.GetChildren())
			},
		},
		{
			name: "create with size config",
			id:   "sized_hsplit",
			config: map[string]interface{}{
				"sizes": []interface{}{0.3, 0.7},
			},
			children:    createTestChildren(2),
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				hsplit := widget.(*SplitWidget)
				assert.Equal(t, "sized_hsplit", hsplit.GetID())
				assert.Len(t, hsplit.GetChildren(), 2)
				// Config should be stored
				assert.NotNil(t, hsplit.Config)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			widget, err := CreateHorizontalSplitWidget(tt.id, tt.config, tt.children)

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

func TestCreateVerticalSplitWidget(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		config      map[string]interface{}
		children    []Widget
		expectError bool
		validate    func(t *testing.T, widget Widget)
	}{
		{
			name:        "create with children",
			id:          "test_vsplit",
			config:      nil,
			children:    createTestChildren(3),
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				assert.Equal(t, "test_vsplit", widget.GetID())
				assert.Equal(t, "vertical_split", widget.GetType())
				assert.Len(t, widget.GetChildren(), 3)
			},
		},
		{
			name:        "create without children",
			id:          "empty_vsplit",
			config:      nil,
			children:    nil,
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				assert.Equal(t, "empty_vsplit", widget.GetID())
				assert.Equal(t, "vertical_split", widget.GetType())
				assert.Empty(t, widget.GetChildren())
			},
		},
		{
			name: "create with size config",
			id:   "sized_vsplit",
			config: map[string]interface{}{
				"sizes": []interface{}{0.25, 0.5, 0.25},
			},
			children:    createTestChildren(3),
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				vsplit := widget.(*SplitWidget)
				assert.Equal(t, "sized_vsplit", vsplit.GetID())
				assert.Len(t, vsplit.GetChildren(), 3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			widget, err := CreateVerticalSplitWidget(tt.id, tt.config, tt.children)

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

func TestCreateGrowWidget(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		config      map[string]interface{}
		children    []Widget
		expectError bool
		validate    func(t *testing.T, widget Widget)
	}{
		{
			name:        "create with default grow",
			id:          "test_grow",
			config:      nil,
			children:    createTestChildren(1),
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				assert.Equal(t, "test_grow", widget.GetID())
				assert.Equal(t, "grow", widget.GetType())
				assert.Len(t, widget.GetChildren(), 1)

				growWidget := widget.(*GrowWidget)
				assert.Equal(t, "1", growWidget.GrowValue) // Default grow value
			},
		},
		{
			name: "create with string grow value",
			id:   "string_grow",
			config: map[string]interface{}{
				"grow": "2",
			},
			children:    nil,
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				growWidget := widget.(*GrowWidget)
				assert.Equal(t, "2", growWidget.GrowValue)
			},
		},
		{
			name: "create with float grow value",
			id:   "float_grow",
			config: map[string]interface{}{
				"grow": 1.5,
			},
			children:    nil,
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				growWidget := widget.(*GrowWidget)
				assert.Equal(t, "2", growWidget.GrowValue) // Should be formatted as integer
			},
		},
		{
			name: "create with int grow value",
			id:   "int_grow",
			config: map[string]interface{}{
				"grow": 3,
			},
			children:    nil,
			expectError: false,
			validate: func(t *testing.T, widget Widget) {
				growWidget := widget.(*GrowWidget)
				assert.Equal(t, "3", growWidget.GrowValue)
			},
		},
		{
			name: "create with invalid grow value type",
			id:   "invalid_grow",
			config: map[string]interface{}{
				"grow": []string{"invalid"},
			},
			children:    nil,
			expectError: false, // Should fall back to default
			validate: func(t *testing.T, widget Widget) {
				growWidget := widget.(*GrowWidget)
				assert.Equal(t, "1", growWidget.GrowValue) // Should use default
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			widget, err := CreateGrowWidget(tt.id, tt.config, tt.children)

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

func TestHorizontalSplitWidgetInit(t *testing.T) {
	children := createTestChildren(2)
	widget, err := CreateHorizontalSplitWidget("test_hsplit", map[string]interface{}{
		"sizes": []interface{}{0.4, 0.6},
	}, children)
	require.NoError(t, err)

	hsplit := widget.(*SplitWidget)

	ctx := context.Background()
	err = hsplit.Init(ctx)
	require.NoError(t, err)

	// Data should be set after init
	assert.NotNil(t, hsplit.Data)

	// Check data structure
	data, ok := hsplit.Data.(*LayoutData)
	require.True(t, ok, "Data should be *LayoutData")

	assert.Equal(t, "horizontal_split", data.Type)
	assert.Equal(t, "horizontal", data.Direction)
	assert.NotNil(t, data.Sizes)

	// LastUpdate should be set
	assert.False(t, hsplit.LastUpdate.IsZero())
}

func TestVerticalSplitWidgetInit(t *testing.T) {
	children := createTestChildren(3)
	widget, err := CreateVerticalSplitWidget("test_vsplit", nil, children)
	require.NoError(t, err)

	vsplit := widget.(*SplitWidget)

	ctx := context.Background()
	err = vsplit.Init(ctx)
	require.NoError(t, err)

	// Data should be set after init
	assert.NotNil(t, vsplit.Data)

	// Check data structure
	data, ok := vsplit.Data.(*LayoutData)
	require.True(t, ok, "Data should be *LayoutData")

	assert.Equal(t, "vertical_split", data.Type)
	assert.Equal(t, "vertical", data.Direction)

	// LastUpdate should be set
	assert.False(t, vsplit.LastUpdate.IsZero())
}

func TestGrowWidgetInit(t *testing.T) {
	children := createTestChildren(1)
	widget, err := CreateGrowWidget("test_grow", map[string]interface{}{
		"grow": "2",
	}, children)
	require.NoError(t, err)

	growWidget := widget.(*GrowWidget)

	ctx := context.Background()
	err = growWidget.Init(ctx)
	require.NoError(t, err)

	// Data should be set after init
	assert.NotNil(t, growWidget.Data)

	// Check data structure
	data, ok := growWidget.Data.(map[string]interface{})
	require.True(t, ok, "Data should be a map")

	assert.Equal(t, "grow", data["type"])
	assert.Equal(t, "2", data["grow_value"])

	// LastUpdate should be set
	assert.False(t, growWidget.LastUpdate.IsZero())
}

func TestLayoutWidgetBasics(t *testing.T) {
	children := createTestChildren(2)

	t.Run("widget properties", func(t *testing.T) {
		widgets := []struct {
			w            Widget
			expectedType string
		}{
			{mustCreate(CreateHorizontalSplitWidget("h", nil, children)), "horizontal_split"},
			{mustCreate(CreateVerticalSplitWidget("v", nil, children)), "vertical_split"},
			{mustCreate(CreateGrowWidget("g", nil, children)), "grow"},
		}

		for _, test := range widgets {
			assert.Equal(t, test.expectedType, test.w.GetType())
			assert.Len(t, test.w.GetChildren(), 2)
			assert.NoError(t, test.w.Close())
		}
	})
}

func mustCreate(w Widget, err error) Widget {
	if err != nil {
		panic(err)
	}
	return w
}

func TestLayoutWidgetNesting(t *testing.T) {
	// Create a nested layout structure
	innerChildren := createTestChildren(2)

	hsplit, err := CreateHorizontalSplitWidget("inner_hsplit", nil, innerChildren)
	require.NoError(t, err)

	vsplit, err := CreateVerticalSplitWidget("inner_vsplit", nil, innerChildren)
	require.NoError(t, err)

	outerChildren := []Widget{hsplit, vsplit}
	outerWidget, err := CreateVerticalSplitWidget("outer_vsplit", nil, outerChildren)
	require.NoError(t, err)

	// Test the nested structure
	assert.Equal(t, "outer_vsplit", outerWidget.GetID())
	assert.Len(t, outerWidget.GetChildren(), 2)

	// First child should be horizontal split
	firstChild := outerWidget.GetChildren()[0]
	assert.Equal(t, "inner_hsplit", firstChild.GetID())
	assert.Equal(t, "horizontal_split", firstChild.GetType())
	assert.Len(t, firstChild.GetChildren(), 2)

	// Second child should be vertical split
	secondChild := outerWidget.GetChildren()[1]
	assert.Equal(t, "inner_vsplit", secondChild.GetID())
	assert.Equal(t, "vertical_split", secondChild.GetType())
	assert.Len(t, secondChild.GetChildren(), 2)
}

func TestGrowWidgetGetGrowValue(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		expected string
	}{
		{
			name:     "default grow value",
			config:   nil,
			expected: "1",
		},
		{
			name:     "string grow value",
			config:   map[string]interface{}{"grow": "3"},
			expected: "3",
		},
		{
			name:     "float grow value",
			config:   map[string]interface{}{"grow": 2.0},
			expected: "2",
		},
		{
			name:     "int grow value",
			config:   map[string]interface{}{"grow": 5},
			expected: "5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			widget, err := CreateGrowWidget("test_grow", tt.config, nil)
			require.NoError(t, err)

			growWidget := widget.(*GrowWidget)
			assert.Equal(t, tt.expected, growWidget.GetGrowValue())
		})
	}
}

// Helper function to create test child widgets.
func createTestChildren(count int) []Widget {
	children := make([]Widget, count)
	for i := 0; i < count; i++ {
		children[i] = &BaseWidget{
			ID:   fmt.Sprintf("child_%d", i),
			Type: "test_child",
		}
	}
	return children
}

// Benchmark tests.
func BenchmarkHorizontalSplitWidgetCreation(b *testing.B) {
	children := createTestChildren(2)
	config := map[string]interface{}{
		"sizes": []interface{}{0.5, 0.5},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widget, err := CreateHorizontalSplitWidget("bench_hsplit", config, children)
		if err != nil {
			b.Fatal(err)
		}
		if widget == nil {
			b.Fatal("widget is nil")
		}
	}
}

func BenchmarkGrowWidgetInit(b *testing.B) {
	widget, err := CreateGrowWidget("bench_grow", map[string]interface{}{
		"grow": "2",
	}, createTestChildren(1))
	if err != nil {
		b.Fatal(err)
	}

	growWidget := widget.(*GrowWidget)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := growWidget.Init(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
