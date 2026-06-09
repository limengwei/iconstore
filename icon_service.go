package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	_ "modernc.org/sqlite"
)

// Icon represents a single SVG icon in the library
type Icon struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Package  string `json:"package"`
	Tags     string `json:"tags"` // comma-separated, stored in DB
	Path     string `json:"path"` // relative path from icons root
}

// SearchResult represents paginated search results
type SearchResult struct {
	Icons    []Icon `json:"icons"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

// ExportOptions for downloading icons
type ExportOptions struct {
	Format    string `json:"format"`    // "svg" or "png"
	Size      int    `json:"size"`      // pixel size for PNG
	Color     string `json:"color"`     // hex color, e.g. "#FF5722"
	OutputDir string `json:"outputDir"` // output directory, empty means temp dir
}

// CategoryInfo represents a category with icon count
type CategoryInfo struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// globalIconService holds the singleton instance for cross-service access
var globalIconService *IconService

// IconService manages the icon library with SQLite backend
type IconService struct {
	mu     sync.RWMutex
	db     *sql.DB
	dbPath string
}

// NewIconService creates a new IconService with SQLite
func NewIconService(app interface{}) *IconService {
	svc := &IconService{}

	if err := svc.openDB(); err != nil {
		fmt.Fprintf(os.Stderr, "IconStore: failed to open database: %v\n", err)
		return svc
	}

	globalIconService = svc
	return svc
}

// findDBPath locates the database file
func findDBPath() string {
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	candidates := []string{
		filepath.Join(exeDir, "iconstore.db"),
		filepath.Join(exeDir, "..", "iconstore.db"),
		"./iconstore.db",
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			abs, _ := filepath.Abs(c)
			// Check if writable; if not, copy to APPDATA
			if f, err := os.OpenFile(abs, os.O_RDWR, 0); err == nil {
				f.Close()
				return abs
			}
			// Not writable, copy to user data dir
			appData := os.Getenv("APPDATA")
			if appData == "" {
				return abs // fallback, will likely fail but no better option
			}
			userDB := filepath.Join(appData, "IconStore", "iconstore.db")
			os.MkdirAll(filepath.Dir(userDB), 0755)
			// Copy if not already there or source is newer
			copyDBIfNeeded(abs, userDB)
			return userDB
		}
	}
	// Default: next to exe
	abs, _ := filepath.Abs(filepath.Join(exeDir, "iconstore.db"))
	return abs
}

// copyDBIfNeeded copies src to dst if dst doesn't exist or src is newer
func copyDBIfNeeded(src, dst string) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return
	}
	dstInfo, err := os.Stat(dst)
	if err == nil && dstInfo.ModTime().Equal(srcInfo.ModTime()) && dstInfo.Size() == srcInfo.Size() {
		return // already up to date
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return
	}
	os.WriteFile(dst, data, 0644)
	dst, _ = filepath.Abs(dst)
	os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
}

// openDB opens/creates the SQLite database
func (s *IconService) openDB() error {
	s.dbPath = findDBPath()
	fmt.Printf("[openDB] dbPath: %s\n", s.dbPath)

	var err error
	s.db, err = sql.Open("sqlite", s.dbPath+"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	// Create tables
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS icons (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			category    TEXT NOT NULL DEFAULT '',
			category_zh TEXT NOT NULL DEFAULT '',
			package     TEXT NOT NULL DEFAULT '',
			parent      TEXT NOT NULL DEFAULT '',
			tags        TEXT NOT NULL DEFAULT '',
			path        TEXT NOT NULL,
			svg_content TEXT NOT NULL DEFAULT '',
			file_mtime  INTEGER NOT NULL DEFAULT 0,
			file_size   INTEGER NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_icons_package ON icons(package);
		CREATE INDEX IF NOT EXISTS idx_icons_category ON icons(category);
		CREATE INDEX IF NOT EXISTS idx_icons_name ON icons(name);
		CREATE INDEX IF NOT EXISTS idx_icons_parent ON icons(parent);

		-- FTS5 virtual table for full-text search
		CREATE VIRTUAL TABLE IF NOT EXISTS icons_fts USING fts5(
			id UNINDEXED,
			name,
			category,
			category_zh,
			package,
			tags,
			content='icons',
			content_rowid='rowid'
		);

		-- Triggers to keep FTS in sync
		CREATE TRIGGER IF NOT EXISTS icons_ai AFTER INSERT ON icons BEGIN
			INSERT INTO icons_fts(rowid, id, name, category, category_zh, package, tags)
			VALUES (new.rowid, new.id, new.name, new.category, new.category_zh, new.package, new.tags);
		END;
		CREATE TRIGGER IF NOT EXISTS icons_ad AFTER DELETE ON icons BEGIN
			INSERT INTO icons_fts(icons_fts, rowid, id, name, category, category_zh, package, tags)
			VALUES ('delete', old.rowid, old.id, old.name, old.category, old.category_zh, old.package, old.tags);
		END;
		CREATE TRIGGER IF NOT EXISTS icons_au AFTER UPDATE ON icons BEGIN
			INSERT INTO icons_fts(icons_fts, rowid, id, name, category, category_zh, package, tags)
			VALUES ('delete', old.rowid, old.id, old.name, old.category, old.category_zh, old.package, old.tags);
			INSERT INTO icons_fts(rowid, id, name, category, category_zh, package, tags)
			VALUES (new.rowid, new.id, new.name, new.category, new.category_zh, new.package, new.tags);
		END;

		-- Collections metadata table
		CREATE TABLE IF NOT EXISTS collections (
			prefix        TEXT PRIMARY KEY,
			name          TEXT NOT NULL DEFAULT '',
			total         INTEGER NOT NULL DEFAULT 0,
			author_name   TEXT NOT NULL DEFAULT '',
			author_url    TEXT NOT NULL DEFAULT '',
			license_title TEXT NOT NULL DEFAULT '',
			license_spdx  TEXT NOT NULL DEFAULT '',
			license_url   TEXT NOT NULL DEFAULT '',
			height        INTEGER NOT NULL DEFAULT 24,
			category      TEXT NOT NULL DEFAULT '',
			palette       INTEGER NOT NULL DEFAULT 0,
			last_modified INTEGER NOT NULL DEFAULT 0,
			categories    TEXT NOT NULL DEFAULT '{}',
			suffixes      TEXT NOT NULL DEFAULT '{}',
			prefixes      TEXT NOT NULL DEFAULT '{}'
		);
	`)
	if err != nil {
		return fmt.Errorf("create tables: %w", err)
	}

	// Populate Chinese translations in tags if needed
	go s.populateChineseTranslations()

	return nil
}

