package widgets

import (
	"context"
	"fmt"
	"time"
)

type ClockWidget struct {
	*BaseWidget
	Format string
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
	}

	return widget, nil
}

func (w *ClockWidget) Update(ctx context.Context) error {
	now := time.Now()
	w.Data = &ClockData{
		Time:    now,
		Format:  w.Format,
		Display: now.Format(w.Format),
	}
	w.LastUpdate = now
	return nil
}