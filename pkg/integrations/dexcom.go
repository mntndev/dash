package integrations

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mntndev/dash/pkg/config"
	dexcomshare "github.com/tgiv014/dexcom-share"
)

type DexcomProvider interface {
	GetDexcomClient() *DexcomClient
}

type DexcomClient struct {
	config         *config.DexcomConfig
	mu             sync.RWMutex
	lastEntry      *dexcomshare.GlucoseEntry
	lastUpdate     time.Time
	historicalData []dexcomshare.GlucoseEntry
	maxHistory     int
}

func NewDexcomClient(cfg *config.DexcomConfig) *DexcomClient {
	return &DexcomClient{
		config:         cfg,
		historicalData: make([]dexcomshare.GlucoseEntry, 0),
		maxHistory:     36, // 3 hours of 5-minute readings
	}
}

func (dc *DexcomClient) FetchGlucoseData() error {
	client, err := dexcomshare.NewClient(dc.config.Username, dc.config.Password)
	if err != nil {
		return fmt.Errorf("failed to create Dexcom client: %w", err)
	}

	log.Printf("Making Dexcom API request...")
	entries, err := client.ReadGlucose(180, dc.maxHistory) // 3 hours instead of 24
	if err != nil {
		return fmt.Errorf("failed to read glucose: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no glucose readings available")
	}

	dc.mu.Lock()
	dc.lastEntry = &entries[0]
	dc.lastUpdate = time.Now()
	dc.historicalData = entries
	dc.mu.Unlock()

	return nil
}

func (dc *DexcomClient) GetLatestGlucose() (*dexcomshare.GlucoseEntry, time.Time, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	if dc.lastEntry == nil {
		return nil, time.Time{}, fmt.Errorf("no glucose data available")
	}

	entry := *dc.lastEntry
	return &entry, dc.lastUpdate, nil
}

func (dc *DexcomClient) GetHistoricalGlucose() ([]dexcomshare.GlucoseEntry, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	// Return a copy of the historical data
	historicalCopy := make([]dexcomshare.GlucoseEntry, len(dc.historicalData))
	copy(historicalCopy, dc.historicalData)
	return historicalCopy, nil
}