// populateChineseTranslations adds Chinese translations to the tags field
// for all icons that don't have Chinese keywords yet.
func (s *IconService) populateChineseTranslations() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already populated
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM icons WHERE tags NOT LIKE '%,%' OR tags = ''`).Scan(&count)
	if count == 0 {
		// Check if any icon has Chinese chars in tags
		var hasChinese int
		s.db.QueryRow(`SELECT COUNT(*) FROM icons WHERE tags GLOB '*[一-龥]*' LIMIT 1`).Scan(&hasChinese)
		if hasChinese > 0 {
			fmt.Println("[populateChineseTranslations] Chinese translations already present, skipping")
			return
		}
	}

	fmt.Println("[populateChineseTranslations] Populating Chinese translations...")

	// Process in batches
	batchSize := 500
	offset := 0
	total := 0

	for {
		rows, err := s.db.Query(`
			SELECT rowid, name, tags FROM icons ORDER BY rowid LIMIT ? OFFSET ?
		`, batchSize, offset)
		if err != nil {
			fmt.Printf("[populateChineseTranslations] Query error: %v\n", err)
			break
		}

		type iconRow struct {
			rowid int
			name  string
			tags  string
		}
		var batch []iconRow
		for rows.Next() {
			var r iconRow
			rows.Scan(&r.rowid, &r.name, &r.tags)
			batch = append(batch, r)
		}
		rows.Close()

		if len(batch) == 0 {
			break
		}

		// Update each icon
		tx, _ := s.db.Begin()
		for _, r := range batch {
			zhKeywords := translateIconName(r.name)
			if zhKeywords == "" {
				continue
			}

			// Append Chinese keywords to existing tags
			newTags := r.tags
			if newTags != "" {
				newTags += ", " + zhKeywords
			} else {
				newTags = zhKeywords
			}

			tx.Exec(`UPDATE icons SET tags = ? WHERE rowid = ?`, newTags, r.rowid)
		}
		tx.Commit()
		total += len(batch)
		offset += batchSize
	}

	fmt.Printf("[populateChineseTranslations] Done. Processed %d icons.\n", total)
}

// Search searches icons by keyword using FTS5
func (s *IconService) Search(query string, pkg string, category string, page int, pageSize int) SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fmt.Printf("[Search] query=%q, pkg=%q, category=%q, page=%d, pageSize=%d\n", query, pkg, category, page, pageSize)

	if pageSize <= 0 {
		pageSize = 60
	}
	if page <= 0 {
		page = 1
	}

	query = strings.TrimSpace(query)
	pkg = strings.TrimSpace(pkg)
	category = strings.TrimSpace(category)

	// Build the filter path from pkg and category
	filterPath := ""
	if pkg != "" && category != "" {
		filterPath = pkg + "/" + category
	} else if pkg != "" {
		filterPath = pkg
	}

	var icons []Icon
	var total int
	offset := (page - 1) * pageSize

	if query != "" {
		// Use FTS5 for full-text search
		ftsQuery := buildFTSQuery(query)

		if filterPath != "" {
			// FTS + filter
			countRow := s.db.QueryRow(`
				SELECT COUNT(*) FROM icons_fts f
				JOIN icons i ON i.id = f.id
				WHERE icons_fts MATCH ? AND (i.package || '/' || i.category) LIKE ?
			`, ftsQuery, filterPath+"%")
			countRow.Scan(&total)

			rows, err := s.db.Query(`
				SELECT i.id, i.name, i.category, i.package, i.tags, i.path
				FROM icons_fts f
				JOIN icons i ON i.id = f.id
				WHERE icons_fts MATCH ? AND (i.package || '/' || i.category) LIKE ?
				ORDER BY rank
				LIMIT ? OFFSET ?
			`, ftsQuery, filterPath+"%", pageSize, offset)
			if err == nil {
				icons = scanIcons(rows)
			}
		} else {
			countRow := s.db.QueryRow(`
				SELECT COUNT(*) FROM icons_fts WHERE icons_fts MATCH ?
			`, ftsQuery)
			countRow.Scan(&total)

			rows, err := s.db.Query(`
				SELECT i.id, i.name, i.category, i.package, i.tags, i.path
				FROM icons_fts f
				JOIN icons i ON i.id = f.id
				WHERE icons_fts MATCH ?
				ORDER BY rank
				LIMIT ? OFFSET ?
			`, ftsQuery, pageSize, offset)
			if err == nil {
				icons = scanIcons(rows)
			}
		}
	} else {
		// No query: simple browse
		if filterPath != "" {
			countRow := s.db.QueryRow(`
				SELECT COUNT(*) FROM icons WHERE (package || '/' || category) LIKE ?
			`, filterPath+"%")
			countRow.Scan(&total)

			rows, err := s.db.Query(`
				SELECT id, name, category, package, tags, path
				FROM icons WHERE (package || '/' || category) LIKE ?
				ORDER BY name LIMIT ? OFFSET ?
			`, filterPath+"%", pageSize, offset)
			if err == nil {
				icons = scanIcons(rows)
			}
		} else {
			countRow := s.db.QueryRow(`SELECT COUNT(*) FROM icons`)
			countRow.Scan(&total)

			rows, err := s.db.Query(`
				SELECT id, name, category, package, tags, path
				FROM icons ORDER BY name LIMIT ? OFFSET ?
			`, pageSize, offset)
			if err == nil {
				icons = scanIcons(rows)
			}
		}
	}

	if icons == nil {
		icons = []Icon{}
	}

	fmt.Printf("[Search] Result: total=%d, returned=%d icons\n", total, len(icons))

	return SearchResult{
		Icons:    icons,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// buildFTSQuery converts user query to FTS5-compatible format
func buildFTSQuery(query string) string {
	// Split into words, join with AND for stricter matching, OR for broader
	words := strings.Fields(query)
	for i, w := range words {
		// Escape FTS5 special characters
		w = strings.NewReplacer(`"`, `""`).Replace(w)
		words[i] = "\"" + w + "\"*"
	}
	return strings.Join(words, " ")
}

