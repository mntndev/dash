package widgets

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/mntndev/dash/pkg/integrations"
)

// Provider interface combines all the services widgets might need.
type Provider interface {
	GetHAClient() *integrations.HomeAssistantClient
	GetDexcomClient() *integrations.DexcomClient
	Emit(event string, data interface{})
	IsFrontendReady() bool
}

type Widget interface {
	GetID() string
	GetType() string
	GetData() interface{}
	Init(ctx context.Context) error
	GetChildren() []Widget
	Close() error
}

// Triggerable interface defines widgets that can be manually triggered.
type Triggerable interface {
	Trigger() error
}

type BrightnessControllable interface {
	SetBrightness(brightness int) error
}

type Configurable interface {
	Configure(config ast.Node) error
}

type Container interface {
	IsContainer() bool
	GetChildren() []Widget
	SetChildren(children []Widget)
}

type BaseWidget struct {
	ID         string      `json:"ID"`
	Type       string      `json:"Type"`
	Config     ast.Node    `json:"-"`
	Data       interface{} `json:"Data"`
	LastUpdate time.Time   `json:"LastUpdate"`
	Children   []Widget    `json:"Children"`
}

func (w *BaseWidget) GetID() string {
	return w.ID
}

func (w *BaseWidget) GetType() string {
	return w.Type
}

func (w *BaseWidget) GetData() interface{} {
	return w.Data
}

func (w *BaseWidget) GetChildren() []Widget {
	return w.Children
}

func (w *BaseWidget) Init(ctx context.Context) error {
	return nil
}

func (w *BaseWidget) Close() error {
	// Default implementation: no cleanup needed
	return nil
}

type WidgetCreator func(id string, config ast.Node, children []Widget, provider Provider) (Widget, error)

type WidgetRegistry struct {
	creators map[string]WidgetCreator
}

func NewWidgetRegistry() *WidgetRegistry {
	return &WidgetRegistry{
		creators: make(map[string]WidgetCreator),
	}
}

func (wr *WidgetRegistry) Register(widgetType string, creator WidgetCreator) {
	wr.creators[widgetType] = creator
}

func (wr *WidgetRegistry) Create(widgetType, id string, config ast.Node, children []Widget, provider Provider) (Widget, error) {
	creator, exists := wr.creators[widgetType]
	if !exists {
		return nil, fmt.Errorf("unsupported widget type: %s", widgetType)
	}
	return creator(id, config, children, provider)
}

func (wr *WidgetRegistry) GetSupportedTypes() []string {
	types := make([]string, 0, len(wr.creators))
	for widgetType := range wr.creators {
		types = append(types, widgetType)
	}
	return types
}

type WidgetFactory interface {
	Create(widgetType string, id string, config ast.Node, children []Widget) (Widget, error)
	GetSupportedTypes() []string
}

type DefaultWidgetFactory struct {
	provider Provider
	registry *WidgetRegistry
}

func NewDefaultWidgetFactory(provider Provider) *DefaultWidgetFactory {
	registry := NewWidgetRegistry()
	registerBuiltinWidgets(registry)

	return &DefaultWidgetFactory{
		provider: provider,
		registry: registry,
	}
}

func registerBuiltinWidgets(registry *WidgetRegistry) {
	registry.Register("home_assistant.entity", CreateHAEntityWidget)

	registry.Register("home_assistant.button", CreateHAButtonWidget)

	registry.Register("home_assistant.switch", CreateHASwitchWidget)

	registry.Register("home_assistant.light", CreateHALightWidget)

	registry.Register("dexcom", CreateDexcomWidget)

	registry.Register("clock", CreateClockWidget)

	registry.Register("horizontal_split", func(id string, config ast.Node, children []Widget, provider Provider) (Widget, error) {
		return CreateHorizontalSplitWidget(id, config, children)
	})

	registry.Register("vertical_split", func(id string, config ast.Node, children []Widget, provider Provider) (Widget, error) {
		return CreateVerticalSplitWidget(id, config, children)
	})

	registry.Register("grow", func(id string, config ast.Node, children []Widget, provider Provider) (Widget, error) {
		return CreateGrowWidget(id, config, children)
	})
}

