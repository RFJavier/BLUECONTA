package main

import (
	"embed"
	"log"

	"contaduria/mvp/backend"
	"contaduria/mvp/backend/bootstrap"
	"contaduria/mvp/shared"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed frontend/dist
var assets embed.FS

//go:embed LICENSE docs/THIRD_PARTY_LICENSES.md docs/MANUAL.md
var legalDocs embed.FS

func main() {
	runtime, err := bootstrap.NewPocketBaseRuntime("./pb_data")
	if err != nil {
		log.Fatalf("failed to bootstrap pocketbase runtime: %v", err)
	}

	app := backend.NewApp(runtime)

	license, err := legalDocs.ReadFile("LICENSE")
	if err != nil {
		log.Fatalf("failed to read LICENSE: %v", err)
	}
	thirdParty, err := legalDocs.ReadFile("docs/THIRD_PARTY_LICENSES.md")
	if err != nil {
		log.Fatalf("failed to read THIRD_PARTY_LICENSES.md: %v", err)
	}
	// manualMd, err := legalDocs.ReadFile("docs/MANUAL.md")
	// if err != nil {
	// 	log.Fatalf("failed to read MANUAL.md: %v", err)
	// }

	app.SetLegalInfo(&shared.LegalInfo{
		License:            string(license),
		ThirdPartyLicenses: string(thirdParty),
	})
	if err := backend.EnsureDocsFiles("./pb_data", string(license), string(thirdParty)); err != nil {
		log.Fatalf("failed to ensure docs files: %v", err)
	}

	// if err := backend.EnsureDocsFiles("./pb_data", string(license), string(thirdParty), string(manualMd)); err != nil {
	// 	log.Fatalf("failed to ensure docs files: %v", err)
	// }

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