func scanIcons(rows *sql.Rows) []Icon {
	defer rows.Close()
	var icons []Icon
	for rows.Next() {
		var ic Icon
		rows.Scan(&ic.ID, &ic.Name, &ic.Category, &ic.Package, &ic.Tags, &ic.Path)
		icons = append(icons, ic)
	}
	if icons == nil {
		icons = []Icon{}
	}
	return icons
}

// GetIcon retrieves a single icon by ID
func (s *IconService) GetIcon(id string) (Icon, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var ic Icon
	err := s.db.QueryRow(`
		SELECT id, name, category, package, tags, path FROM icons WHERE id = ?
	`, id).Scan(&ic.ID, &ic.Name, &ic.Category, &ic.Package, &ic.Tags, &ic.Path)
	if err != nil {
		return Icon{}, fmt.Errorf("icon not found: %s", id)
	}
	return ic, nil
}

// GetIconSVG returns the SVG content of an icon
func (s *IconService) GetIconSVG(id string) (string, error) {
	s.mu.RLock()
	var svgContent string
	err := s.db.QueryRow(`SELECT svg_content FROM icons WHERE id = ?`, id).Scan(&svgContent)
	s.mu.RUnlock()

	if err != nil {
		return "", fmt.Errorf("icon not found: %s", id)
	}

	if svgContent != "" {
		return svgContent, nil
	}

	return "", fmt.Errorf("icon %s has no SVG content in database", id)
}

