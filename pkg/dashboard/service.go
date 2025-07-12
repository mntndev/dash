package dashboard

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mntndev/dash/pkg/config"
	"github.com/mntndev/dash/pkg/integrations"
	"github.com/mntndev/dash/pkg/widgets"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type DashboardService struct {
	config        *config.Config
	widgetManager *widgets.WidgetManager
	haClient      *integrations.HomeAssistantClient
	dexcomClient  *integrations.DexcomClient
	eventEmitter  EventEmitter
	app           *application.App
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	initialized   bool
	rootWidget    widgets.Widget
}

type EventEmitter interface {
	Emit(event string, data interface{})
}

type WailsEventEmitter struct {
	app *application.App
}

func (w *WailsEventEmitter) Emit(event string, data interface{}) {
	if w.app != nil {
		w.app.Event.Emit(event, data)
	}
}

// WidgetData represents the dynamic data of a widget
type WidgetData struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Data       interface{} `json:"data"`
	LastUpdate time.Time   `json:"last_update"`
}

// DashboardInfo contains the static dashboard information
type DashboardInfo struct {
	Title      string                 `json:"title"`
	Theme      string                 `json:"theme"`
	RootWidget widgets.Widget         `json:"root_widget"`
	Status     map[string]interface{} `json:"status"`
}

func NewDashboardService(app *application.App) *DashboardService {
	ctx, cancel := context.WithCancel(context.Background())

	// Load config with fallback to default
	configPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Printf("Failed to load config: %v. Using default config.", err)
		cfg = &config.Config{
			Dashboard: config.DashboardConfig{
				Title: "My Dashboard",
				Theme: "dark",
				Widget: config.WidgetConfig{
					Type: "clock",
					Config: map[string]interface{}{
						"format": "15:04:05",
					},
				},
			},
			Integrations: config.IntegrationsConfig{},
		}
	}

	// Create event emitter
	eventEmitter := &WailsEventEmitter{app: app}

	service := &DashboardService{
		config:       cfg,
		eventEmitter: eventEmitter,
		app:          app,
		ctx:          ctx,
		cancel:       cancel,
	}

	return service
}

func (ds *DashboardService) Initialize() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	widgetFactory := widgets.NewDefaultWidgetFactory(ds)
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

	log.Printf("Creating and initializing widgets...")
	if err := ds.createWidgets(); err != nil {
		return fmt.Errorf("failed to create widgets: %w", err)
	}
	
	if ds.rootWidget == nil {
		return fmt.Errorf("root widget not found")
	}
	log.Printf("Widgets created and initialized successfully")

	ds.initialized = true

	dashboardInfo := ds.getDashboardInfo()
	ds.eventEmitter.Emit("dashboard_info", dashboardInfo)

	log.Printf("Dashboard service initialized successfully")
	return nil
}

func (ds *DashboardService) createWidgets() error {
	log.Printf("Creating root widget of type: %s", ds.config.Dashboard.Widget.Type)
	log.Printf("Root widget has %d children", len(ds.config.Dashboard.Widget.Children))

	// Create root widget with depth-first approach
	rootWidget, err := ds.createWidgetWithChildren(ds.config.Dashboard.Widget, "root")
	if err != nil {
		return fmt.Errorf("failed to create root widget: %w", err)
	}

	// Store the root widget
	ds.rootWidget = rootWidget
	log.Printf("Created widget hierarchy with root: %s", rootWidget.GetID())

	return nil
}

func (ds *DashboardService) createWidgetWithChildren(config config.WidgetConfig, idPrefix string) (widgets.Widget, error) {
	widgetID := fmt.Sprintf("%s_%s", idPrefix, config.Type)
	log.Printf("Creating widget: %s (type: %s)", widgetID, config.Type)

	// First, create all child widgets depth-first
	var childWidgets []widgets.Widget
	for i, childConfig := range config.Children {
		childID := fmt.Sprintf("%s_child_%d", idPrefix, i)
		childWidget, err := ds.createWidgetWithChildren(childConfig, childID)
		if err != nil {
			return nil, fmt.Errorf("failed to create child widget %d: %w", i, err)
		}
		childWidgets = append(childWidgets, childWidget)
	}

	// Create the parent widget with ID and children at creation time
	widget, err := ds.widgetManager.GetFactory().Create(config.Type, widgetID, config.Config, childWidgets)
	if err != nil {
		return nil, fmt.Errorf("failed to create widget %s: %w", widgetID, err)
	}

	// Initialize the widget immediately since it now has everything it needs
	log.Printf("Initializing widget: %s (type: %s)", widgetID, config.Type)
	if err := widget.Init(ds.ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize widget %s: %w", widgetID, err)
	}

	if len(childWidgets) > 0 {
		log.Printf("Created and initialized widget %s with %d children", widgetID, len(childWidgets))
	} else {
		log.Printf("Successfully created and initialized widget: %s", widgetID)
	}

	return widget, nil
}

