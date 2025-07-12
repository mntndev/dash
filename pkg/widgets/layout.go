package widgets

import (
	"context"
)

// HorizontalSplitWidget represents a horizontal split layout container
type HorizontalSplitWidget struct {
	*BaseWidget
	Sizes []float64 // Relative sizes for each child
}

// VerticalSplitWidget represents a vertical split layout container
type VerticalSplitWidget struct {
	*BaseWidget
	Sizes []float64 // Relative sizes for each child
}


type LayoutData struct {
	Type     string    `json:"type"`
	Sizes    []float64 `json:"sizes,omitempty"`
	Children []Widget  `json:"children,omitempty"`
}

func CreateHorizontalSplitWidget(config map[string]interface{}) (Widget, error) {
	sizes := []float64{}
	if sizesConfig, ok := config["sizes"].([]interface{}); ok {
		for _, size := range sizesConfig {
			if s, ok := size.(float64); ok {
				sizes = append(sizes, s)
			}
		}
	}
	
	widget := &HorizontalSplitWidget{
		BaseWidget: &BaseWidget{
			ID:       generateWidgetID(),
			Type:     "horizontal_split",
			Config:   config,
			Children: []Widget{},
		},
		Sizes: sizes,
	}
	
	return widget, nil
}

func CreateVerticalSplitWidget(config map[string]interface{}) (Widget, error) {
	sizes := []float64{}
	if sizesConfig, ok := config["sizes"].([]interface{}); ok {
		for _, size := range sizesConfig {
			if s, ok := size.(float64); ok {
				sizes = append(sizes, s)
			}
		}
	}
	
	widget := &VerticalSplitWidget{
		BaseWidget: &BaseWidget{
			ID:       generateWidgetID(),
			Type:     "vertical_split",
			Config:   config,
			Children: []Widget{},
		},
		Sizes: sizes,
	}
	
	return widget, nil
}


func (w *HorizontalSplitWidget) Update(ctx context.Context) error {
	w.Data = &LayoutData{
		Type:     "horizontal_split",
		Sizes:    w.Sizes,
		Children: w.Children,
	}
	return nil
}

func (w *VerticalSplitWidget) Update(ctx context.Context) error {
	w.Data = &LayoutData{
		Type:     "vertical_split",
		Sizes:    w.Sizes,
		Children: w.Children,
	}
	return nil
}


func (w *HorizontalSplitWidget) IsContainer() bool {
	return true
}

func (w *VerticalSplitWidget) IsContainer() bool {
	return true
}


func (w *HorizontalSplitWidget) Configure(config map[string]interface{}) error {
	w.BaseWidget.Configure(config)
	
	// Update sizes from config
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

func (w *VerticalSplitWidget) Configure(config map[string]interface{}) error {
	w.BaseWidget.Configure(config)
	
	// Update sizes from config
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

