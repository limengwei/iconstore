<script setup>
import { ref, watch } from 'vue'
import * as IconService from '../../bindings/iconstore/iconservice.js'

const props = defineProps({
  selectedPackage: String,
  selectedSubCategory: String
})

const emit = defineEmits(['select-subcategory'])
const subCategories = ref([])

async function loadSubCategories(pkg) {
  if (!pkg) {
    subCategories.value = []
    return
  }
  try {
    subCategories.value = await IconService.GetCategoriesByPackage(pkg)
  } catch (err) {
    console.error('Failed to load subcategories:', err)
    subCategories.value = []
  }
}

watch(() => props.selectedPackage, (pkg) => {
  loadSubCategories(pkg)
}, { immediate: true })

function selectSub(cat) {
  emit('select-subcategory', cat === props.selectedSubCategory ? '' : cat)
}
</script>

<template>
  <div class="sub-categories" v-if="subCategories.length">
    <button
      class="sub-cat-item"
      :class="{ active: !selectedSubCategory }"
      @click="selectSub('')"
    >
      All
    </button>
    <button
      v-for="cat in subCategories"
      :key="cat.name"
      class="sub-cat-item"
      :class="{ active: selectedSubCategory === cat.name }"
      @click="selectSub(cat.name)"
    >
      {{ cat.name }}
      <span class="sub-cat-count">{{ cat.count }}</span>
    </button>
  </div>
</template>
