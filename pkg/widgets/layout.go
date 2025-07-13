package widgets

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type SplitConfig struct {
	Sizes []float64 `yaml:"sizes"`
}

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

func CreateHorizontalSplitWidget(id string, config ast.Node, children []Widget) (Widget, error) {
	return createSplitWidget(id, config, children, "horizontal_split", "horizontal")
}

func CreateVerticalSplitWidget(id string, config ast.Node, children []Widget) (Widget, error) {
	return createSplitWidget(id, config, children, "vertical_split", "vertical")
}

func createSplitWidget(id string, config ast.Node, children []Widget, widgetType, direction string) (Widget, error) {
	// Parse config using NodeToValue
	var splitConfig SplitConfig
	if config != nil {
		if err := yaml.NodeToValue(config, &splitConfig); err != nil {
			return nil, fmt.Errorf("failed to parse split config: %w", err)
		}
	}

	sizes := splitConfig.Sizes

	widget := &SplitWidget{
		BaseWidget: &BaseWidget{
			ID:       id,
			Type:     widgetType,
			Config:   config,
			Children: children,
		},
		Direction: direction,
		Sizes:     sizes,
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
