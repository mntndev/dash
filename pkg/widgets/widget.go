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
	Configure(config map[string]interface{}) error
	SetID(id string)
	GetChildren() []Widget
	SetChildren(children []Widget)
	IsContainer() bool
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

type WidgetFactory interface {
	Create(widgetType string, config map[string]interface{}) (Widget, error)
	GetSupportedTypes() []string
}

type DefaultWidgetFactory struct {
	haProvider     integrations.HAProvider
	dexcomProvider integrations.DexcomProvider
}

func NewDefaultWidgetFactory(haProvider integrations.HAProvider, dexcomProvider integrations.DexcomProvider) *DefaultWidgetFactory {
	return &DefaultWidgetFactory{
		haProvider:     haProvider,
		dexcomProvider: dexcomProvider,
	}
}


func (f *DefaultWidgetFactory) Create(widgetType string, config map[string]interface{}) (Widget, error) {
	switch widgetType {
	case "home_assistant.entity":
		return CreateHAEntityWidget(config, f.haProvider)
	case "home_assistant.button":
		return CreateHAButtonWidget(config, f.haProvider)
	case "home_assistant.switch":
		return CreateHASwitchWidget(config, f.haProvider)
	case "home_assistant.light":
		return CreateHALightWidget(config, f.haProvider)
	case "dexcom":
		return CreateDexcomWidget(config, f.dexcomProvider)
	case "clock":
		return CreateClockWidget(config)
	case "horizontal_split":
		return CreateHorizontalSplitWidget(config)
	case "vertical_split":
		return CreateVerticalSplitWidget(config)
	case "grow":
		return CreateGrowWidget(config)
	default:
		return nil, fmt.Errorf("unsupported widget type: %s", widgetType)
	}
}

func (f *DefaultWidgetFactory) GetSupportedTypes() []string {
	return []string{
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
		if err := widget.Update(ctx); err != nil {
			return fmt.Errorf("failed to update widget %s: %w", id, err)
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