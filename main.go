package main

import (
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"

	"github.com/mntndev/dash/pkg/config"
	"github.com/mntndev/dash/pkg/dashboard"
)

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("Dash - Dashboard"))

		// Load configuration
		configPath := config.GetDefaultConfigPath()
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			log.Printf("Failed to load config: %v. Using defaults.", err)
			cfg = getDefaultConfig()
		}

		// Set additional window options based on config
		if cfg.Dashboard.Fullscreen {
			log.Println("Fullscreen mode enabled")
			// Set a large window size for fullscreen-like experience
			w.Option(app.Fullscreen.Option())
		} else {
			w.Option(app.Size(1200, 800))
		}

		if err := run(w, cfg); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

type App struct {
	dashService *dashboard.DashboardService
	theme       *material.Theme
}

func run(w *app.Window, cfg *config.Config) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	// Create dashboard service
	dashService := dashboard.NewDashboardService(w)

	// Initialize the service
	if err := dashService.Initialize(); err != nil {
		log.Printf("Failed to initialize dashboard service: %v", err)
	}

	dashApp := &App{
		dashService: dashService,
		theme:       th,
	}

	var ops op.Ops
	for {
		e := w.Event()
		if e, ok := e.(app.FrameEvent); ok {
			gtx := app.NewContext(&ops, e)
			dashApp.Layout(gtx)
			e.Frame(gtx.Ops)
		}
		if _, ok := e.(app.DestroyEvent); ok {
			return nil
		}
	}
}

func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	// Get the root widget from the dashboard service
	rootWidget := a.dashService.GetRootWidget()

	if rootWidget == nil {
		// Show loading or error state
		title := material.H3(a.theme, "Loading Dashboard...")
		title.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		title.Alignment = text.Middle
		return layout.Center.Layout(gtx, title.Layout)
	}

	// Render the root widget
	return rootWidget.Layout(gtx)
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
