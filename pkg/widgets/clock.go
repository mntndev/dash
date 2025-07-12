package widgets

import (
	"context"
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

func CreateClockWidget(config map[string]interface{}, service ServiceProvider) (Widget, error) {
	format := getStringConfig(config, "format", "15:04:05")
	
	widget := &ClockWidget{
		BaseWidget: &BaseWidget{
			ID:       generateWidgetID(),
			Type:     "clock",
			Config:   config,
		},
		Format: format,
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