//go:build desktop

package desktop

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/cqi/my_agentmux/internal/agent"
	"github.com/cqi/my_agentmux/internal/config"
	"github.com/cqi/my_agentmux/internal/monitor"
	"github.com/cqi/my_agentmux/internal/session"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// App struct
type App struct {
	ctx     context.Context
	cfg     *config.Config
	session *SessionService
	terminal *TerminalService
	monitor *MonitorService
}

// NewApp creates a new App struct
func NewApp(cfg *config.Config, session *SessionService, terminal *TerminalService, monitor *MonitorService) *App {
	return &App{
		cfg:      cfg,
		session:  session,
		terminal: terminal,
		monitor:  monitor,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.session.startup(ctx)
	a.terminal.startup(ctx)
	a.monitor.startup(ctx)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// Run starts the Wails application
func Run(assets embed.FS, cfg *config.Config) error {
	// Initialize core services
	mgr, err := session.NewManager(cfg)
	if err != nil {
		return fmt.Errorf("initializing session manager: %w", err)
	}

	runner := agent.NewRunner(cfg, mgr)

	logger, err := monitor.NewLogger(cfg.LogsDir(), 50)
	if err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}

	watcher := monitor.NewWatcher(mgr.TmuxClient(), logger, 500)
	health := monitor.NewHealthMonitor(watcher, 5*time.Second)

	// Initialize desktop services
	sessionSvc := NewSessionService(mgr, runner)
	terminalSvc := NewTerminalService(mgr)
	monitorSvc := NewMonitorService(health, watcher, logger)

	// Create an instance of the app structure
	app := NewApp(cfg, sessionSvc, terminalSvc, monitorSvc)

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "AgentMux",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
			sessionSvc,
			terminalSvc,
			monitorSvc,
		},
	})

	if err != nil {
		return err
	}

	return nil
}
