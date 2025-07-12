package widgets

import (
	"context"
	"fmt"
	"time"
)

type Widget interface {
	GetID() string
	GetType() string
	GetData() interface{}
	Update(ctx context.Context) error
	Configure(config map[string]interface{}) error
	SetID(id string)
	// Layout support
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
	// Base implementation does nothing
	return nil
}

type WidgetFactory interface {
	Create(widgetType string, config map[string]interface{}) (Widget, error)
	GetSupportedTypes() []string
}

type DefaultWidgetFactory struct {
	creators map[string]func(map[string]interface{}) (Widget, error)
}

func NewDefaultWidgetFactory() *DefaultWidgetFactory {
	factory := &DefaultWidgetFactory{
		creators: make(map[string]func(map[string]interface{}) (Widget, error)),
	}
	
	factory.RegisterCreator("home_assistant.entity", CreateHAEntityWidget)
	factory.RegisterCreator("home_assistant.button", CreateHAButtonWidget)
	factory.RegisterCreator("home_assistant.switch", CreateHASwitchWidget)
	factory.RegisterCreator("clock", CreateClockWidget)
	factory.RegisterCreator("horizontal_split", CreateHorizontalSplitWidget)
	factory.RegisterCreator("vertical_split", CreateVerticalSplitWidget)
	factory.RegisterCreator("grow", CreateGrowWidget)
	
	return factory
}

func (f *DefaultWidgetFactory) RegisterCreator(widgetType string, creator func(map[string]interface{}) (Widget, error)) {
	f.creators[widgetType] = creator
}

func (f *DefaultWidgetFactory) Create(widgetType string, config map[string]interface{}) (Widget, error) {
	creator, exists := f.creators[widgetType]
	if !exists {
		return nil, fmt.Errorf("unsupported widget type: %s", widgetType)
	}
	
	return creator(config)
}

func (f *DefaultWidgetFactory) GetSupportedTypes() []string {
	types := make([]string, 0, len(f.creators))
	for t := range f.creators {
		types = append(types, t)
	}
	return types
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
	
	// Set the widget's internal ID to match the key used in the manager
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

// GrowWidget - A widget that applies flex-grow behavior to itself or its child
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