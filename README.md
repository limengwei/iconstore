# IconStore

English | [中文](README_CN.md)

A cross-platform SVG icon search, browse, and download tool built with [Wails v3](https://v3.wails.io/) (Go + Vue 3). It comes with a built-in SQLite database for full-text icon search and includes an MCP (Model Context Protocol) server for AI assistant integration.

## Features

- **Icon Search** — Full-text search powered by SQLite FTS5, supporting keyword matching across icon name, category, tags, and package
- **Browse by Package/Category** — Sidebar navigation to browse icons organized by packages and sub-categories
- **Icon Preview** — View icons at multiple sizes (16/24/32/48px) with real-time color preview
- **Export** — Download icons as SVG or PNG with custom size and color; native system save dialog
- **Copy SVG** — One-click copy of raw SVG code to clipboard
- **MCP Server** — Built-in MCP server (port 9393) for integration with AI tools like Claude Desktop, Cursor, etc.
- **Cross-Platform** — Runs on Windows, macOS, and Linux via Wails v3

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go, SQLite (modernc.org/sqlite), FTS5 |
| Frontend | Vue 3, Vite |
| Desktop Framework | Wails v3 |
| SVG Rendering | oksvg + rasterx (PNG export) |

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) >= 1.25
- [Node.js](https://nodejs.org/) (for frontend build)
- [Wails v3 CLI](https://v3.wails.io/)

### Development

```bash
wails3 dev
```

### Build

```bash
wails3 build
```

### MCP Standalone Mode

Run IconStore as an MCP server without the GUI:

```bash
iconstore --mcp          # default port 9393
iconstore --mcp 8080     # custom port
```

## MCP Integration

The embedded MCP server starts automatically when the desktop app runs, or can be launched standalone with `--mcp`.

### Available Tools

| Tool | Description |
|------|-------------|
| `search_icons` | Search icons by keyword, with optional category filter and pagination |
| `get_icon` | Get raw SVG content by icon ID |
| `export_icon` | Export icon as SVG or PNG with custom color and size |

### Client Configuration

Add to your MCP client config (e.g. Claude Desktop, Cursor):

```json
{
  "mcpServers": {
    "iconstore": {
      "url": "http://127.0.0.1:9393/mcp"
    }
  }
}
```

## Project Structure

```
├── main.go              # Entry point, desktop app & MCP server startup
├── icon_service.go      # Icon search, export, SQLite management
├── mcp_server.go        # MCP JSON-RPC server implementation
├── dialog_service.go    # Native file save dialog
├── iconstore.db         # SQLite database (icons + FTS index)
├── frontend/
│   ├── src/
│   │   ├── App.vue             # Main layout with search & pagination
│   │   ├── components/
│   │   │   ├── IconGrid.vue    # Icon grid display
│   │   │   ├── IconDetail.vue  # Icon detail panel with export options
│   │   │   ├── SearchBar.vue   # Search input
│   │   │   ├── Sidebar.vue     # Package/category navigation
│   │   │   ├── SubCategories.vue
│   │   │   ├── TitleBar.vue    # Custom frameless title bar
│   │   │   ├── MCPDialog.vue   # MCP configuration dialog
│   │   │   └── Toast.vue       # Notification toast
│   │   └── main.js
│   ├── bindings/               # Wails auto-generated JS bindings
│   └── package.json
├── build/                      # Platform-specific build configs
│   ├── windows/
│   ├── darwin/
│   ├── linux/
│   ├── android/
│   └── ios/
└── Taskfile.yml
```

## License

This project uses icons from various open-source icon libraries. See individual package licenses for details.
