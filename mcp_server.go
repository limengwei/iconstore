package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// MCP JSON-RPC types

type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type MCPTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type MCPToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// MCPServer provides MCP protocol server over HTTP
type MCPServer struct {
	iconService *IconService
}

// NewMCPServer creates a new MCP server
func NewMCPServer(iconService *IconService) *MCPServer {
	return &MCPServer{
		iconService: iconService,
	}
}

// HandleHTTP handles JSON-RPC requests over HTTP POST
func (s *MCPServer) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		json.NewEncoder(w).Encode(MCPResponse{
			JSONRPC: "2.0",
			Error:   &MCPError{Code: -32700, Message: "Read error"},
		})
		return
	}

	var req MCPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		json.NewEncoder(w).Encode(MCPResponse{
			JSONRPC: "2.0",
			Error:   &MCPError{Code: -32700, Message: "Parse error: " + err.Error()},
		})
		return
	}

	resp := s.handleRequest(req)
	json.NewEncoder(w).Encode(resp)
}

// HandleSSE handles SSE connections for streamable HTTP transport
func (s *MCPServer) HandleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "event: endpoint\ndata: /mcp\n\n")
	flusher.Flush()

	<-r.Context().Done()
}

func (s *MCPServer) handleRequest(req MCPRequest) MCPResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "notifications/initialized":
		return MCPResponse{}
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "resources/list":
		return s.handleResourcesList(req)
	case "ping":
		return MCPResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{}}
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &MCPError{Code: -32601, Message: "Method not found: " + req.Method},
		}
	}
}

func (s *MCPServer) handleInitialize(req MCPRequest) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": false,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "iconstore",
				"version": "1.0.4",
			},
		},
	}
}

func (s *MCPServer) handleToolsList(req MCPRequest) MCPResponse {
	tools := []MCPTool{
		{
			Name:        "search_icons",
			Description: "Search for SVG icons by keyword. Returns matching icons with name, category, and tags.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search keyword (matches icon name, tags, category)",
					},
					"category": map[string]interface{}{
						"type":        "string",
						"description": "Filter by category (e.g. 'material-design-icons/navigation')",
					},
					"page": map[string]interface{}{
						"type":        "integer",
						"description": "Page number (1-based), default 1",
					},
					"pageSize": map[string]interface{}{
						"type":        "integer",
						"description": "Results per page, default 20",
					},
				},
			},
		},
		{
			Name:        "get_icon",
			Description: "Get an icon's SVG content by its ID. Returns the raw SVG content.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Icon ID (from search results, format: package/category/name)",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "export_icon",
			Description: "Export an icon with custom color and format. Returns the exported file path.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Icon ID",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "Export format: 'svg' or 'png'",
						"enum":        []string{"svg", "png"},
					},
					"size": map[string]interface{}{
						"type":        "integer",
						"description": "Size in pixels (for PNG), default 24",
					},
					"color": map[string]interface{}{
						"type":        "string",
						"description": "Fill color as hex (e.g. '#FF5722'), optional",
					},
					"directory": map[string]interface{}{
						"type":        "string",
						"description": "Output directory path. If not specified, uses system temp directory.",
					},
				},
				"required": []string{"id"},
			},
		},
	}
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]interface{}{"tools": tools},
	}
}

func (s *MCPServer) handleToolsCall(req MCPRequest) MCPResponse {
	var params MCPToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &MCPError{Code: -32602, Message: "Invalid params: " + err.Error()},
		}
	}

	var contents []MCPContent

	switch params.Name {
	case "search_icons":
		contents = s.toolSearchIcons(params.Arguments)
	case "get_icon":
		contents = s.toolGetIcon(params.Arguments)
	case "export_icon":
		contents = s.toolExportIcon(params.Arguments)
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &MCPError{Code: -32602, Message: "Unknown tool: " + params.Name},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]interface{}{"content": contents},
	}
}

func (s *MCPServer) toolSearchIcons(args map[string]interface{}) []MCPContent {
	query, _ := args["query"].(string)
	category, _ := args["category"].(string)
	page := intArg(args, "page", 1)
	pageSize := intArg(args, "pageSize", 20)

	result := s.iconService.Search(query, "", category, page, pageSize)

	text := fmt.Sprintf("Found %d icons", result.Total)
	if query != "" {
		text += fmt.Sprintf(" matching '%s'", query)
	}
	if category != "" {
		text += fmt.Sprintf(" in category '%s'", category)
	}
	text += fmt.Sprintf(" (page %d/%d):\n\n", result.Page, (result.Total+pageSize-1)/pageSize)

	for _, icon := range result.Icons {
		text += fmt.Sprintf("- [%s] %s (tags: %s)\n", icon.ID, icon.Name, icon.Tags)
	}

	if result.Total > pageSize {
		text += fmt.Sprintf("\n... and %d more results. Use page parameter to see more.", result.Total-pageSize*page)
	}

	return []MCPContent{{Type: "text", Text: text}}
}

func (s *MCPServer) toolGetIcon(args map[string]interface{}) []MCPContent {
	id, _ := args["id"].(string)
	if id == "" {
		return []MCPContent{{Type: "text", Text: "Error: id parameter is required"}}
	}

	svgContent, err := s.iconService.GetIconSVG(id)
	if err != nil {
		return []MCPContent{{Type: "text", Text: "Error: " + err.Error()}}
	}

	icon, _ := s.iconService.GetIcon(id)

	return []MCPContent{
		{Type: "text", Text: fmt.Sprintf("Icon: %s\nCategory: %s\nPackage: %s\n\n%s", icon.Name, icon.Category, icon.Package, svgContent)},
	}
}

func (s *MCPServer) toolExportIcon(args map[string]interface{}) []MCPContent {
	id, _ := args["id"].(string)
	if id == "" {
		return []MCPContent{{Type: "text", Text: "Error: id parameter is required"}}
	}

	format, _ := args["format"].(string)
	if format == "" {
		format = "svg"
	}
	size := intArg(args, "size", 24)
	color, _ := args["color"].(string)
	directory, _ := args["directory"].(string)

	outPath, err := s.iconService.ExportIcon(id, ExportOptions{
		Format:    format,
		Size:      size,
		Color:     color,
		OutputDir: directory,
	})
	if err != nil {
		return []MCPContent{{Type: "text", Text: "Error: " + err.Error()}}
	}
	return []MCPContent{
		{Type: "text", Text: outPath},
	}
}

func (s *MCPServer) handleResourcesList(req MCPRequest) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]interface{}{"resources": []interface{}{}},
	}
}

func intArg(args map[string]interface{}, key string, defaultVal int) int {
	v, ok := args[key]
	if !ok {
		return defaultVal
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return defaultVal
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
