//go:build windows

package main

import (
	"context"
	"embed"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"watermark-studio/internal/app"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Cria instância da aplicação
	a := app.New()

	// Configuração da janela
	err := wails.Run(&options.App{
		Title:             "Watermark Studio",
		Width:             1024,
		Height:            768,
		MinWidth:          900,
		MinHeight:         650,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 245, G: 245, B: 250, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		OnStartup:        a.Startup,
		OnDomReady:       a.DomReady,
		OnBeforeClose:    a.BeforeClose,
		OnShutdown:       a.Shutdown,
		Bind: []interface{}{
			a,
		},
	})

	if err != nil {
		runtime.LogErrorf("Erro ao iniciar aplicação: %v", err)
	}
}