func (f *DefaultWidgetFactory) Create(widgetType, id string, config ast.Node, children []Widget) (Widget, error) {
	return f.registry.Create(widgetType, id, config, children, f.provider)
}

func (f *DefaultWidgetFactory) GetSupportedTypes() []string {
	return f.registry.GetSupportedTypes()
}

type WidgetManager struct {
	widgets map[string]Widget
	factory WidgetFactory
}

func NewWidgetManager(factory WidgetFactory) *WidgetManager {
	return &WidgetManager{
		widgets: make(map[string]Widget),
		factory: factory,
	}
}

func (wm *WidgetManager) CreateWidget(id, widgetType string, config ast.Node, children []Widget) error {
	widget, err := wm.factory.Create(widgetType, id, config, children)
	if err != nil {
		return fmt.Errorf("failed to create widget %s: %w", id, err)
	}

	wm.widgets[id] = widget
	return nil
}

func (wm *WidgetManager) GetWidget(id string) (Widget, bool) {
	widget, exists := wm.widgets[id]
	return widget, exists
}

func (wm *WidgetManager) GetAllWidgets() map[string]Widget {
	return wm.widgets
}

func (wm *WidgetManager) StoreWidget(id string, widget Widget) {
	wm.widgets[id] = widget
}

func (wm *WidgetManager) GetFactory() WidgetFactory {
	return wm.factory
}

func (wm *WidgetManager) RemoveWidget(id string) {
	if widget, exists := wm.widgets[id]; exists {
		if err := widget.Close(); err != nil {
			log.Printf("Failed to close widget %s: %v", id, err)
		}
		delete(wm.widgets, id)
	}
}

type GrowWidget struct {
	*BaseWidget
	GrowValue string
}

// GrowConfig represents configuration for grow widgets.
type GrowConfig struct {
	Grow interface{} `yaml:"grow"`
}

func CreateGrowWidget(id string, config ast.Node, children []Widget) (Widget, error) {
	// Parse grow value from config using NodeToValue
	var growConfig GrowConfig
	if config != nil {
		if err := yaml.NodeToValue(config, &growConfig); err != nil {
			return nil, fmt.Errorf("failed to parse grow config: %w", err)
		}
	}

	// Convert grow value to string, default to "1"
	growValue := "1"
	if growConfig.Grow != nil {
		switch val := growConfig.Grow.(type) {
		case string:
			growValue = strings.Trim(val, `"`)
		case float64:
			growValue = fmt.Sprintf("%.0f", val)
		case int:
			growValue = fmt.Sprintf("%d", val)
		case int64:
			growValue = fmt.Sprintf("%d", val)
		case uint64:
			growValue = fmt.Sprintf("%d", val)
		default:
			// If we can't handle the type, fallback to default
			growValue = "1"
		}
	}

	widget := &GrowWidget{
		BaseWidget: &BaseWidget{
			ID:       id,
			Type:     "grow",
			Config:   config,
			Children: children,
		},
		GrowValue: growValue,
	}

	return widget, nil
}

func (w *GrowWidget) Init(ctx context.Context) error {
	w.Data = map[string]interface{}{
		"type":       w.Type,
		"grow_value": w.GrowValue,
	}
	w.LastUpdate = time.Now()
	return nil
}

func (w *GrowWidget) GetGrowValue() string {
	return w.GrowValue
}

func (w *GrowWidget) IsContainer() bool {
	return true // Grow widgets can always contain children
}

func (w *GrowWidget) SetChildren(children []Widget) {
	w.Children = children
	// Update data timestamp when children are set
	w.Data = map[string]interface{}{
		"type":       w.Type,
		"grow_value": w.GrowValue,
	}
	w.LastUpdate = time.Now()
}

func (w *GrowWidget) Close() error {
	return nil
}
