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
	Value         int             `json:"value"`
	Trend         string          `json:"trend"`
	Timestamp     time.Time       `json:"timestamp"`
	Unit          string          `json:"unit"`
	Historical    []DexcomReading `json:"historical,omitempty"`
	LowThreshold  int             `json:"low_threshold"`
	HighThreshold int             `json:"high_threshold"`
}

type DexcomReading struct {
	Value     int       `json:"value"`
	Trend     string    `json:"trend"`
	Timestamp time.Time `json:"timestamp"`
}

func CreateDexcomWidget(id string, config map[string]interface{}, children []Widget, provider Provider) (Widget, error) {
	widget := &DexcomWidget{
		BaseWidget: &BaseWidget{
			ID:       id,
			Type:     "dexcom",
			Config:   config,
			Children: children,
		},
		dexcomProvider: provider,
	}

	return widget, nil
}

func (w *DexcomWidget) getLowThreshold() int {
	if lowThreshold, ok := w.Config["low_threshold"].(int); ok {
		return lowThreshold
	}
	if lowThreshold, ok := w.Config["low_threshold"].(float64); ok {
		return int(lowThreshold)
	}
	return 70 // default
}

func (w *DexcomWidget) getHighThreshold() int {
	if highThreshold, ok := w.Config["high_threshold"].(int); ok {
		return highThreshold
	}
	if highThreshold, ok := w.Config["high_threshold"].(float64); ok {
		return int(highThreshold)
	}
	return 160 // default
}

func (w *DexcomWidget) Init(ctx context.Context) error {
	w.LastUpdate = time.Now()
	
	dexcomClient := w.dexcomProvider.GetDexcomClient()
	if dexcomClient == nil || !dexcomClient.IsConnected() {
		// Dexcom client not connected yet, start a goroutine to wait for connection
		go w.waitForConnectionAndUpdate(ctx)
		return nil
	}
	
	// Start data update asynchronously to avoid blocking widget initialization
	go func() {
		if err := w.updateData(); err != nil {
			fmt.Printf("Failed to update Dexcom data during init: %v\n", err)
		}
	}()
	
	return nil
}

func (w *DexcomWidget) waitForConnectionAndUpdate(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	// Timeout after 30 seconds
	timeout := time.NewTimer(30 * time.Second)
	defer timeout.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-timeout.C:
			fmt.Printf("Timeout waiting for Dexcom connection\n")
			return
		case <-ticker.C:
			dexcomClient := w.dexcomProvider.GetDexcomClient()
			if dexcomClient != nil && dexcomClient.IsConnected() {
				if err := w.updateData(); err != nil {
					fmt.Printf("Failed to update Dexcom data: %v\n", err)
				} else {
					fmt.Printf("Dexcom widget successfully connected and updated\n")
				}
				return
			}
		}
	}
}

func (w *DexcomWidget) updateData() error {
	dexcomClient := w.dexcomProvider.GetDexcomClient()
	if dexcomClient == nil || !dexcomClient.IsConnected() {
		return fmt.Errorf("Dexcom client not connected")
	}

	latest, lastUpdate, err := dexcomClient.GetLatestGlucose()
	if err != nil {
		return fmt.Errorf("failed to get glucose data: %w", err)
	}

	trendString := formatTrendString(latest.Trend)

	// Try WT first, then DT as fallback
	timestamp := parseDexcomTimestamp(latest.WT)
	if timestamp.IsZero() {
		timestamp = parseDexcomTimestamp(latest.DT)
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	// Get historical data
	historical, err := dexcomClient.GetHistoricalGlucose()
	var historicalReadings []DexcomReading
	if err != nil {
		historicalReadings = make([]DexcomReading, 0)
	} else {
		// Convert historical data
		historicalReadings = make([]DexcomReading, 0, len(historical))
		for _, entry := range historical {
			// Try WT first, then DT as fallback
			histTimestamp := parseDexcomTimestamp(entry.WT)
			if histTimestamp.IsZero() {
				histTimestamp = parseDexcomTimestamp(entry.DT)
			}
			if histTimestamp.IsZero() {
				continue
			}

			histTrend := formatTrendString(entry.Trend)
			historicalReadings = append(historicalReadings, DexcomReading{
				Value:     entry.Value,
				Trend:     histTrend,
				Timestamp: histTimestamp,
			})
		}
	}

	w.Data = &DexcomData{
		Value:         latest.Value,
		Trend:         trendString,
		Timestamp:     timestamp,
		Unit:          "mg/dL",
		Historical:    historicalReadings,
		LowThreshold:  w.getLowThreshold(),
		HighThreshold: w.getHighThreshold(),
	}
	w.LastUpdate = lastUpdate
	return nil
}

func parseDexcomTimestamp(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}

	var timestampStr string

	// Dexcom timestamps can come in different formats: "/Date(milliseconds)/" or "Date(milliseconds)"
	if strings.HasPrefix(dateStr, "/Date(") && strings.HasSuffix(dateStr, ")/") {
		// Format: /Date(milliseconds)/
		timestampStr = strings.TrimPrefix(dateStr, "/Date(")
		timestampStr = strings.TrimSuffix(timestampStr, ")/")
	} else if strings.HasPrefix(dateStr, "Date(") && strings.HasSuffix(dateStr, ")") {
		// Format: Date(milliseconds)
		timestampStr = strings.TrimPrefix(dateStr, "Date(")
		timestampStr = strings.TrimSuffix(timestampStr, ")")
	} else {
		return time.Time{}
	}

	// Handle timezone offset if present (like "-0600")
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

func (w *DexcomWidget) Close() error {
	return nil
}

func formatTrendString(trend string) string {
	switch trend {
	case "1":
		return "↗↗"
	case "2":
		return "↗"
	case "3":
		return "↗"
	case "4":
		return "→"
	case "5":
		return "↘"
	case "6":
		return "↘"
	case "7":
		return "↘↘"
	default:
		return "?"
	}
}
