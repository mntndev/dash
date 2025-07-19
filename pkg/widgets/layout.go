package widgets

import (
	"context"
	"log"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"github.com/goccy/go-yaml/ast"
)

// HStack widget - horizontal layout using Flex with Rigid children
type HStackWidget struct {
	*BaseWidget
}

// VStack widget - vertical layout using Flex with Rigid children
type VStackWidget struct {
	*BaseWidget
}

// HFlex widget - horizontal flex layout with Flexed children
type HFlexWidget struct {
	*BaseWidget
}

// VFlex widget - vertical flex layout with Flexed children
type VFlexWidget struct {
	*BaseWidget
}

func CreateHStackWidget(id string, config ast.Node, children []Widget, provider Provider, window *app.Window) (Widget, error) {
	widget := &HStackWidget{
		BaseWidget: &BaseWidget{
			ID:       id,
			Type:     "hstack",
			Config:   config,
			Children: children,
			window:   window,
		},
	}
	return widget, nil
}

func CreateVStackWidget(id string, config ast.Node, children []Widget, provider Provider, window *app.Window) (Widget, error) {
	widget := &VStackWidget{
		BaseWidget: &BaseWidget{
			ID:       id,
			Type:     "vstack",
			Config:   config,
			Children: children,
			window:   window,
		},
	}
	return widget, nil
}

func CreateHFlexWidget(id string, config ast.Node, children []Widget, provider Provider, window *app.Window) (Widget, error) {
	widget := &HFlexWidget{
		BaseWidget: &BaseWidget{
			ID:       id,
			Type:     "hflex",
			Config:   config,
			Children: children,
			window:   window,
		},
	}
	return widget, nil
}

func CreateVFlexWidget(id string, config ast.Node, children []Widget, provider Provider, window *app.Window) (Widget, error) {
	widget := &VFlexWidget{
		BaseWidget: &BaseWidget{
			ID:       id,
			Type:     "vflex",
			Config:   config,
			Children: children,
			window:   window,
		},
	}
	return widget, nil
}

func (w *HStackWidget) Init(ctx context.Context) error {
	w.LastUpdate = time.Now()
	return nil
}

func (w *HStackWidget) IsContainer() bool {
	return true
}

func (w *HStackWidget) SetChildren(children []Widget) {
	w.Children = children
	w.LastUpdate = time.Now()
}

func (w *HStackWidget) Close() error {
	for _, child := range w.Children {
		if err := child.Close(); err != nil {
			log.Printf("Failed to close child widget: %v", err)
		}
	}
	return nil
}

func (w *HStackWidget) Layout(gtx layout.Context) layout.Dimensions {
	children := w.GetChildren()
	if len(children) == 0 {
		return layout.Dimensions{}
	}

	// Use Flex with Rigid children for horizontal stacking
	var flexChildren []layout.FlexChild
	for _, child := range children {
		flexChildren = append(flexChildren, layout.Rigid(child.Layout))
	}

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, flexChildren...)
}

func (w *VStackWidget) Init(ctx context.Context) error {
	w.LastUpdate = time.Now()
	return nil
}

func (w *VStackWidget) IsContainer() bool {
	return true
}

func (w *VStackWidget) SetChildren(children []Widget) {
	w.Children = children
	w.LastUpdate = time.Now()
}

func (w *VStackWidget) Close() error {
	for _, child := range w.Children {
		if err := child.Close(); err != nil {
			log.Printf("Failed to close child widget: %v", err)
		}
	}
	return nil
}

func (w *VStackWidget) Layout(gtx layout.Context) layout.Dimensions {
	children := w.GetChildren()
	if len(children) == 0 {
		return layout.Dimensions{}
	}

	// Use Flex with Rigid children for vertical stacking
	var flexChildren []layout.FlexChild
	for _, child := range children {
		flexChildren = append(flexChildren, layout.Rigid(child.Layout))
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, flexChildren...)
}

func (w *HFlexWidget) Init(ctx context.Context) error {
	w.LastUpdate = time.Now()
	return nil
}

func (w *HFlexWidget) IsContainer() bool {
	return true
}

func (w *HFlexWidget) SetChildren(children []Widget) {
	w.Children = children
	w.LastUpdate = time.Now()
}

func (w *HFlexWidget) Close() error {
	for _, child := range w.Children {
		if err := child.Close(); err != nil {
			log.Printf("Failed to close child widget: %v", err)
		}
	}
	return nil
}

func (w *HFlexWidget) Layout(gtx layout.Context) layout.Dimensions {
	children := w.GetChildren()
	if len(children) == 0 {
		return layout.Dimensions{}
	}

	// Use Flex for equal distribution horizontally
	var flexChildren []layout.FlexChild
	for _, child := range children {
		flexChildren = append(flexChildren, layout.Flexed(1, child.Layout))
	}

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, flexChildren...)
}

func (w *VFlexWidget) Init(ctx context.Context) error {
	w.LastUpdate = time.Now()
	return nil
}

func (w *VFlexWidget) IsContainer() bool {
	return true
}

func (w *VFlexWidget) SetChildren(children []Widget) {
	w.Children = children
	w.LastUpdate = time.Now()
}

func (w *VFlexWidget) Close() error {
	for _, child := range w.Children {
		if err := child.Close(); err != nil {
			log.Printf("Failed to close child widget: %v", err)
		}
	}
	return nil
}

func (w *VFlexWidget) Layout(gtx layout.Context) layout.Dimensions {
	children := w.GetChildren()
	if len(children) == 0 {
		return layout.Dimensions{}
	}

	// Use Flex for equal distribution vertically
	var flexChildren []layout.FlexChild
	for _, child := range children {
		flexChildren = append(flexChildren, layout.Flexed(1, child.Layout))
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, flexChildren...)
}