// ExportIcon exports an icon with the given options, returns the file path
func (s *IconService) ExportIcon(id string, options ExportOptions) (string, error) {
	svgContent, err := s.GetIconSVG(id)
	if err != nil {
		return "", err
	}

	svgBytes := []byte(svgContent)

	// Apply color if specified
	if options.Color != "" {
		svgBytes = applySVGColor(svgBytes, options.Color)
	}

	// Determine output directory
	var outDir string
	if options.OutputDir != "" {
		outDir = options.OutputDir
	} else {
		outDir = filepath.Join(os.TempDir(), "iconstore-exports")
	}
	os.MkdirAll(outDir, 0755)

	// Get icon name for filename
	s.mu.RLock()
	var name string
	s.db.QueryRow(`SELECT name FROM icons WHERE id = ?`, id).Scan(&name)
	s.mu.RUnlock()

	safeName := sanitizeFilename(name)

	switch options.Format {
	case "svg", "":
		outPath := filepath.Join(outDir, safeName+".svg")
		err = os.WriteFile(outPath, svgBytes, 0644)
		if err != nil {
			return "", err
		}
		return outPath, nil

	case "png":
		if options.Size <= 0 {
			options.Size = 24
		}
		return convertSVGToPNG(svgBytes, safeName, options.Size, outDir)

	default:
		return "", fmt.Errorf("unsupported format: %s", options.Format)
	}
}

// GetCategories returns all categories with icon counts
func (s *IconService) GetCategories() []CategoryInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT package || '/' || category, COUNT(*)
		FROM icons
		GROUP BY package, category
		ORDER BY COUNT(*) DESC
	`)
	if err != nil {
		return []CategoryInfo{}
	}
	defer rows.Close()

	var cats []CategoryInfo
	for rows.Next() {
		var c CategoryInfo
		rows.Scan(&c.Name, &c.Count)
		cats = append(cats, c)
	}
	if cats == nil {
		cats = []CategoryInfo{}
	}
	return cats
}

// GetPackages returns all packages with icon counts
func (s *IconService) GetPackages() []CategoryInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT package, COUNT(*)
		FROM icons
		GROUP BY package
		ORDER BY COUNT(*) DESC
	`)
	if err != nil {
		return []CategoryInfo{}
	}
	defer rows.Close()

	var pkgs []CategoryInfo
	for rows.Next() {
		var c CategoryInfo
		rows.Scan(&c.Name, &c.Count)
		pkgs = append(pkgs, c)
	}
	if pkgs == nil {
		pkgs = []CategoryInfo{}
	}
	return pkgs
}

// GetCategoriesByPackage returns categories (subfolders) for a given package
func (s *IconService) GetCategoriesByPackage(pkg string) []CategoryInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT category, COUNT(*)
		FROM icons
		WHERE package = ?
		GROUP BY category
		ORDER BY COUNT(*) DESC
	`, pkg)
	if err != nil {
		return []CategoryInfo{}
	}
	defer rows.Close()

	var cats []CategoryInfo
	for rows.Next() {
		var c CategoryInfo
		rows.Scan(&c.Name, &c.Count)
		cats = append(cats, c)
	}
	if cats == nil {
		cats = []CategoryInfo{}
	}
	return cats
}

// GetStats returns icon library statistics
func (s *IconService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalIcons, totalCategories int
	s.db.QueryRow(`SELECT COUNT(*) FROM icons`).Scan(&totalIcons)
	s.db.QueryRow(`SELECT COUNT(DISTINCT package || '/' || category) FROM icons`).Scan(&totalCategories)

	// Package counts
	rows, err := s.db.Query(`SELECT package, COUNT(*) FROM icons GROUP BY package ORDER BY COUNT(*) DESC`)
	if err != nil {
		rows.Close()
	}
	pkgCount := make(map[string]int)
	if err == nil {
		for rows.Next() {
			var p string
			var c int
			rows.Scan(&p, &c)
			pkgCount[p] = c
		}
		rows.Close()
	}

	return map[string]interface{}{
		"totalIcons": totalIcons,
		"packages":   pkgCount,
		"categories": totalCategories,
	}
}

// Close closes the database
func (s *IconService) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

// applySVGColor replaces or adds fill color to SVG
func applySVGColor(svg []byte, color string) []byte {
	content := string(svg)

	// Replace currentColor with the specified color
	content = strings.ReplaceAll(content, "currentColor", color)

	// Remove fill attributes but preserve fill="none" (transparent areas)
	// Go regexp doesn't support lookahead, so use a placeholder approach
	re := regexp.MustCompile(` fill="[^"]*"`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		if match == ` fill="none"` {
			return match
		}
		return ""
	})

	// Add fill to the root <svg> tag
	svgTagRe := regexp.MustCompile(`(?i)(<svg[^>]*)(>)`)
	if svgTagRe.MatchString(content) {
		content = svgTagRe.ReplaceAllString(content, fmt.Sprintf(`$1 fill="%s"$2`, color))
	}

	return []byte(content)
}

