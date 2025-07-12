package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mntndev/dash/pkg/config"
	"github.com/mntndev/dash/pkg/integrations"
	"github.com/mntndev/dash/pkg/widgets"
)

type DashboardService struct {
	config          *config.Config
	widgetManager   *widgets.WidgetManager
	haClient        *integrations.HomeAssistantClient
	dexcomClient    *integrations.DexcomClient
	eventEmitter    EventEmitter
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	updateInterval  time.Duration
	initialized     bool
}

type EventEmitter interface {
	Emit(event string, data interface{})
}

type WidgetData struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Data     interface{}  `json:"data"`
	LastUpdate time.Time  `json:"last_update"`
	Children []WidgetData `json:"children,omitempty"`
}


type DashboardData struct {
	Title   string                 `json:"title"`
	Theme   string                 `json:"theme"`
	Widget  WidgetData            `json:"widget"`
	Status  map[string]interface{} `json:"status"`
}

func NewDashboardService(config *config.Config, eventEmitter EventEmitter) *DashboardService {
	ctx, cancel := context.WithCancel(context.Background())
	
	service := &DashboardService{
		config:         config,
		eventEmitter:   eventEmitter,
		ctx:            ctx,
		cancel:         cancel,
		updateInterval: 5 * time.Second,
	}
	
	
	return service
}

func (ds *DashboardService) Initialize() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	
	widgetFactory := widgets.NewDefaultWidgetFactory(ds, ds)
	widgetManager := widgets.NewWidgetManager(widgetFactory)
	ds.widgetManager = widgetManager
	
	if ds.config.Integrations.HomeAssistant != nil {
		log.Printf("Initializing Home Assistant client...")
		ds.haClient = integrations.NewHomeAssistantClient(ds.config.Integrations.HomeAssistant)
		go func() {
			log.Printf("Connecting to Home Assistant...")
			if err := ds.haClient.Connect(); err != nil {
				log.Printf("Failed to connect to Home Assistant: %v", err)
			} else {
				log.Printf("Successfully connected to Home Assistant")
				if err := ds.haClient.SubscribeEvents("state_changed"); err != nil {
					log.Printf("Failed to subscribe to state_changed events: %v", err)
				}
				go ds.handleHAEvents()
			}
		}()
	}
	
	if ds.config.Integrations.Dexcom != nil {
		log.Printf("Initializing Dexcom client...")
		ds.dexcomClient = integrations.NewDexcomClient(ds.config.Integrations.Dexcom)
		go func() {
			log.Printf("Connecting to Dexcom...")
			if err := ds.dexcomClient.Connect(); err != nil {
				log.Printf("Failed to connect to Dexcom: %v", err)
			} else {
				log.Printf("Successfully connected to Dexcom")
			}
		}()
	}
	
	log.Printf("Creating widgets...")
	if err := ds.createWidgets(); err != nil {
		return fmt.Errorf("failed to create widgets: %w", err)
	}
	log.Printf("Widgets created successfully")
	
	log.Printf("Starting update loop...")
	go ds.updateLoop()
	
	ds.initialized = true
	
	dashboardData := ds.getDashboardData()
	ds.eventEmitter.Emit("dashboard_update", dashboardData)
	
	log.Printf("Dashboard service initialized successfully")
	return nil
}

func (ds *DashboardService) createWidgets() error {
	log.Printf("Creating root widget of type: %s", ds.config.Dashboard.Widget.Type)
	return ds.createWidgetFromConfig(ds.config.Dashboard.Widget, "root")
}

func (ds *DashboardService) createWidgetFromConfig(config config.WidgetConfig, idPrefix string) error {
	widgetID := fmt.Sprintf("%s_%s", idPrefix, config.Type)
	log.Printf("Creating widget: %s (type: %s)", widgetID, config.Type)
	
	if err := ds.widgetManager.CreateWidget(widgetID, config.Type, config.Config); err != nil {
		return fmt.Errorf("failed to create widget %s: %w", widgetID, err)
	}
	log.Printf("Successfully created widget: %s", widgetID)
	
	
	for i, childConfig := range config.Children {
		childID := fmt.Sprintf("%s_child_%d", idPrefix, i)
		if err := ds.createWidgetFromConfig(childConfig, childID); err != nil {
			return err
		}
	}
	
	return nil
}

func (ds *DashboardService) updateLoop() {
	ticker := time.NewTicker(ds.updateInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ds.ctx.Done():
			return
		case <-ticker.C:
			if err := ds.updateWidgets(); err != nil {
				log.Printf("Error updating widgets: %v", err)
			}
		}
	}
}

func (ds *DashboardService) updateWidgets() error {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	
	if err := ds.widgetManager.UpdateAll(ds.ctx); err != nil {
		return err
	}
	
	dashboardData := ds.getDashboardData()
	ds.eventEmitter.Emit("dashboard_update", dashboardData)
	
	return nil
}

func (ds *DashboardService) handleHAEvents() {
	eventChan := ds.haClient.GetEventChannel()
	
	for {
		select {
		case <-ds.ctx.Done():
			return
		case event := <-eventChan:
			if event.EventType == "state_changed" {
				if err := ds.updateWidgets(); err != nil {
					log.Printf("Error updating widgets after HA event: %v", err)
				}
			}
		}
	}
}

