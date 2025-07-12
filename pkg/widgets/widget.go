package widgets

import (
	"context"
	"fmt"
	"time"

	"github.com/mntndev/dash/pkg/integrations"
)

type Widget interface {
	GetID() string
	GetType() string
	GetData() interface{}
	Update(ctx context.Context) error
	ShouldUpdate() bool
	Configure(config map[string]interface{}) error
	SetID(id string)
	GetChildren() []Widget
	SetChildren(children []Widget)
	IsContainer() bool
}

// Capability interfaces for widgets
type Triggerable interface {
	Trigger() error
}

type BrightnessControllable interface {
	SetBrightness(brightness int) error
}

type Configurable interface {
	Configure(config map[string]interface{}) error
}

type Container interface {
	IsContainer() bool
	GetChildren() []Widget
	SetChildren(children []Widget)
}

type BaseWidget struct {
	ID       string
	Type     string
	Config   map[string]interface{}
	Data     interface{}
	LastUpdate time.Time
	Children []Widget
}

func (w *BaseWidget) GetID() string {
	return w.ID
}

func (w *BaseWidget) SetID(id string) {
	w.ID = id
}

func (w *BaseWidget) GetType() string {
	return w.Type
}


func (w *BaseWidget) GetData() interface{} {
	return w.Data
}

func (w *BaseWidget) Configure(config map[string]interface{}) error {
	w.Config = config
	return nil
}

func (w *BaseWidget) GetChildren() []Widget {
	return w.Children
}

func (w *BaseWidget) SetChildren(children []Widget) {
	w.Children = children
}

func (w *BaseWidget) IsContainer() bool {
	return len(w.Children) > 0
}

func (w *BaseWidget) Update(ctx context.Context) error {
	return nil
}

func (w *BaseWidget) ShouldUpdate() bool {
	// Default implementation: update every time
	return true
}

type WidgetCreator func(config map[string]interface{}, haProvider integrations.HAProvider, dexcomProvider integrations.DexcomProvider) (Widget, error)

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

func (wr *WidgetRegistry) Create(widgetType string, config map[string]interface{}, haProvider integrations.HAProvider, dexcomProvider integrations.DexcomProvider) (Widget, error) {
	creator, exists := wr.creators[widgetType]
	if !exists {
		return nil, fmt.Errorf("unsupported widget type: %s", widgetType)
	}
	return creator(config, haProvider, dexcomProvider)
}

func (wr *WidgetRegistry) GetSupportedTypes() []string {
	types := make([]string, 0, len(wr.creators))
	for widgetType := range wr.creators {
		types = append(types, widgetType)
	}
	return types
}

type WidgetFactory interface {
	Create(widgetType string, config map[string]interface{}) (Widget, error)
	GetSupportedTypes() []string
}

type DefaultWidgetFactory struct {
	haProvider     integrations.HAProvider
	dexcomProvider integrations.DexcomProvider
	registry       *WidgetRegistry
}

func NewDefaultWidgetFactory(haProvider integrations.HAProvider, dexcomProvider integrations.DexcomProvider) *DefaultWidgetFactory {
	registry := NewWidgetRegistry()
	registerBuiltinWidgets(registry)
	
	return &DefaultWidgetFactory{
		haProvider:     haProvider,
		dexcomProvider: dexcomProvider,
		registry:       registry,
	}
}

func registerBuiltinWidgets(registry *WidgetRegistry) {
	registry.Register("home_assistant.entity", func(config map[string]interface{}, haProvider integrations.HAProvider, dexcomProvider integrations.DexcomProvider) (Widget, error) {
		return CreateHAEntityWidget(config, haProvider)
	})
	
	registry.Register("home_assistant.button", func(config map[string]interface{}, haProvider integrations.HAProvider, dexcomProvider integrations.DexcomProvider) (Widget, error) {
		return CreateHAButtonWidget(config, haProvider)
	})
	
	registry.Register("home_assistant.switch", func(config map[string]interface{}, haProvider integrations.HAProvider, dexcomProvider integrations.DexcomProvider) (Widget, error) {
		return CreateHASwitchWidget(config, haProvider)
	})
	
	registry.Register("home_assistant.light", func(config map[string]interface{}, haProvider integrations.HAProvider, dexcomProvider integrations.DexcomProvider) (Widget, error) {
		return CreateHALightWidget(config, haProvider)
	})
	
	registry.Register("dexcom", func(config map[string]interface{}, haProvider integrations.HAProvider, dexcomProvider integrations.DexcomProvider) (Widget, error) {
		return CreateDexcomWidget(config, dexcomProvider)
	})
	
	registry.Register("clock", func(config map[string]interface{}, haProvider integrations.HAProvider, dexcomProvider integrations.DexcomProvider) (Widget, error) {
		return CreateClockWidget(config)
	})
	
	registry.Register("horizontal_split", func(config map[string]interface{}, haProvider integrations.HAProvider, dexcomProvider integrations.DexcomProvider) (Widget, error) {
		return CreateHorizontalSplitWidget(config)
	})
	
	registry.Register("vertical_split", func(config map[string]interface{}, haProvider integrations.HAProvider, dexcomProvider integrations.DexcomProvider) (Widget, error) {
		return CreateVerticalSplitWidget(config)
	})
	
	registry.Register("grow", func(config map[string]interface{}, haProvider integrations.HAProvider, dexcomProvider integrations.DexcomProvider) (Widget, error) {
		return CreateGrowWidget(config)
	})
}


func (f *DefaultWidgetFactory) Create(widgetType string, config map[string]interface{}) (Widget, error) {
	return f.registry.Create(widgetType, config, f.haProvider, f.dexcomProvider)
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

func (wm *WidgetManager) CreateWidget(id string, widgetType string, config map[string]interface{}) error {
	widget, err := wm.factory.Create(widgetType, config)
	if err != nil {
		return fmt.Errorf("failed to create widget %s: %w", id, err)
	}
	
	widget.SetID(id)
	
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

func (wm *WidgetManager) RemoveWidget(id string) {
	delete(wm.widgets, id)
}

func (wm *WidgetManager) UpdateAll(ctx context.Context) error {
	for id, widget := range wm.widgets {
		// Only update widgets that need updating (reactive updates)
		if widget.ShouldUpdate() {
			if err := widget.Update(ctx); err != nil {
				return fmt.Errorf("failed to update widget %s: %w", id, err)
			}
		}
	}
	return nil
}

type GrowWidget struct {
	*BaseWidget
}

func CreateGrowWidget(config map[string]interface{}) (Widget, error) {
	widget := &GrowWidget{
		BaseWidget: &BaseWidget{
			ID:       generateWidgetID(),
			Type:     "grow",
			Config:   config,
			Children: []Widget{},
		},
	}
	
	return widget, nil
}

func (w *GrowWidget) Update(ctx context.Context) error {
	w.LastUpdate = time.Now()
	return nil
}

func (w *GrowWidget) IsContainer() bool {
	return len(w.Children) > 0
}