package widgets

import (
	"context"
	"time"
)

type ClockWidget struct {
	*BaseWidget
	Format     string
	lastSecond int
}

type ClockData struct {
	Time    time.Time `json:"time"`
	Format  string    `json:"format"`
	Display string    `json:"display"`
}

func CreateClockWidget(id string, config map[string]interface{}, children []Widget) (Widget, error) {
	format, _ := config["format"].(string)
	if format == "" {
		format = "15:04:05" // default format
	}

	widget := &ClockWidget{
		BaseWidget: &BaseWidget{
			ID:       id,
			Type:     "clock",
			Config:   config,
			Children: children,
		},
		Format:     format,
		lastSecond: -1, // Initialize to -1 to force first update
	}

	return widget, nil
}

func (w *ClockWidget) Init(ctx context.Context) error {
	now := time.Now()
	w.lastSecond = now.Second()

	w.Data = &ClockData{
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
			w.Data = &ClockData{
				Time:    now,
				Format:  w.Format,
				Display: now.Format(w.Format),
			}
			w.LastUpdate = now
		}
	}
}

func (w *ClockWidget) Close() error {
	return nil
}
