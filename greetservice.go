package main

import (
	"log"

	"github.com/mntndev/dash/pkg/config"
	"github.com/mntndev/dash/pkg/dashboard"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type DashboardAppService struct {
	dashboardService *dashboard.DashboardService
	eventEmitter     *WailsEventEmitter
}

type WailsEventEmitter struct {
	app *application.App
}

func (w *WailsEventEmitter) Emit(event string, data interface{}) {
	if w.app != nil {
		w.app.Event.Emit(event, data)
	}
}

func NewDashboardAppService(app *application.App) *DashboardAppService {
	eventEmitter := &WailsEventEmitter{app: app}
	
	configPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Printf("Failed to load config: %v. Using default config.", err)
		cfg = getDefaultConfig()
	}
	
	dashboardService := dashboard.NewDashboardService(cfg, eventEmitter)
	// Initialize dashboard service in a goroutine to avoid blocking app startup
	go func() {
		if err := dashboardService.Initialize(); err != nil {
			log.Printf("Failed to initialize dashboard service: %v", err)
		}
	}()
	
	return &DashboardAppService{
		dashboardService: dashboardService,
		eventEmitter:     eventEmitter,
	}
}

func (s *DashboardAppService) GetDashboardData() (string, error) {
	return s.dashboardService.GetDashboardJSON()
}

func (s *DashboardAppService) TriggerWidget(widgetID string) error {
	return s.dashboardService.TriggerWidget(widgetID)
}

func (s *DashboardAppService) SetLightBrightness(widgetID string, brightness int) error {
	return s.dashboardService.SetLightBrightness(widgetID, brightness)
}

func (s *DashboardAppService) ReloadConfig() error {
	return s.dashboardService.ReloadConfig(config.GetDefaultConfigPath())
}

func getDefaultConfig() *config.Config {
	return &config.Config{
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