// sanitizeFilename makes a string safe for use as a filename
func sanitizeFilename(s string) string {
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			result = append(result, r)
		}
	}
	safe := string(result)
	if safe == "" {
		safe = "icon"
	}
	return safe
}

// convertSVGToPNG renders SVG to PNG at the specified pixel size
func convertSVGToPNG(svgContent []byte, name string, size int, tmpDir string) (string, error) {
	svgStr := string(svgContent)

	// Replace currentColor with black since oksvg does not support it
	svgStr = strings.ReplaceAll(svgStr, "currentColor", "#000000")

	// Remove unsupported CSS units from width/height so oksvg can parse them.
	// oksvg only supports cm, mm, px, pt but not em, rem, %, etc.
	unitRe := regexp.MustCompile(` (width|height)="([0-9.]*)(em|rem|%|ex|ch|vw|vh|vmin|vmax)"`)
	svgStr = unitRe.ReplaceAllString(svgStr, "")

	// First parse to get original viewBox
	icon, err := oksvg.ReadIconStream(strings.NewReader(svgStr))
	if err != nil {
		return "", fmt.Errorf("failed to parse SVG: %w", err)
	}

	if icon.ViewBox.W == 0 || icon.ViewBox.H == 0 {
		icon.ViewBox.W = 24
		icon.ViewBox.H = 24
	}

	// Calculate scale factor
	scale := float64(size) / icon.ViewBox.W

	// Pre-scale stroke-width in the SVG source since oksvg's transform
	// does not scale stroke width (only coordinates are transformed)
	if scale != 1.0 {
		strokeRe := regexp.MustCompile(`stroke-width="([0-9]*\.?[0-9]+)"`)
		svgStr = strokeRe.ReplaceAllStringFunc(svgStr, func(match string) string {
			submatches := strokeRe.FindStringSubmatch(match)
			if len(submatches) < 2 {
				return match
			}
			origWidth := 0.0
			fmt.Sscanf(submatches[1], "%f", &origWidth)
			newWidth := origWidth * scale
			return fmt.Sprintf(`stroke-width="%.2f"`, newWidth)
		})
	}

	// Re-parse with scaled stroke-width
	icon, err = oksvg.ReadIconStream(strings.NewReader(svgStr))
	if err != nil {
		return "", fmt.Errorf("failed to parse SVG: %w", err)
	}

	if icon.ViewBox.W == 0 || icon.ViewBox.H == 0 {
		icon.ViewBox.W = 24
		icon.ViewBox.H = 24
	}

	// SetTarget maps the viewBox to the target rectangle with proper scaling
	icon.SetTarget(0, 0, float64(size), float64(size))

	// Create rasterizer
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	scanner := rasterx.NewScannerGV(size, size, img, img.Bounds())
	raster := rasterx.NewDasher(size, size, scanner)

	icon.Draw(raster, 1.0)

	// Write PNG
	pngPath := filepath.Join(tmpDir, fmt.Sprintf("%s_%d.png", name, size))
	f, err := os.Create(pngPath)
	if err != nil {
		return "", fmt.Errorf("failed to create PNG file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return "", fmt.Errorf("failed to encode PNG: %w", err)
	}

	return pngPath, nil
}

// toJSON is a helper for MCP responses
func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return `{"error": "` + err.Error() + `"}`
	}
	return string(b)
}
