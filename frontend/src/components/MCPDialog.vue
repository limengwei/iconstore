<script setup>
import { ref } from 'vue'

const props = defineProps({
  visible: Boolean
})

const emit = defineEmits(['update:visible'])

const activeTab = ref('tools')

const tools = [
  { name: 'search_icons', desc: 'Search for SVG icons by keyword. Returns matching icons with name, category, and tags.' },
  { name: 'get_icon', desc: "Get an icon's SVG content by its ID. Returns the raw SVG content." },
  { name: 'export_icon', desc: 'Export an icon with custom color and format. Returns the exported file path.' }
]

const mcpCommand = 'iconstore --mcp'

function copyCommand() {
  navigator.clipboard.writeText(mcpCommand)
}
</script>

<template>
  <Transition name="modal">
    <div v-if="visible" class="mcp-overlay" @click.self="emit('update:visible', false)">
      <div class="mcp-dialog">
        <div class="mcp-header">
          <div class="mcp-title">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/>
            </svg>
            <h2>MCP Server</h2>
          </div>
          <button class="close-btn" @click="emit('update:visible', false)">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>

        <div class="mcp-tabs">
          <button :class="{ active: activeTab === 'tools' }" @click="activeTab = 'tools'">Tools</button>
          <button :class="{ active: activeTab === 'config' }" @click="activeTab = 'config'">Configuration</button>
        </div>

        <div class="mcp-body">
          <div v-if="activeTab === 'tools'" class="tools-list">
            <div v-for="tool in tools" :key="tool.name" class="tool-card">
              <div class="tool-name">{{ tool.name }}</div>
              <div class="tool-desc">{{ tool.desc }}</div>
            </div>
          </div>

          <div v-if="activeTab === 'config'" class="config-section">
            <p class="config-hint">Start the MCP server:</p>
            <div class="code-block">
              <pre>iconstore --mcp [port]

# Default port: 9393
# Example: iconstore --mcp 9393</pre>
            </div>
            <p class="config-hint" style="margin-top: 14px">Add to your MCP client configuration (e.g. Claude Desktop, Cursor):</p>
            <div class="code-block">
              <pre>{
  "mcpServers": {
    "iconstore": {
      "url": "http://127.0.0.1:9393/mcp"
    }
  }
}</pre>
            </div>
            <div class="config-note">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/>
              </svg>
              Make sure <code>iconstore</code> is in your PATH, or use the full path to the executable.
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.mcp-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
}
.mcp-dialog {
  width: 560px;
  max-width: 90vw;
  max-height: 80vh;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
}
.mcp-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border);
}
.mcp-title {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--accent);
}
.mcp-title h2 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}
.mcp-tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--border);
  padding: 0 20px;
}
.mcp-tabs button {
  padding: 10px 16px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-muted);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}
.mcp-tabs button:hover {
  color: var(--text-primary);
}
.mcp-tabs button.active {
  color: var(--accent);
  border-bottom-color: var(--accent);
}
.mcp-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}
.tools-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.tool-card {
  padding: 14px 16px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 10px;
}
.tool-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--accent);
  font-family: monospace;
  margin-bottom: 4px;
}
.tool-desc {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
}
.config-hint {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 12px;
}
.code-block {
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 16px;
  overflow-x: auto;
}
.code-block pre {
  font-family: 'Cascadia Code', 'Fira Code', monospace;
  font-size: 13px;
  color: var(--text-primary);
  line-height: 1.6;
  white-space: pre;
}
.config-note {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-top: 14px;
  padding: 10px 14px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.5;
}
.config-note svg {
  flex-shrink: 0;
  margin-top: 1px;
}
.config-note code {
  background: var(--bg-primary);
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 12px;
  color: var(--accent);
}
.modal-enter-active {
  animation: fadeIn 0.2s ease;
}
.modal-leave-active {
  animation: fadeIn 0.2s ease reverse;
}
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>
