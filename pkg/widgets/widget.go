package widgets

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/mntndev/dash/pkg/integrations"
)

// Provider interface combines all the services widgets might need.
type Provider interface {
	GetHAClient() *integrations.HomeAssistantClient
	GetDexcomClient() *integrations.DexcomClient
}

type Widget interface {
	GetID() string
	GetType() string
	Init(ctx context.Context) error
	GetChildren() []Widget
	Close() error
	Layout(gtx layout.Context) layout.Dimensions
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
	LastUpdate time.Time   `json:"LastUpdate"`
	Children   []Widget    `json:"Children"`
	window     *app.Window `json:"-"`
}

func (w *BaseWidget) GetID() string {
	return w.ID
}

func (w *BaseWidget) GetType() string {
	return w.Type
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

func (w *BaseWidget) Layout(gtx layout.Context) layout.Dimensions {
	// Default implementation: render as unknown widget
	text := fmt.Sprintf("Unknown widget: %s", w.GetType())
	th := material.NewTheme()
	label := material.Body1(th, text)
	label.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	return label.Layout(gtx)
}

func (w *BaseWidget) Invalidate() {
	if w.window != nil {
		w.window.Invalidate()
	}
}

type WidgetCreator func(id string, config ast.Node, children []Widget, provider Provider, window *app.Window) (Widget, error)

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

func (wr *WidgetRegistry) Create(widgetType, id string, config ast.Node, children []Widget, provider Provider, window *app.Window) (Widget, error) {
	creator, exists := wr.creators[widgetType]
	if !exists {
		return nil, fmt.Errorf("unsupported widget type: %s", widgetType)
	}
	return creator(id, config, children, provider, window)
}

func (wr *WidgetRegistry) GetSupportedTypes() []string {
	types := make([]string, 0, len(wr.creators))
	for widgetType := range wr.creators {
		types = append(types, widgetType)
	}
	return types
}

type WidgetFactory interface {
	Create(widgetType string, id string, config ast.Node, children []Widget, window *app.Window) (Widget, error)
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

	// New layout widgets
	registry.Register("hstack", CreateHStackWidget)
	registry.Register("vstack", CreateVStackWidget)
	registry.Register("hflex", CreateHFlexWidget)
	registry.Register("vflex", CreateVFlexWidget)

	registry.Register("grow", func(id string, config ast.Node, children []Widget, provider Provider, window *app.Window) (Widget, error) {
		return CreateGrowWidgetWithWindow(id, config, children, window)
	})
}

func (f *DefaultWidgetFactory) Create(widgetType, id string, config ast.Node, children []Widget, window *app.Window) (Widget, error) {
	return f.registry.Create(widgetType, id, config, children, f.provider, window)
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

func (wm *WidgetManager) CreateWidget(id, widgetType string, config ast.Node, children []Widget, window *app.Window) error {
	widget, err := wm.factory.Create(widgetType, id, config, children, window)
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
	return createGrowWidget(id, config, children, nil)
}

func CreateGrowWidgetWithWindow(id string, config ast.Node, children []Widget, window *app.Window) (Widget, error) {
	return createGrowWidget(id, config, children, window)
}

func createGrowWidget(id string, config ast.Node, children []Widget, window *app.Window) (Widget, error) {
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
			window:   window,
		},
		GrowValue: growValue,
	}

	return widget, nil
}

func (w *GrowWidget) Init(ctx context.Context) error {
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
	w.LastUpdate = time.Now()
}

func (w *GrowWidget) Close() error {
	return nil
}

func (w *GrowWidget) Layout(gtx layout.Context) layout.Dimensions {
	children := w.GetChildren()
	if len(children) == 0 {
		return layout.Dimensions{}
	}
	// Grow widget just renders its first child
	return children[0].Layout(gtx)
}
