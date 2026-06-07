<script setup>
import { ref, onMounted, computed, inject } from 'vue'
import * as IconService from '../../bindings/iconstore/iconservice.js'
import * as DialogService from '../../bindings/iconstore/dialogservice.js'

const props = defineProps({
  icon: Object
})

const emit = defineEmits(['close'])

const showToast = inject('showToast')

const svgContent = ref('')
const exportFormat = ref('svg')
const exportSize = ref(24)
const exportColor = ref('#AA00FF')
const showColorPicker = ref(false)
const exporting = ref(false)

const sizeOptions = [16, 20, 24, 32, 48, 64, 128, 256, 512]

const previewColor = computed(() => {
  return exportColor.value || ''
})

const previewSvg = computed(() => {
  return svgContent.value
})

onMounted(async () => {
  try {
    svgContent.value = await IconService.GetIconSVG(props.icon.id)
  } catch (err) {
    console.error('Failed to load icon SVG:', err)
  }
})

async function onExport() {
  exporting.value = true
  try {
    const path = await DialogService.SaveIconDialog(
      props.icon.id,
      exportFormat.value,
      exportSize.value,
      exportColor.value
    )
    if (path) {
      showToast('下载完成')
    }
  } catch (err) {
    showToast('Export failed: ' + err.message, 'error')
  } finally {
    exporting.value = false
  }
}

async function copySVG() {
  try {
    await navigator.clipboard.writeText(svgContent.value)
    showToast('SVG 已复制到剪贴板')
  } catch (err) {
    showToast('复制失败: ' + err.message, 'error')
  }
}

function setColor(color) {
  exportColor.value = color
  showColorPicker.value = false
}

// const presetColors = [
//   '#000000', '#FFFFFF', '#FF5722', '#E91E63', '#9C27B0',
//   '#2196F3', '#00BCD4', '#4CAF50', '#FFC107', '#FF9800',
//   '#795548', '#607D8B', '#F44336', '#3F51B5', '#009688'
// ]

const presetColors = [
  '#000000', '#FFFFFF', '#D50000', '#C51162', '#AA00FF',
  '#6200EA', '#304FFE', '#2962FF', '#0091EA', '#00B8D4',
  '#00BFA5', '#00C853', '#64DD17', '#AEEA00', '#FFD600',
  '#FFAB00','#FF6D00','#DD2C00','#3E2723','#212121','#263238'
]


</script>

<template>
  <div class="detail-overlay" @click.self="emit('close')">
    <div class="detail-panel">
      <div class="detail-header">
        <h2>{{ icon.name }}</h2>
        <button class="close-btn" @click="emit('close')">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="detail-body">
        <div class="preview-section">
          <div class="preview-large" :style="previewColor ? { color: previewColor } : {}" v-html="previewSvg"></div>
          <div class="preview-sizes" v-if="previewSvg">
            <div v-for="s in [16, 24, 32, 48]" :key="s" class="preview-size-item">
              <div class="preview-box" :style="{ width: s + 'px', height: s + 'px', color: previewColor || undefined }" v-html="previewSvg"></div>
              <span>{{ s }}px</span>
            </div>
          </div>
        </div>

        <div class="info-section">
          <div class="info-row">
            <span class="info-label">Package</span>
            <span class="info-value">{{ icon.package }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">Tags</span>
            <div class="tags-list">
              <span v-for="tag in (icon.tags || '').split(',').slice(0, 8)" :key="tag" class="tag">{{ tag.trim() }}</span>
            </div>
          </div>
        </div>

        <div class="export-section">
          <h3>Export</h3>

          <div class="export-options">
            <div class="option-group">
              <label>Format</label>
              <div class="format-btns">
                <button
                  :class="{ active: exportFormat === 'svg' }"
                  @click="exportFormat = 'svg'"
                >SVG</button>
                <button
                  :class="{ active: exportFormat === 'png' }"
                  @click="exportFormat = 'png'"
                >PNG</button>
              </div>
            </div>

            <div class="option-group" v-if="exportFormat === 'png'">
              <label>Size (px)</label>
              <div class="size-btns">
                <button
                  v-for="s in sizeOptions"
                  :key="s"
                  :class="{ active: exportSize === s }"
                  @click="exportSize = s"
                >{{ s }}</button>
              </div>
            </div>

            <div class="option-group">
              <label>Color</label>
              <div class="color-row">
                <button
                  class="color-preview"
                  :style="{ background: exportColor || 'currentColor' }"
                  @click="showColorPicker = !showColorPicker"
                ></button>
                <input
                  type="text"
                  v-model="exportColor"
                  placeholder="None (default)"
                  class="color-input"
                />
                <button
                  v-if="exportColor"
                  class="clear-color"
                  @click="exportColor = ''"
                >Reset</button>
              </div>
              <div  class="color-palette">
                <button
                  v-for="c in presetColors"
                  :key="c"
                  class="color-swatch"
                  :style="{ background: c }"
                  @click="setColor(c)"
                ></button>
              </div>
            </div>
          </div>

          <div class="export-actions">
            <button class="btn-primary" @click="onExport" :disabled="exporting">
              {{ exporting ? 'Exporting...' : 'Download ' + exportFormat.toUpperCase() }}
            </button>
            <button class="btn-secondary" @click="copySVG">
              Copy SVG Code
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
