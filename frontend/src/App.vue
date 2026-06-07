<script setup>
import { ref, onMounted, computed, provide } from 'vue'
import IconGrid from './components/IconGrid.vue'
import IconDetail from './components/IconDetail.vue'
import SearchBar from './components/SearchBar.vue'
import Sidebar from './components/Sidebar.vue'
import TitleBar from './components/TitleBar.vue'
import Toast from './components/Toast.vue'
import * as IconService from '../bindings/iconstore/iconservice.js'

const searchQuery = ref('')
const selectedPackage = ref('')
const currentPage = ref(1)
const pageSize = 96
const searchResult = ref({ icons: [], total: 0, page: 1, pageSize: 96 })
const selectedIcon = ref(null)
const showDetail = ref(false)
const stats = ref({ totalIcons: 0, packages: {}, categories: 0 })
const loading = ref(false)

const toastVisible = ref(false)
const toastMessage = ref('')
const toastType = ref('success')

function showToast(message, type = 'success') {
  toastMessage.value = message
  toastType.value = type
  toastVisible.value = true
}

provide('showToast', showToast)

const totalPages = computed(() => Math.ceil(searchResult.value.total / pageSize))

async function doSearch() {
  loading.value = true
  try {
    const result = await IconService.Search(searchQuery.value, selectedPackage.value, '', currentPage.value, pageSize)
    searchResult.value = result
  } catch (err) {
    console.error('Search failed:', err)
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  try {
    stats.value = await IconService.GetStats()
  } catch (err) {
    console.error('Failed to load stats:', err)
  }
}

function onSearch(query) {
  searchQuery.value = query
  currentPage.value = 1
  doSearch()
}

function onPackageSelect(pkg) {
  selectedPackage.value = pkg
  currentPage.value = 1
  doSearch()
}

function onPageChange(page) {
  currentPage.value = page
  doSearch()
}

function onIconClick(icon) {
  selectedIcon.value = icon
  showDetail.value = true
}

function onCloseDetail() {
  showDetail.value = false
  selectedIcon.value = null
}

onMounted(() => {
  loadStats()
  doSearch()
})
</script>

<template>
  <div class="app-layout">
    <TitleBar />
    <Toast :message="toastMessage" :type="toastType" v-model:visible="toastVisible" />
    <div class="app-body">
      <Sidebar
        :stats="stats"
        :selectedPackage="selectedPackage"
        @select-package="onPackageSelect"
      />
      <div class="main-content">
        <SearchBar @search="onSearch" />
        <IconGrid
          :icons="searchResult.icons"
          :loading="loading"
          @icon-click="onIconClick"
        />
        <div class="toolbar">
          <span class="result-count">
            {{ searchResult.total }} icons found
          </span>
          <div class="pagination" v-if="totalPages > 1">
            <button
              class="page-btn"
              :disabled="currentPage <= 1"
              @click="onPageChange(currentPage - 1)"
            >&laquo;</button>
            <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
            <button
              class="page-btn"
              :disabled="currentPage >= totalPages"
              @click="onPageChange(currentPage + 1)"
            >&raquo;</button>
          </div>
        </div>
      </div>
      <IconDetail
        v-if="showDetail && selectedIcon"
        :icon="selectedIcon"
        @close="onCloseDetail"
      />
    </div>
  </div>
</template>
