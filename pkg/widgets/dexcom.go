package widgets

import (
	"context"
	"fmt"
	"time"

	"github.com/mntndev/dash/pkg/integrations"
)

type DexcomWidget struct {
	*BaseWidget
	dexcomProvider integrations.DexcomProvider
}

type DexcomData struct {
	Value     int       `json:"value"`
	Trend     string    `json:"trend"`
	Timestamp time.Time `json:"timestamp"`
	Unit      string    `json:"unit"`
}

func CreateDexcomWidget(config map[string]interface{}, dexcomProvider integrations.DexcomProvider) (Widget, error) {
	widget := &DexcomWidget{
		BaseWidget: &BaseWidget{
			ID:       generateWidgetID(),
			Type:     "dexcom",
			Config:   config,
		},
		dexcomProvider: dexcomProvider,
	}

	return widget, nil
}

func (w *DexcomWidget) Update(ctx context.Context) error {
	dexcomClient := w.dexcomProvider.GetDexcomClient()
	if dexcomClient == nil || !dexcomClient.IsConnected() {
		return fmt.Errorf("Dexcom client not connected")
	}

	latest, lastUpdate, err := dexcomClient.GetLatestGlucose()
	if err != nil {
		return fmt.Errorf("failed to get glucose data: %w", err)
	}

	var trendString string
	switch latest.Trend {
	case "1":
		trendString = "↗↗"
	case "2":
		trendString = "↗"
	case "3":
		trendString = "↗"
	case "4":
		trendString = "→"
	case "5":
		trendString = "↘"
	case "6":
		trendString = "↘"
	case "7":
		trendString = "↘↘"
	default:
		trendString = "?"
	}

	var timestamp time.Time
	if latest.ST != "" {
		timestamp = time.Now()
	} else {
		timestamp = time.Now()
	}

	w.Data = &DexcomData{
		Value:     latest.Value,
		Trend:     trendString,
		Timestamp: timestamp,
		Unit:      "mg/dL",
	}
	w.LastUpdate = lastUpdate
	return nil
}

