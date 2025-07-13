package widgets

import (
	"context"
	"fmt"
	"log"
	"time"
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
}

func CreateHorizontalSplitWidget(id string, config map[string]interface{}, children []Widget) (Widget, error) {
	return createSplitWidget(id, config, children, "horizontal_split", "horizontal")
}

func CreateVerticalSplitWidget(id string, config map[string]interface{}, children []Widget) (Widget, error) {
	return createSplitWidget(id, config, children, "vertical_split", "vertical")
}

func createSplitWidget(id string, config map[string]interface{}, children []Widget, widgetType, direction string) (Widget, error) {
	var splitConfig SplitConfig
	parser := NewConfigParser()
	if err := parser.ParseConfig(config, &splitConfig); err != nil {
		return nil, fmt.Errorf("invalid split configuration: %w", err)
	}

	widget := &SplitWidget{
		BaseWidget: &BaseWidget{
			ID:       id,
			Type:     widgetType,
			Config:   config,
			Children: children,
		},
		Direction: direction,
		Sizes:     splitConfig.Sizes,
	}

	return widget, nil
}

func (w *SplitWidget) Init(ctx context.Context) error {
	w.Data = &LayoutData{
		Type:      w.Type,
		Direction: w.Direction,
		Sizes:     w.Sizes,
	}
	w.LastUpdate = time.Now()
	return nil
}

func (w *SplitWidget) IsContainer() bool {
	return true
}

func (w *SplitWidget) SetChildren(children []Widget) {
	w.Children = children
	// Update data timestamp when children are set
	w.Data = &LayoutData{
		Type:      w.Type,
		Direction: w.Direction,
		Sizes:     w.Sizes,
	}
	w.LastUpdate = time.Now()
}

func (w *SplitWidget) Close() error {
	for _, child := range w.Children {
		if err := child.Close(); err != nil {
			log.Printf("Failed to close child widget: %v", err)
		}
	}
	return nil
}
