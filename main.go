package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
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
	th := cfg.Theme()

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

type Layoutable interface {
	Layout(gtx layout.Context) layout.Dimensions
}

func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	// Get the root widget from the dashboard service
	var rootWidget Layoutable = a.dashService.GetRootWidget()

	if rootWidget == nil {
		// Show loading or error state
		title := material.H3(a.theme, "Loading Dashboard...")
		title.Alignment = text.Middle
		rootWidget = title
	}

	return layout.Background{}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, a.theme.Bg)
			return layout.Dimensions{Size: gtx.Constraints.Min}
		}, func(gtx layout.Context) layout.Dimensions {
			return rootWidget.Layout(gtx)
		})
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
