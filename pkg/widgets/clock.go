package widgets

import (
	"context"
	"fmt"
	"time"
)

type ClockWidget struct {
	*BaseWidget
	Format string
	lastSecond int
}

type ClockData struct {
	Time   time.Time `json:"time"`
	Format string    `json:"format"`
	Display string   `json:"display"`
}

func CreateClockWidget(config map[string]interface{}) (Widget, error) {
	var clockConfig ClockConfig
	parser := NewConfigParser()
	if err := parser.ParseConfig(config, &clockConfig); err != nil {
		return nil, fmt.Errorf("invalid clock configuration: %w", err)
	}
	
	widget := &ClockWidget{
		BaseWidget: &BaseWidget{
			ID:       generateWidgetID(),
			Type:     "clock",
			Config:   config,
		},
		Format: clockConfig.Format,
		lastSecond: -1, // Initialize to -1 to force first update
	}

	return widget, nil
}

func (w *ClockWidget) ShouldUpdate() bool {
	now := time.Now()
	currentSecond := now.Second()
	
	// Update if the second has changed
	if currentSecond != w.lastSecond {
		return true
	}
	
	return false
}

func (w *ClockWidget) Update(ctx context.Context) error {
	now := time.Now()
	w.lastSecond = now.Second()
	
	w.Data = &ClockData{
		Time:    now,
		Format:  w.Format,
		Display: now.Format(w.Format),
	}
	w.LastUpdate = now
	return nil
}