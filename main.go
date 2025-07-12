package main

import (
	"embed"
	_ "embed"
	"log"
	"time"

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

	dashService := NewDashboardAppService(app)

	app.RegisterService(application.NewService(dashService))

	log.Println("Creating window...")
	window := app.Window.New()
	log.Println("Window created with defaults")

	window.SetTitle("Dash - Dashboard")

	// Check if fullscreen is enabled in config
	if dashService.IsFullscreenEnabled() {
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

	err := app.Run()

	if err != nil {
		log.Fatal(err)
	}
}
