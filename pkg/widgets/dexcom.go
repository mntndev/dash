package widgets

import (
	"context"
	"fmt"
	"time"

	"github.com/mntndev/dash/pkg/integrations"
)

type DexcomWidget struct {
	*BaseWidget
	DexcomClient *integrations.DexcomClient
}

type DexcomData struct {
	Value     int       `json:"value"`
	Trend     string    `json:"trend"`
	Timestamp time.Time `json:"timestamp"`
	Unit      string    `json:"unit"`
}

func CreateDexcomWidget(config map[string]interface{}, service ServiceProvider) (Widget, error) {
	widget := &DexcomWidget{
		BaseWidget: &BaseWidget{
			ID:       generateWidgetID(),
			Type:     "dexcom",
			Config:   config,
		},
		DexcomClient: service.GetDexcomClient(),
	}

	return widget, nil
}

func (w *DexcomWidget) Update(ctx context.Context) error {
	if w.DexcomClient == nil || !w.DexcomClient.IsConnected() {
		return fmt.Errorf("Dexcom client not connected")
	}

	// Get the latest glucose reading from the integration client
	latest, lastUpdate, err := w.DexcomClient.GetLatestGlucose()
	if err != nil {
		return fmt.Errorf("failed to get glucose data: %w", err)
	}

	// Convert trend to human-readable string
	var trendString string
	switch latest.Trend {
	case "1":
		trendString = "↗↗" // Double up
	case "2":
		trendString = "↗"  // Single up
	case "3":
		trendString = "↗"  // Forty-five up
	case "4":
		trendString = "→"  // Flat
	case "5":
		trendString = "↘"  // Forty-five down
	case "6":
		trendString = "↘"  // Single down
	case "7":
		trendString = "↘↘" // Double down
	default:
		trendString = "?"
	}

	// Parse the timestamp string - Dexcom uses .NET JSON date format
	var timestamp time.Time
	if latest.ST != "" {
		// Extract the timestamp from .NET JSON date format: /Date(1136239445000)/
		// For now, let's just use the current time and log the actual format for debugging
		timestamp = time.Now()
		// TODO: Implement proper .NET JSON date parsing
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

// SetDexcomClient method removed - client is now injected during widget creation