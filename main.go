package main

import (
	"embed"
	_ "embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Check if running in MCP mode via command line flag
	if len(os.Args) > 1 && os.Args[1] == "--mcp" {
		runMCPServer()
		return
	}

	runDesktopApp()
}

func runDesktopApp() {
	iconSvc := NewIconService(nil)

	singleInstance := &application.SingleInstanceOptions{
		UniqueID: "com.mindspace.iconstore",
	}

	app := application.New(application.Options{
		Name:        "IconStore",
		Description: "SVG Icon Search & Download Tool",
		Services: []application.Service{
			application.NewService(iconSvc),
			application.NewService(NewDialogService(nil)),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		SingleInstance: singleInstance,
	})

	// Start embedded MCP HTTP server in background
	go startEmbeddedMCP(iconSvc)

	// Create main window
	mainWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "IconStore - Icon Search & Download",
		Frameless: true,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		Windows: application.WindowsWindow{
			DisableFramelessWindowDecorations: false,
		},
		Width:               1440,
		Height:              1000,
		DisableResize:       true,
		MaximiseButtonState: application.ButtonHidden,
		BackgroundColour:    application.NewRGB(15, 15, 17),
		URL:                 "/",
	})

	// Set single instance callback after window is created
	singleInstance.OnSecondInstanceLaunch = func(data application.SecondInstanceData) {
		mainWindow.Show()
		mainWindow.Focus()
	}

	// Hide window on close instead of quitting
	mainWindow.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		mainWindow.Hide()
		event.Cancel()
	})

	// Setup system tray
	trayMenu := application.NewMenu()
	trayMenu.Add("显示主窗口").OnClick(func(ctx *application.Context) {
		mainWindow.Show()
		mainWindow.Focus()
	})
	trayMenu.AddSeparator()
	trayMenu.Add("退出").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	tray := app.SystemTray.New()
	tray.SetLabel("IconStore")
	tray.SetTooltip("IconStore - Icon Search & Download")
	tray.SetMenu(trayMenu)
	tray.OnClick(func() {
		mainWindow.Show()
		mainWindow.Focus()
	})

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func runMCPServer() {
	svc := &IconService{}

	if err := svc.openDB(); err != nil {
		fmt.Fprintf(os.Stderr, "IconStore MCP: failed to open database: %v\n", err)
		return
	}

	stats := svc.GetStats()
	fmt.Fprintf(os.Stderr, "IconStore MCP: loaded %v icons\n", stats["totalIcons"])

	// Support --mcp [port], default 9393
	port := 9393
	if len(os.Args) > 2 {
		if p, err := strconv.Atoi(os.Args[2]); err == nil && p > 0 {
			port = p
		}
	}

	server := NewMCPServer(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/sse", server.HandleSSE)
	mux.HandleFunc("/mcp", server.HandleHTTP)
	mux.HandleFunc("/", server.HandleHTTP)

	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	fmt.Fprintf(os.Stderr, "IconStore MCP: listening on http://%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "IconStore MCP: server error: %v\n", err)
	}
}

// startEmbeddedMCP starts the MCP HTTP server in the same process
func startEmbeddedMCP(iconSvc *IconService) {
	server := NewMCPServer(iconSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("/sse", server.HandleSSE)
	mux.HandleFunc("/mcp", server.HandleHTTP)
	mux.HandleFunc("/", server.HandleHTTP)

	addr := net.JoinHostPort("127.0.0.1", "9393")
	fmt.Fprintf(os.Stderr, "[MCP] listening on http://%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "[MCP] server error: %v\n", err)
	}
}
