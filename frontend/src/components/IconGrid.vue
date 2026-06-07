<script setup>
import { ref, watch, onMounted } from 'vue'
import * as IconService from '../../bindings/iconstore/iconservice.js'

const props = defineProps({
  icons: Array,
  loading: Boolean
})

const emit = defineEmits(['icon-click'])
const svgCache = ref({})

async function loadSvg(icon) {
  if (svgCache.value[icon.id]) return
  try {
    let svg = await IconService.GetIconSVG(icon.id)
    // 清洗 SVG：统一颜色处理
    svg = normalizeSvg(svg)
    svgCache.value[icon.id] = svg
  } catch (err) {
    console.error('Failed to load SVG:', err)
  }
}

function normalizeSvg(svg) {
  // 移除硬编码颜色，让图标继承 CSS color
  svg = svg.replace(/ fill="#[0-9A-Fa-f]{3,8}"/g, '')
  // 移除纯背景矩形路径（仅匹配已知的空矩形）
  svg = svg.replace(/<path[^>]*d="M0 0h24v24H0z"[^>]*\/?>/g, '')
  return svg
}

watch(() => props.icons, (newIcons) => {
  if (newIcons) {
    newIcons.forEach(icon => loadSvg(icon))
  }
}, { immediate: true })

function onClick(icon) {
  emit('icon-click', icon)
}
</script>

<template>
  <div class="icon-grid-wrapper">
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <span>Loading icons...</span>
    </div>
    <div v-else-if="icons.length === 0" class="empty-state">
      <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.5">
        <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
      </svg>
      <p>No icons found. Try a different search term.</p>
    </div>
    <div v-else class="icon-grid">
      <div
        v-for="icon in icons"
        :key="icon.id"
        class="icon-card"
        @click="onClick(icon)"
        :title="icon.name"
      >
        <div class="icon-preview">
          <div v-if="svgCache[icon.id]" class="icon-svg" v-html="svgCache[icon.id]"></div>
          <div v-else class="icon-placeholder">
            <div class="spinner-small"></div>
          </div>
        </div>
        <div class="icon-label" :title="icon.name">{{ icon.name }}</div>
      </div>
    </div>
  </div>
</template>
