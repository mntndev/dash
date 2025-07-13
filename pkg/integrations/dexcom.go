package integrations

import (
	"context"
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
	client         *dexcomshare.Client
	connected      bool
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	lastEntry      *dexcomshare.GlucoseEntry
	lastUpdate     time.Time
	updateInterval time.Duration
	historicalData []dexcomshare.GlucoseEntry
	maxHistory     int
}

func NewDexcomClient(config *config.DexcomConfig) *DexcomClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &DexcomClient{
		config:         config,
		ctx:            ctx,
		cancel:         cancel,
		updateInterval: 1 * time.Minute,
		historicalData: make([]dexcomshare.GlucoseEntry, 0),
		maxHistory:     36, // 3 hours of 5-minute readings
	}
}

func (dc *DexcomClient) Connect() error {
	client, err := dexcomshare.NewClient(dc.config.Username, dc.config.Password)
	if err != nil {
		return fmt.Errorf("failed to create Dexcom client: %w", err)
	}

	dc.mu.Lock()
	dc.client = client
	dc.connected = true
	dc.mu.Unlock()

	go dc.updateLoop()

	return nil
}

func (dc *DexcomClient) updateLoop() {
	_ = dc.fetchGlucoseData()

	ticker := time.NewTicker(dc.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dc.ctx.Done():
			return
		case <-ticker.C:
			_ = dc.fetchGlucoseData()
		}
	}
}

func (dc *DexcomClient) fetchGlucoseData() error {
	dc.mu.RLock()
	client := dc.client
	connected := dc.connected
	dc.mu.RUnlock()

	if !connected || client == nil {
		return fmt.Errorf("dexcom client not connected")
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

	if !dc.connected {
		return nil, time.Time{}, fmt.Errorf("dexcom client not connected")
	}

	if dc.lastEntry == nil {
		return nil, time.Time{}, fmt.Errorf("no glucose data available")
	}

	entry := *dc.lastEntry
	return &entry, dc.lastUpdate, nil
}

func (dc *DexcomClient) GetHistoricalGlucose() ([]dexcomshare.GlucoseEntry, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	if !dc.connected {
		return nil, fmt.Errorf("dexcom client not connected")
	}

	// Return a copy of the historical data
	historicalCopy := make([]dexcomshare.GlucoseEntry, len(dc.historicalData))
	copy(historicalCopy, dc.historicalData)
	return historicalCopy, nil
}

func (dc *DexcomClient) IsConnected() bool {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.connected
}

func (dc *DexcomClient) Close() error {
	dc.cancel()
	dc.mu.Lock()
	dc.connected = false
	dc.mu.Unlock()
	return nil
}
