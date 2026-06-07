<script setup>
import { ref, onMounted } from 'vue'
import * as IconService from '../../bindings/iconstore/iconservice.js'
import MCPDialog from './MCPDialog.vue'

const props = defineProps({
  stats: Object,
  selectedPackage: String
})

const emit = defineEmits(['select-package'])
const packages = ref([])
const showMCPDialog = ref(false)

onMounted(async () => {
  try {
    packages.value = await IconService.GetPackages()
  } catch (err) {
    console.error('Failed to load packages:', err)
  }
})

function selectPackage(pkg) {
  emit('select-package', pkg === props.selectedPackage ? '' : pkg)
}
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar-section">
      <h3 class="section-title">All Icons</h3>
      <button
        class="category-item"
        :class="{ active: !selectedPackage }"
        @click="selectPackage('')"
      >
        <span class="cat-name">All</span>
        <span class="cat-count" v-if="stats.totalIcons">{{ stats.totalIcons }}</span>
      </button>
    </div>

    <div class="sidebar-section" v-if="packages.length">
      <h3 class="section-title">Packages</h3>
      <div class="categories-list">
        <button
          v-for="pkg in packages"
          :key="pkg.name"
          class="category-item"
          :class="{ active: selectedPackage === pkg.name }"
          @click="selectPackage(pkg.name)"
        >
          <span class="cat-name">{{ pkg.name }}</span>
          <span class="cat-count">{{ pkg.count }}</span>
        </button>
      </div>
    </div>

    <div class="sidebar-footer">
      <button class="mcp-btn" @click="showMCPDialog = true">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/>
        </svg>
        <span>MCP Server</span>
      </button>
    </div>
  </aside>

  <MCPDialog v-model:visible="showMCPDialog" />
</template>
