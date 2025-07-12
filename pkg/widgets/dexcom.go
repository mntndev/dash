package widgets

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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

	timestamp := parseDexcomTimestamp(latest.WT)
	if timestamp.IsZero() {
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

func parseDexcomTimestamp(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	
	// Dexcom timestamps come in format: "/Date(milliseconds)/"
	if !strings.HasPrefix(dateStr, "/Date(") || !strings.HasSuffix(dateStr, ")/") {
		return time.Time{}
	}
	
	// Extract the timestamp part
	timestampStr := strings.TrimPrefix(dateStr, "/Date(")
	timestampStr = strings.TrimSuffix(timestampStr, ")/")
	
	// Handle timezone offset if present (like "-0700")
	if idx := strings.LastIndex(timestampStr, "+"); idx > 0 {
		timestampStr = timestampStr[:idx]
	} else if idx := strings.LastIndex(timestampStr, "-"); idx > 0 {
		timestampStr = timestampStr[:idx]
	}
	
	// Parse milliseconds
	milliseconds, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return time.Time{}
	}
	
	// Convert to time.Time (Dexcom uses milliseconds since Unix epoch)
	return time.Unix(milliseconds/1000, (milliseconds%1000)*1000000).UTC()
}

