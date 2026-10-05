package gui

import (
	"context"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// singleInstanceID scopes the single-instance lock: every `differ` GUI process
// shares it, so later launches open tabs in the first one's window.
const singleInstanceID = "net.dannyeldridge.differ"

// Run opens the GUI window with repoPath as its first tab. If a differ window
// is already running, Run hands repoPath to it and returns. Otherwise it blocks
// until the window closes.
func Run(repoPath string) error {
	// Wails forwards os.Args[1:] to the running instance when this is a second
	// launch, so stamp them with the repo (see repoFlag for why not the cwd).
	os.Args = []string{os.Args[0], repoFlag, repoPath}

	app := newApp()
	return wails.Run(&options.App{
		Title:                    "differ",
		Width:                    1200,
		Height:                   800,
		EnableDefaultContextMenu: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind: []interface{}{app},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               singleInstanceID,
			OnSecondInstanceLaunch: app.handleSecondInstance,
		},
		OnStartup: func(ctx context.Context) {
			app.ctx = ctx
			setAppIcon()
			app.openRepo(repoPath)
		},
	})
}
