package gui

import (
	"context"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// Run opens the GUI window rooted at repoPath. It blocks until the window closes.
func Run(repoPath string) error {
	app := &App{repoPath: repoPath}
	return wails.Run(&options.App{
		Title:                    "differ",
		Width:                    1200,
		Height:                   800,
		EnableDefaultContextMenu: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind: []interface{}{app},
		OnStartup: func(ctx context.Context) {
			app.ctx = ctx
			app.startWatcher()
		},
	})
}
