package main

import (
	"embed"
	"log"

	"contaduria/mvp/backend"
	"contaduria/mvp/backend/bootstrap"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	runtime, err := bootstrap.NewPocketBaseRuntime("./pb_data")
	if err != nil {
		log.Fatalf("failed to bootstrap pocketbase runtime: %v", err)
	}

	app := backend.NewApp(runtime)

	err = wails.Run(&options.App{
		Title:             "BLUECONTA-LITE",
		Width:             1120,
		Height:            760,
		MinWidth:          900,
		MinHeight:         600,
		DisableResize:     false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.Startup,
		OnShutdown: app.Shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatalf("failed to run wails app: %v", err)
	}
}