func (ds *DashboardService) emitDashboardInfo() {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	dashboardInfo := ds.getDashboardInfo()
	ds.eventEmitter.Emit("dashboard_info", dashboardInfo)
}

func (ds *DashboardService) getDashboardInfo() DashboardInfo {
	status := make(map[string]interface{})

	if !ds.initialized {
		return DashboardInfo{
			Title:      ds.config.Dashboard.Title,
			Theme:      ds.config.Dashboard.Theme,
			RootWidget: nil, // No root widget during initialization
			Status:     map[string]interface{}{"initializing": true},
		}
	}

	if ds.haClient != nil {
		status["home_assistant"] = ds.haClient.IsConnected()
	}
	if ds.dexcomClient != nil {
		status["dexcom"] = ds.dexcomClient.IsConnected()
	}

	return DashboardInfo{
		Title:      ds.config.Dashboard.Title,
		Theme:      ds.config.Dashboard.Theme,
		RootWidget: ds.rootWidget,
		Status:     status,
	}
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

func (ds *DashboardService) Emit(event string, data interface{}) {
	ds.eventEmitter.Emit(event, data)
}



func (ds *DashboardService) GetConfig() *config.Config {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.config
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

// Wails Service Lifecycle Methods

// ServiceStartup is called when the service is being started
func (ds *DashboardService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	log.Println("Dashboard service starting up...")
	
	// Initialize the dashboard service in a goroutine to avoid blocking startup
	go func() {
		if err := ds.Initialize(); err != nil {
			log.Printf("Failed to initialize dashboard service: %v", err)
		} else {
			log.Println("Dashboard service initialized successfully")
		}
	}()
	
	return nil
}

// ServiceShutdown is called when the service is being shut down
func (ds *DashboardService) ServiceShutdown(ctx context.Context) error {
	log.Println("Dashboard service shutting down...")
	return ds.Close()
}

// Wails Exported Methods

// GetDashboardInfo returns the current dashboard structure and status for Wails
func (ds *DashboardService) GetDashboardInfo(ctx context.Context) (*DashboardInfo, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	info := ds.getDashboardInfo()
	return &info, nil
}

// GetWidgetData returns the current data for a specific widget for Wails
func (ds *DashboardService) GetWidgetData(ctx context.Context, widgetID string) (*WidgetData, error) {
	// This method doesn't exist yet, I need to add it
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	
	widget, exists := ds.widgetManager.GetWidget(widgetID)
	if !exists {
		return nil, fmt.Errorf("widget %s not found", widgetID)
	}
	
	return &WidgetData{
		ID:         widget.GetID(),
		Type:       widget.GetType(),
		Data:       widget.GetData(),
		LastUpdate: time.Now(), // You may want to get this from the widget
	}, nil
}

// TriggerWidget activates a triggerable widget for Wails
func (ds *DashboardService) TriggerWidget(ctx context.Context, widgetID string) error {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	widget, exists := ds.widgetManager.GetWidget(widgetID)
	if !exists {
		return fmt.Errorf("widget not found: %s", widgetID)
	}

	if triggerable, ok := widget.(widgets.Triggerable); ok {
		return triggerable.Trigger()
	}

	return fmt.Errorf("widget %s does not support triggering", widgetID)
}

// SetLightBrightness controls the brightness of a light widget for Wails
func (ds *DashboardService) SetLightBrightness(ctx context.Context, widgetID string, brightness int) error {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	widget, exists := ds.widgetManager.GetWidget(widgetID)
	if !exists {
		return fmt.Errorf("widget not found: %s", widgetID)
	}

	if brightnessControllable, ok := widget.(widgets.BrightnessControllable); ok {
		return brightnessControllable.SetBrightness(brightness)
	}

	return fmt.Errorf("widget %s does not support brightness control", widgetID)
}

// ReloadConfig reloads the dashboard configuration for Wails
func (ds *DashboardService) ReloadConfig(ctx context.Context) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	configPath := config.GetDefaultConfigPath()
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

	widgetFactory := widgets.NewDefaultWidgetFactory(ds)
	ds.widgetManager = widgets.NewWidgetManager(widgetFactory)

	return ds.Initialize()
}

// IsFullscreenEnabled returns whether fullscreen mode is enabled
func (ds *DashboardService) IsFullscreenEnabled(ctx context.Context) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.config.Dashboard.Fullscreen
}
