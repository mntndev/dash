package widgets

import (
	"context"
	"testing"

	"github.com/goccy/go-yaml/ast"
	"github.com/mntndev/dash/pkg/integrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockProvider implements Provider interface for testing.
type MockProvider struct {
	haClient     *integrations.HomeAssistantClient
	dexcomClient *integrations.DexcomClient
	events       []MockEvent
}

type MockEvent struct {
	EventName string
	Data      interface{}
}

func (m *MockProvider) GetHAClient() *integrations.HomeAssistantClient {
	return m.haClient
}

func (m *MockProvider) GetDexcomClient() *integrations.DexcomClient {
	return m.dexcomClient
}

func (m *MockProvider) Emit(event string, data interface{}) {
	m.events = append(m.events, MockEvent{
		EventName: event,
		Data:      data,
	})
}

func (m *MockProvider) GetEmittedEvents() []MockEvent {
	return m.events
}

func (m *MockProvider) ClearEvents() {
	m.events = nil
}

func TestWidgetRegistry(t *testing.T) {
	t.Run("register and create widgets", func(t *testing.T) {
		registry := NewWidgetRegistry()
		provider := &MockProvider{}

		// Test registering a widget
		registry.Register("test_widget", func(id string, config ast.Node, children []Widget, provider Provider) (Widget, error) {
			return &BaseWidget{
				ID:   id,
				Type: "test_widget",
			}, nil
		})

		// Test creating the widget
		widget, err := registry.Create("test_widget", "test_id", nil, nil, provider)
		require.NoError(t, err)
		assert.Equal(t, "test_id", widget.GetID())
		assert.Equal(t, "test_widget", widget.GetType())

		// Test supported types
		types := registry.GetSupportedTypes()
		assert.Contains(t, types, "test_widget")
	})

	t.Run("create unsupported widget type", func(t *testing.T) {
		registry := NewWidgetRegistry()
		provider := &MockProvider{}

		widget, err := registry.Create("nonexistent", "test_id", nil, nil, provider)
		assert.Error(t, err)
		assert.Nil(t, widget)
		assert.Contains(t, err.Error(), "unsupported widget type")
	})

	t.Run("builtin widgets are registered", func(t *testing.T) {
		registry := NewWidgetRegistry()
		registerBuiltinWidgets(registry)

		expectedTypes := []string{
			"home_assistant.entity",
			"home_assistant.button",
			"home_assistant.switch",
			"home_assistant.light",
			"dexcom",
			"clock",
			"horizontal_split",
			"vertical_split",
			"grow",
		}

		supportedTypes := registry.GetSupportedTypes()
		for _, expectedType := range expectedTypes {
			assert.Contains(t, supportedTypes, expectedType, "Widget type %s should be registered", expectedType)
		}
	})
}

func TestDefaultWidgetFactory(t *testing.T) {
	provider := &MockProvider{}
	factory := NewDefaultWidgetFactory(provider)

	t.Run("create widgets with provider", func(t *testing.T) {
		// Test creating a clock widget
		widget, err := factory.Create("clock", "test_clock", configToNode(map[string]interface{}{
			"format": "15:04:05",
		}), nil)
		require.NoError(t, err)
		assert.Equal(t, "test_clock", widget.GetID())
		assert.Equal(t, "clock", widget.GetType())
	})

	t.Run("create layout widget with children", func(t *testing.T) {
		// Create child widgets first
		child1, err := factory.Create("clock", "child1", nil, nil)
		require.NoError(t, err)

		child2, err := factory.Create("clock", "child2", nil, nil)
		require.NoError(t, err)

		children := []Widget{child1, child2}

		// Create horizontal split with children
		widget, err := factory.Create("horizontal_split", "test_split", nil, children)
		require.NoError(t, err)
		assert.Equal(t, "test_split", widget.GetID())
		assert.Equal(t, "horizontal_split", widget.GetType())
		assert.Len(t, widget.GetChildren(), 2)
	})

	t.Run("get supported types", func(t *testing.T) {
		types := factory.GetSupportedTypes()
		assert.Contains(t, types, "clock")
		assert.Contains(t, types, "horizontal_split")
		assert.Contains(t, types, "vertical_split")
		assert.Contains(t, types, "grow")
	})

	t.Run("create unsupported widget", func(t *testing.T) {
		widget, err := factory.Create("unsupported_type", "test_id", nil, nil)
		assert.Error(t, err)
		assert.Nil(t, widget)
	})
}

func TestWidgetManager(t *testing.T) {
	provider := &MockProvider{}
	factory := NewDefaultWidgetFactory(provider)
	manager := NewWidgetManager(factory)

	t.Run("create and store widget", func(t *testing.T) {
		err := manager.CreateWidget("test_widget", "clock", configToNode(map[string]interface{}{
			"format": "15:04:05",
		}), nil)
		require.NoError(t, err)

		// Retrieve the widget
		widget, exists := manager.GetWidget("test_widget")
		assert.True(t, exists)
		assert.Equal(t, "test_widget", widget.GetID())
		assert.Equal(t, "clock", widget.GetType())
	})

	t.Run("get non-existent widget", func(t *testing.T) {
		widget, exists := manager.GetWidget("nonexistent")
		assert.False(t, exists)
		assert.Nil(t, widget)
	})

	t.Run("store widget directly", func(t *testing.T) {
		testWidget := &BaseWidget{
			ID:   "direct_widget",
			Type: "test",
		}

		manager.StoreWidget("direct_widget", testWidget)

		widget, exists := manager.GetWidget("direct_widget")
		assert.True(t, exists)
		assert.Equal(t, "direct_widget", widget.GetID())
	})

	t.Run("get all widgets", func(t *testing.T) {
		// Clear any existing widgets by creating a new manager
		testManager := NewWidgetManager(factory)

		// Add a few widgets
		err := testManager.CreateWidget("widget1", "clock", nil, nil)
		require.NoError(t, err)

		err = testManager.CreateWidget("widget2", "grow", nil, nil)
		require.NoError(t, err)

		allWidgets := testManager.GetAllWidgets()
		assert.Len(t, allWidgets, 2)
		assert.Contains(t, allWidgets, "widget1")
		assert.Contains(t, allWidgets, "widget2")
	})

	t.Run("remove widget", func(t *testing.T) {
		// Add a widget
		err := manager.CreateWidget("remove_me", "clock", nil, nil)
		require.NoError(t, err)

		// Verify it exists
		_, exists := manager.GetWidget("remove_me")
		assert.True(t, exists)

		// Remove it
		manager.RemoveWidget("remove_me")

		// Verify it's gone
		_, exists = manager.GetWidget("remove_me")
		assert.False(t, exists)
	})

	t.Run("get factory", func(t *testing.T) {
		retrievedFactory := manager.GetFactory()
		assert.Equal(t, factory, retrievedFactory)
	})

	t.Run("create widget with invalid type", func(t *testing.T) {
		err := manager.CreateWidget("invalid_widget", "nonexistent_type", nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create widget")

		// Widget should not be stored
		_, exists := manager.GetWidget("invalid_widget")
		assert.False(t, exists)
	})
}

func TestBaseWidget(t *testing.T) {
	widget := &BaseWidget{
		ID:   "test_widget",
		Type: "test_type",
		Data: "test_data",
	}

	t.Run("getters work correctly", func(t *testing.T) {
		assert.Equal(t, "test_widget", widget.GetID())
		assert.Equal(t, "test_type", widget.GetType())
		assert.Equal(t, "test_data", widget.GetData())
		assert.Empty(t, widget.GetChildren()) // Should be empty by default
	})

	t.Run("init and close work", func(t *testing.T) {
		ctx := context.Background()

		err := widget.Init(ctx)
		assert.NoError(t, err) // BaseWidget Init should not error

		err = widget.Close()
		assert.NoError(t, err) // BaseWidget Close should not error
	})

	t.Run("children management", func(t *testing.T) {
		child1 := &BaseWidget{ID: "child1", Type: "test"}
		child2 := &BaseWidget{ID: "child2", Type: "test"}

		widget.Children = []Widget{child1, child2}

		children := widget.GetChildren()
		assert.Len(t, children, 2)
		assert.Equal(t, "child1", children[0].GetID())
		assert.Equal(t, "child2", children[1].GetID())
	})
}

func TestWidgetInterfaces(t *testing.T) {
	var widget Widget = &BaseWidget{}
	assert.Implements(t, (*Widget)(nil), &BaseWidget{})

	// Test basic interface detection
	_, isContainer := widget.(Container)
	assert.False(t, isContainer)
}

// Test that all built-in widget types can be created successfully.
func TestBuiltinWidgetCreation(t *testing.T) {
	provider := &MockProvider{}
	factory := NewDefaultWidgetFactory(provider)

	testCases := []struct {
		name       string
		widgetType string
		config     map[string]interface{}
		shouldWork bool
	}{
		{
			name:       "clock widget",
			widgetType: "clock",
			config:     map[string]interface{}{"format": "15:04:05"},
			shouldWork: true,
		},
		{
			name:       "horizontal split",
			widgetType: "horizontal_split",
			config:     nil,
			shouldWork: true,
		},
		{
			name:       "vertical split",
			widgetType: "vertical_split",
			config:     nil,
			shouldWork: true,
		},
		{
			name:       "grow widget",
			widgetType: "grow",
			config:     map[string]interface{}{"grow": "2"},
			shouldWork: true,
		},
		{
			name:       "home assistant entity",
			widgetType: "home_assistant.entity",
			config:     map[string]interface{}{"entity_id": "sensor.test"},
			shouldWork: true,
		},
		{
			name:       "home assistant button",
			widgetType: "home_assistant.button",
			config: map[string]interface{}{
				"entity_id": "button.test",
				"service":   "press",
				"domain":    "button",
			},
			shouldWork: true,
		},
		{
			name:       "home assistant switch",
			widgetType: "home_assistant.switch",
			config:     map[string]interface{}{"entity_id": "switch.test"},
			shouldWork: true,
		},
		{
			name:       "home assistant light",
			widgetType: "home_assistant.light",
			config:     map[string]interface{}{"entity_id": "light.test"},
			shouldWork: true,
		},
		{
			name:       "dexcom widget",
			widgetType: "dexcom",
			config: map[string]interface{}{
				"low_threshold":  70,
				"high_threshold": 180,
			},
			shouldWork: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			widget, err := factory.Create(tc.widgetType, "test_"+tc.widgetType, configToNode(tc.config), nil)

			if tc.shouldWork {
				assert.NoError(t, err, "Should be able to create %s widget", tc.widgetType)
				assert.NotNil(t, widget)
				assert.Equal(t, tc.widgetType, widget.GetType())
				assert.Equal(t, "test_"+tc.widgetType, widget.GetID())
			} else {
				assert.Error(t, err)
				assert.Nil(t, widget)
			}
		})
	}
}

// Benchmark tests for performance.
func BenchmarkWidgetCreation(b *testing.B) {
	provider := &MockProvider{}
	factory := NewDefaultWidgetFactory(provider)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widget, err := factory.Create("clock", "bench_clock", configToNode(map[string]interface{}{
			"format": "15:04:05",
		}), nil)
		if err != nil {
			b.Fatal(err)
		}
		if widget == nil {
			b.Fatal("widget is nil")
		}
	}
}
