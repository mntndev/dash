package widgets

import (
	"context"
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/mntndev/dash/pkg/integrations"
)

type DexcomConfig struct {
	LowThreshold  int `yaml:"low_threshold"`
	HighThreshold int `yaml:"high_threshold"`
}

type DexcomWidget struct {
	*BaseWidget
	dexcomProvider integrations.DexcomProvider
	provider       Provider
	lowThreshold   int
	highThreshold  int
	data           *DexcomData
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

func CreateDexcomWidget(id string, config ast.Node, children []Widget, provider Provider, window *app.Window) (Widget, error) {
	// Parse config using NodeToValue
	var dexcomConfig DexcomConfig
	if config != nil {
		if err := yaml.NodeToValue(config, &dexcomConfig); err != nil {
			return nil, fmt.Errorf("failed to parse dexcom config: %w", err)
		}
	}

	// Set defaults
	lowThreshold := dexcomConfig.LowThreshold
	if lowThreshold == 0 {
		lowThreshold = 70
	}
	highThreshold := dexcomConfig.HighThreshold
	if highThreshold == 0 {
		highThreshold = 160
	}

	widget := &DexcomWidget{
		BaseWidget: &BaseWidget{
			ID:       id,
			Type:     "dexcom",
			Config:   config,
			Children: children,
			window:   window,
		},
		dexcomProvider: provider,
		provider:       provider,
		lowThreshold:   lowThreshold,
		highThreshold:  highThreshold,
	}

	return widget, nil
}

func (w *DexcomWidget) getLowThreshold() int {
	return w.lowThreshold
}

func (w *DexcomWidget) getHighThreshold() int {
	return w.highThreshold
}

func (w *DexcomWidget) Init(ctx context.Context) error {
	w.LastUpdate = time.Now()

	// Always start the connection check asynchronously to avoid any blocking
	// during widget initialization
	go w.waitForConnectionAndUpdate(ctx)

	return nil
}

func (w *DexcomWidget) waitForConnectionAndUpdate(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Try to fetch data immediately
	if err := w.updateData(); err != nil {
		fmt.Printf("Failed to update Dexcom data: %v\n", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Use a timeout for potentially blocking provider calls
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("Recovered from panic in Dexcom provider call: %v\n", r)
					}
				}()

				if err := w.updateData(); err != nil {
					fmt.Printf("Failed to update Dexcom data: %v\n", err)
				}
			}()
		}
	}
}

func (w *DexcomWidget) updateData() error {
	dexcomClient := w.dexcomProvider.GetDexcomClient()
	if dexcomClient == nil {
		return fmt.Errorf("dexcom client not available")
	}

	// Fetch fresh data from the API
	if err := dexcomClient.FetchGlucoseData(); err != nil {
		return fmt.Errorf("failed to fetch glucose data: %w", err)
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

	data := &DexcomData{
		Value:         latest.Value,
		Trend:         trendString,
		Timestamp:     timestamp,
		Unit:          "mg/dL",
		Historical:    historicalReadings,
		LowThreshold:  w.getLowThreshold(),
		HighThreshold: w.getHighThreshold(),
	}

	w.setDataAndInvalidate(data)
	w.LastUpdate = lastUpdate
	return nil
}

func (w *DexcomWidget) setDataAndInvalidate(data *DexcomData) {
	w.data = data
	w.LastUpdate = time.Now()
	w.Invalidate()
}

func parseDexcomTimestamp(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}

	var timestampStr string

	// Dexcom timestamps can come in different formats: "/Date(milliseconds)/" or "Date(milliseconds)"
	switch {
	case strings.HasPrefix(dateStr, "/Date(") && strings.HasSuffix(dateStr, ")/"):
		// Format: /Date(milliseconds)/
		timestampStr = strings.TrimPrefix(dateStr, "/Date(")
		timestampStr = strings.TrimSuffix(timestampStr, ")/")
	case strings.HasPrefix(dateStr, "Date(") && strings.HasSuffix(dateStr, ")"):
		timestampStr = strings.TrimPrefix(dateStr, "Date(")
		timestampStr = strings.TrimSuffix(timestampStr, ")")
	default:
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

func (w *DexcomWidget) Layout(gtx layout.Context) layout.Dimensions {
	text := "Dexcom"
	if w.data != nil {
		text = fmt.Sprintf("%d %s %s", w.data.Value, w.data.Unit, w.data.Trend)
	}

	th := material.NewTheme()
	label := material.H4(th, text)
	label.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	return label.Layout(gtx)
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