func (ds *DashboardService) getDashboardData() DashboardData {
	status := make(map[string]interface{})
	
	if !ds.initialized {
		return DashboardData{
			Title:  ds.config.Dashboard.Title,
			Theme:  ds.config.Dashboard.Theme,
			Widget: WidgetData{
				ID:   "loading",
				Type: "loading",
				Data: map[string]string{"message": "Initializing dashboard..."},
			},
			Status: map[string]interface{}{"initializing": true},
		}
	}
	
	if ds.haClient != nil {
		status["home_assistant"] = ds.haClient.IsConnected()
	}
	if ds.dexcomClient != nil {
		status["dexcom"] = ds.dexcomClient.IsConnected()
	}
	
	rootWidget := ds.buildWidgetFromID("root", ds.config.Dashboard.Widget)
	
	return DashboardData{
		Title:  ds.config.Dashboard.Title,
		Theme:  ds.config.Dashboard.Theme,
		Widget: rootWidget,
		Status: status,
	}
}

func (ds *DashboardService) buildWidgetFromID(idPrefix string, config config.WidgetConfig) WidgetData {
	widgetID := fmt.Sprintf("%s_%s", idPrefix, config.Type)
	
	if widget, exists := ds.widgetManager.GetWidget(widgetID); exists {
		widgetData := WidgetData{
			ID:         widget.GetID(),
			Type:       widget.GetType(),
			Data:       widget.GetData(),
			LastUpdate: ds.getWidgetLastUpdate(widget),
		}
		
		if widget.IsContainer() && len(config.Children) > 0 {
			children := make([]WidgetData, len(config.Children))
			for i, childConfig := range config.Children {
				childID := fmt.Sprintf("%s_child_%d", idPrefix, i)
				children[i] = ds.buildWidgetFromID(childID, childConfig)
			}
			widgetData.Children = children
		}
		
		return widgetData
	}
	
	return WidgetData{
		ID:   widgetID,
		Type: config.Type,
		Data: nil,
	}
}

func (ds *DashboardService) getWidgetLastUpdate(widget widgets.Widget) time.Time {
	switch w := widget.(type) {
	case *widgets.HAEntityWidget:
		if w.BaseWidget != nil {
			return w.BaseWidget.LastUpdate
		}
	case *widgets.HAButtonWidget:
		if w.BaseWidget != nil {
			return w.BaseWidget.LastUpdate
		}
	case *widgets.HASwitchWidget:
		if w.BaseWidget != nil {
			return w.BaseWidget.LastUpdate
		}
	case *widgets.HALightWidget:
		if w.BaseWidget != nil {
			return w.BaseWidget.LastUpdate
		}
	case *widgets.DexcomWidget:
		if w.BaseWidget != nil {
			return w.BaseWidget.LastUpdate
		}
	case *widgets.ClockWidget:
		if w.BaseWidget != nil {
			return w.BaseWidget.LastUpdate
		}
	case *widgets.GrowWidget:
		if w.BaseWidget != nil {
			return w.BaseWidget.LastUpdate
		}
	}
	return time.Now()
}

func (ds *DashboardService) GetDashboardData() DashboardData {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.getDashboardData()
}

func (ds *DashboardService) GetHAClient() *integrations.HomeAssistantClient {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.haClient
}

func (ds *DashboardService) GetDexcomClient() *integrations.DexcomClient {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.dexcomClient
}

func (ds *DashboardService) TriggerWidget(widgetID string) error {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	
	widget, exists := ds.widgetManager.GetWidget(widgetID)
	if !exists {
		return fmt.Errorf("widget not found: %s", widgetID)
	}
	
	if buttonWidget, ok := widget.(*widgets.HAButtonWidget); ok {
		return buttonWidget.Trigger()
	}
	
	if switchWidget, ok := widget.(*widgets.HASwitchWidget); ok {
		return switchWidget.Trigger()
	}
	
	if lightWidget, ok := widget.(*widgets.HALightWidget); ok {
		return lightWidget.Trigger()
	}
	
	return fmt.Errorf("widget %s does not support triggering", widgetID)
}

func (ds *DashboardService) SetLightBrightness(widgetID string, brightness int) error {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	
	widget, exists := ds.widgetManager.GetWidget(widgetID)
	if !exists {
		return fmt.Errorf("widget not found: %s", widgetID)
	}
	
	if lightWidget, ok := widget.(*widgets.HALightWidget); ok {
		return lightWidget.SetBrightness(brightness)
	}
	
	return fmt.Errorf("widget %s is not a light widget", widgetID)
}

func (ds *DashboardService) GetConfig() *config.Config {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.config
}

func (ds *DashboardService) ReloadConfig(configPath string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	
	newConfig, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load new config: %w", err)
	}
	
	ds.config = newConfig
	
	if ds.haClient != nil {
		ds.haClient.Close()
		ds.haClient = nil
	}
	
	if ds.dexcomClient != nil {
		ds.dexcomClient.Close()
		ds.dexcomClient = nil
	}
	
	widgetFactory := widgets.NewDefaultWidgetFactory(ds, ds)
	ds.widgetManager = widgets.NewWidgetManager(widgetFactory)
	
	return ds.Initialize()
}

func (ds *DashboardService) Close() error {
	ds.cancel()
	
	if ds.haClient != nil {
		ds.haClient.Close()
	}
	
	if ds.dexcomClient != nil {
		ds.dexcomClient.Close()
	}
	
	return nil
}

func (ds *DashboardService) GetDashboardJSON() (string, error) {
	data := ds.GetDashboardData()
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal dashboard data: %w", err)
	}
	return string(jsonData), nil
}