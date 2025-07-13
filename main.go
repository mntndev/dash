package main

import (
	"embed"
	"log"
	"time"

	"github.com/mntndev/dash/pkg/config"
	"github.com/mntndev/dash/pkg/dashboard"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "dash",
		Description: "A configurable dashboard for external services",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Create and register the dashboard service
	dashService := dashboard.NewDashboardService(app)
	app.RegisterService(application.NewService(dashService))

	log.Println("Creating window...")
	window := app.Window.New()
	log.Println("Window created with defaults")

	window.SetTitle("Dash - Dashboard")

	// For window configuration, we'll need to load config directly here
	// since service methods now require context and are not available at this point
	configPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Printf("Failed to load config for window setup: %v. Using defaults.", err)
		cfg = getDefaultConfig()
	}

	// Check if fullscreen is enabled in config
	if cfg.Dashboard.Fullscreen {
		log.Println("Fullscreen mode enabled")
		window.Fullscreen()
	} else {
		window.SetSize(1200, 800)
		window.Center()
	}

	log.Println("Window configured, showing...")
	window.Show()
	log.Println("Window show() called")

	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			app.Event.Emit("time", now)
			time.Sleep(time.Second)
		}
	}()

	runErr := app.Run()

	if runErr != nil {
		log.Fatal(runErr)
	}
}

func getDefaultConfig() *config.Config {
	return &config.Config{
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
