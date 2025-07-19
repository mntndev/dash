package dashboard

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gioui.org/app"
	"github.com/mntndev/dash/pkg/config"
	"github.com/mntndev/dash/pkg/integrations"
	"github.com/mntndev/dash/pkg/widgets"
)

type DashboardService struct {
	config        *config.Config
	widgetManager *widgets.WidgetManager
	haClient      *integrations.HomeAssistantClient
	dexcomClient  *integrations.DexcomClient
	eventEmitter  EventEmitter
	window        *app.Window
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	initialized   bool
	rootWidget    widgets.Widget
}

type EventEmitter interface {
	Emit(event string, data interface{})
}


type GioEventEmitter struct {
	// For now, just log events - we'll expand this later
}

func (g *GioEventEmitter) Emit(event string, data interface{}) {
	log.Printf("Event emitted: %s", event)
}

// WidgetData represents the dynamic data of a widget.
type WidgetData struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Data       interface{} `json:"data"`
	LastUpdate time.Time   `json:"last_update"`
}

// DashboardInfo contains the static dashboard information.
type DashboardInfo struct {
	Title      string                 `json:"title"`
	Theme      string                 `json:"theme"`
	RootWidget widgets.Widget         `json:"root_widget"`
	Status     map[string]interface{} `json:"status"`
}

func NewDashboardService(window *app.Window) *DashboardService {
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
					Type:   "clock",
					Config: nil,
				},
			},
			Integrations: config.IntegrationsConfig{},
		}
	}

	// Create a simple event emitter for Gio
	eventEmitter := &GioEventEmitter{}

	service := &DashboardService{
		config:       cfg,
		eventEmitter: eventEmitter,
		window:       window,
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

func (ds *DashboardService) createWidgetWithChildren(widgetConfig config.WidgetConfig, idPrefix string) (widgets.Widget, error) {
	widgetID := fmt.Sprintf("%s_%s", idPrefix, widgetConfig.Type)
	log.Printf("Creating widget: %s (type: %s)", widgetID, widgetConfig.Type)

	// First, create all child widgets depth-first
	var childWidgets []widgets.Widget
	for i, childConfig := range widgetConfig.Children {
		childID := fmt.Sprintf("%s_child_%d", idPrefix, i)
		childWidget, err := ds.createWidgetWithChildren(childConfig, childID)
		if err != nil {
			return nil, fmt.Errorf("failed to create child widget %d: %w", i, err)
		}
		childWidgets = append(childWidgets, childWidget)
	}

	// Create the parent widget with ID and children at creation time
	widget, err := ds.widgetManager.GetFactory().Create(widgetConfig.Type, widgetID, widgetConfig.Config, childWidgets, ds.window)
	if err != nil {
		return nil, fmt.Errorf("failed to create widget %s: %w", widgetID, err)
	}

	// Initialize the widget immediately since it now has everything it needs
	log.Printf("Initializing widget: %s (type: %s)", widgetID, widgetConfig.Type)
	if err := widget.Init(ds.ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize widget %s: %w", widgetID, err)
	}

	// Store the widget in the manager so it can be found by TriggerWidget
	ds.widgetManager.StoreWidget(widgetID, widget)

	if len(childWidgets) > 0 {
		log.Printf("Created and initialized widget %s with %d children", widgetID, len(childWidgets))
	} else {
		log.Printf("Successfully created and initialized widget: %s", widgetID)
	}

	return widget, nil
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
		status["dexcom"] = true // Always available for stateless API
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
		if err := ds.haClient.Close(); err != nil {
			log.Printf("Failed to close Home Assistant client: %v", err)
		}
	}

	// Dexcom client is stateless, no cleanup needed

	return nil
}


// GetRootWidget returns the root widget for Gio UI rendering
func (ds *DashboardService) GetRootWidget() widgets.Widget {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.rootWidget
}
