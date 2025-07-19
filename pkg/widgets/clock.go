package widgets

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type ClockConfig struct {
	Format string `yaml:"format"`
}

type ClockWidget struct {
	*BaseWidget
	Format     string
	lastSecond int
	provider   Provider
	data       *ClockData
	theme      *material.Theme
}

type ClockData struct {
	Time    time.Time `json:"time"`
	Format  string    `json:"format"`
	Display string    `json:"display"`
}

func CreateClockWidget(id string, config ast.Node, children []Widget, provider Provider, window *app.Window, theme *material.Theme) (Widget, error) {
	// Parse config using NodeToValue
	var clockConfig ClockConfig
	if config != nil {
		if err := yaml.NodeToValue(config, &clockConfig); err != nil {
			return nil, fmt.Errorf("failed to parse clock config: %w", err)
		}
	}

	format := strings.Trim(clockConfig.Format, `"`)
	if format == "" {
		format = "15:04:05" // default format
	}

	widget := &ClockWidget{
		BaseWidget: &BaseWidget{
			ID:       id,
			Type:     "clock",
			Config:   config,
			Children: children,
			window:   window,
		},
		Format:     format,
		lastSecond: -1, // Initialize to -1 to force first update
		provider:   provider,
		theme:      theme,
	}

	return widget, nil
}

func (w *ClockWidget) Init(ctx context.Context) error {
	now := time.Now()
	w.lastSecond = now.Second()

	w.data = &ClockData{
		Time:    now,
		Format:  w.Format,
		Display: now.Format(w.Format),
	}
	w.LastUpdate = now

	// Start a goroutine to update clock every second
	go w.startClockUpdater(ctx)
	return nil
}

func (w *ClockWidget) startClockUpdater(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			w.lastSecond = now.Second()
			w.data = &ClockData{
				Time:    now,
				Format:  w.Format,
				Display: now.Format(w.Format),
			}
			w.LastUpdate = now

			// Trigger window redraw
			w.Invalidate()
		}
	}
}

func (w *ClockWidget) Close() error {
	return nil
}

func (w *ClockWidget) Layout(gtx layout.Context) layout.Dimensions {
	clock_text := "Clock"
	if w.data != nil {
		clock_text = w.data.Display
	}

	label := material.H3(w.theme, clock_text)
	label.Alignment = text.Middle
	return label.Layout(gtx)
}
