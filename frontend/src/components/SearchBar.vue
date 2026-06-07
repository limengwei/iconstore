<script setup>
import { ref } from 'vue'

const query = ref('')
const emit = defineEmits(['search'])

function onInput() {
  // Debounced search handled by parent
}

function onSubmit() {
  emit('search', query.value)
}

function onClear() {
  query.value = ''
  emit('search', '')
}
</script>

<template>
  <div class="search-bar">
    <div class="search-input-wrapper">
      <svg class="search-icon" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"/>
        <line x1="21" y1="21" x2="16.65" y2="16.65"/>
      </svg>
      <input
        type="text"
        v-model="query"
        placeholder="Search icons by name, tag, or category..."
        class="search-input"
        @keyup.enter="onSubmit"
        @input="onInput"
      />
      <button v-if="query" class="clear-btn" @click="onClear" title="Clear">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
        </svg>
      </button>
    </div>
  </div>
</template>
