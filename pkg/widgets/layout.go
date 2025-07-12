package widgets

import (
	"context"
	"fmt"
)

type SplitWidget struct {
	*BaseWidget
	Direction string
	Sizes     []float64
}


type LayoutData struct {
	Type      string    `json:"type"`
	Direction string    `json:"direction,omitempty"`
	Sizes     []float64 `json:"sizes,omitempty"`
	Children  []Widget  `json:"children,omitempty"`
}

func CreateHorizontalSplitWidget(config map[string]interface{}) (Widget, error) {
	return createSplitWidget(config, "horizontal_split", "horizontal")
}

func CreateVerticalSplitWidget(config map[string]interface{}) (Widget, error) {
	return createSplitWidget(config, "vertical_split", "vertical")
}

func createSplitWidget(config map[string]interface{}, widgetType, direction string) (Widget, error) {
	var splitConfig SplitConfig
	parser := NewConfigParser()
	if err := parser.ParseConfig(config, &splitConfig); err != nil {
		return nil, fmt.Errorf("invalid split configuration: %w", err)
	}
	
	widget := &SplitWidget{
		BaseWidget: &BaseWidget{
			ID:       generateWidgetID(),
			Type:     widgetType,
			Config:   config,
			Children: []Widget{},
		},
		Direction: direction,
		Sizes:     splitConfig.Sizes,
	}
	
	return widget, nil
}


func (w *SplitWidget) Update(ctx context.Context) error {
	w.Data = &LayoutData{
		Type:      w.Type,
		Direction: w.Direction,
		Sizes:     w.Sizes,
		Children:  w.Children,
	}
	return nil
}


func (w *SplitWidget) IsContainer() bool {
	return true
}


func (w *SplitWidget) Configure(config map[string]interface{}) error {
	w.BaseWidget.Configure(config)
	
	if sizesConfig, ok := config["sizes"].([]interface{}); ok {
		sizes := []float64{}
		for _, size := range sizesConfig {
			if s, ok := size.(float64); ok {
				sizes = append(sizes, s)
			}
		}
		w.Sizes = sizes
	}
	
	return nil
}

