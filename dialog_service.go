package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// DialogService handles native file dialogs
type DialogService struct {
	app *application.App
}

// NewDialogService creates a new DialogService
func NewDialogService(app interface{}) *DialogService {
	return &DialogService{}
}

// SaveIconDialog opens a system save dialog and saves the icon to the chosen path.
// Returns the saved file path, or empty string if cancelled.
func (d *DialogService) SaveIconDialog(id string, format string, size int, color string) (string, error) {
	if d.app == nil {
		d.app = application.Get()
	}

	// Use the shared global iconService instance
	iconSvc := globalIconService
	if iconSvc == nil {
		return "", fmt.Errorf("icon service not initialized")
	}

	svgContent, err := iconSvc.GetIconSVG(id)
	if err != nil {
		return "", err
	}

	icon, _ := iconSvc.GetIcon(id)
	safeName := sanitizeFilename(icon.Name)

	svgBytes := []byte(svgContent)
	if color != "" {
		svgBytes = applySVGColor(svgBytes, color)
	}

	ext := ".svg"
	filterName := "SVG Files"
	filterPattern := "*.svg"
	if format == "png" {
		ext = fmt.Sprintf("_%d.png", size)
		filterName = "PNG Files"
		filterPattern = "*.png"
	}
	defaultFilename := safeName + ext

	// Show native save dialog
	result, err := d.app.Dialog.SaveFile().
		SetFilename(defaultFilename).
		AddFilter(filterName, filterPattern).
		PromptForSingleSelection()
	if err != nil {
		return "", fmt.Errorf("save dialog error: %w", err)
	}
	if result == "" {
		return "", nil // User cancelled
	}

	// Write the file
	if format == "png" {
		if size <= 0 {
			size = 24
		}
		tmpDir := filepath.Join(os.TempDir(), "iconstore-exports")
		os.MkdirAll(tmpDir, 0755)
		pngPath, err := convertSVGToPNG(svgBytes, safeName, size, tmpDir)
		if err != nil {
			return "", err
		}
		data, err := os.ReadFile(pngPath)
		if err != nil {
			return "", err
		}
		return result, os.WriteFile(result, data, 0644)
	}

	return result, os.WriteFile(result, svgBytes, 0644)
}
