<script setup lang="ts">
import type { Component } from 'vue'
import { ChevronDown, Filter, Search } from 'lucide-vue-next'
import { computed, onMounted, onUnmounted, ref } from 'vue'

export interface WorkspaceDirectorySortOption {
  key: string
  label: string
  icon?: Component
}

const props = withDefaults(
  defineProps<{
    modelValue: string
    total: number
    selectedSortLabel: string
    sortOptions: WorkspaceDirectorySortOption[]
    searchPlaceholder?: string
    filterButtonLabel?: string
    sortCaption?: string
    totalSuffix?: string
    showSearch?: boolean
    filterPanelKicker?: string
    filterPanelTitle?: string
    filterPanelWidth?: string
    resetLabel?: string
    resetDisabled?: boolean
    showFilter?: boolean
    showTotal?: boolean
  }>(),
  {
    searchPlaceholder: '输入关键词检索...',
    filterButtonLabel: '筛选',
    sortCaption: '排序:',
    totalSuffix: '项',
    showSearch: true,
    filterPanelKicker: 'Filter Stack',
    filterPanelTitle: '高级筛选',
    filterPanelWidth: '24rem',
    resetLabel: '清空筛选',
    resetDisabled: false,
    showFilter: true,
    showTotal: true,
  }
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
  selectSort: [option: WorkspaceDirectorySortOption]
  resetFilters: []
}>()

const isFilterOpen = ref(false)
const isSortOpen = ref(false)
const filterToggleRef = ref<HTMLButtonElement | null>(null)
const filterPanelRef = ref<HTMLDivElement | null>(null)
const sortButtonRef = ref<HTMLButtonElement | null>(null)
const sortMenuRef = ref<HTMLDivElement | null>(null)

const hasSortOptions = computed(() => props.sortOptions.length > 0)

function closeFilter(): void {
  isFilterOpen.value = false
}

function closeSort(): void {
  isSortOpen.value = false
}

function handleSearchInput(event: Event): void {
  const target = event.target
  emit('update:modelValue', target instanceof HTMLInputElement ? target.value : '')
}

function toggleFilter(): void {
  isFilterOpen.value = !isFilterOpen.value
  if (isFilterOpen.value) {
    isSortOpen.value = false
  }
}

function toggleSort(): void {
  if (!hasSortOptions.value) {
    return
  }
  isSortOpen.value = !isSortOpen.value
  if (isSortOpen.value) {
    isFilterOpen.value = false
  }
}

function handleSelectSort(option: WorkspaceDirectorySortOption): void {
  emit('selectSort', option)
  closeSort()
}

let removeWindowListeners: (() => void) | null = null

onMounted(() => {
  const handleClickOutside = (event: MouseEvent) => {
    const target = event.target
    if (!(target instanceof Node)) {
      closeFilter()
      closeSort()
      return
    }

    const clickedInsideFilter =
      filterToggleRef.value?.contains(target) || filterPanelRef.value?.contains(target)
    const clickedInsideSort =
      sortButtonRef.value?.contains(target) || sortMenuRef.value?.contains(target)

    if (!clickedInsideFilter) {
      closeFilter()
    }

    if (!clickedInsideSort) {
      closeSort()
    }
  }

  const handleEscape = (event: KeyboardEvent) => {
    if (event.key === 'Escape') {
      closeFilter()
      closeSort()
    }
  }

  window.addEventListener('click', handleClickOutside)
  window.addEventListener('keydown', handleEscape)
  removeWindowListeners = () => {
    window.removeEventListener('click', handleClickOutside)
    window.removeEventListener('keydown', handleEscape)
  }
})

onUnmounted(() => {
  removeWindowListeners?.()
})
</script>

<template>
  <div class="workspace-directory-toolbar">
    <div class="workspace-directory-toolbar__main">
      <label
        v-if="showSearch"
        class="workspace-directory-toolbar__search"
      >
        <Search class="workspace-directory-toolbar__search-icon h-3.5 w-3.5" />
        <input
          :value="modelValue"
          type="text"
          class="workspace-directory-toolbar__search-input"
          :placeholder="searchPlaceholder"
          @input="handleSearchInput"
        >
      </label>

      <button
        v-if="showFilter"
        ref="filterToggleRef"
        type="button"
        class="workspace-directory-toolbar__filter-toggle"
        :class="{ 'workspace-directory-toolbar__filter-toggle--active': isFilterOpen }"
        @click.stop="toggleFilter"
      >
        <Filter class="h-3.5 w-3.5" />
        {{ filterButtonLabel }}
      </button>
    </div>

    <div class="workspace-directory-toolbar__meta">
      <div
        v-if="hasSortOptions"
        class="workspace-directory-toolbar__sort"
      >
        <span class="workspace-directory-toolbar__sort-caption">{{ sortCaption }}</span>
        <button
          ref="sortButtonRef"
          type="button"
          class="workspace-directory-toolbar__sort-button"
          @click.stop="toggleSort"
        >
          <span class="workspace-directory-toolbar__sort-label">{{ selectedSortLabel }}</span>
          <ChevronDown
            class="h-3.5 w-3.5 transition-transform"
            :class="{ 'rotate-180': isSortOpen }"
          />
        </button>

        <div
          v-if="isSortOpen"
          ref="sortMenuRef"
          class="workspace-directory-toolbar__sort-menu"
        >
          <div class="workspace-directory-toolbar__menu-title">
            Sort Strategy
          </div>
          <div class="workspace-directory-toolbar__menu-list">
            <button
              v-for="option in sortOptions"
              :key="option.key"
              type="button"
              class="workspace-directory-toolbar__menu-item"
              :class="{
                'workspace-directory-toolbar__menu-item--active':
                  option.label === selectedSortLabel,
              }"
              @click="handleSelectSort(option)"
            >
              <span class="workspace-directory-toolbar__menu-item-content">
                <component
                  :is="option.icon"
                  v-if="option.icon"
                  class="h-3.5 w-3.5"
                />
                {{ option.label }}
              </span>
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="showTotal"
        class="workspace-directory-toolbar__count-pill"
      >
        共 <span class="workspace-directory-toolbar__count-value">{{ total }}</span> {{ totalSuffix }}
      </div>
    </div>

    <div
      v-if="showFilter && isFilterOpen"
      ref="filterPanelRef"
      class="workspace-directory-toolbar__filter-panel"
      :style="{ width: filterPanelWidth }"
    >
      <div class="workspace-directory-toolbar__filter-panel-header">
        <div>
          <div class="workspace-overline">
            {{ filterPanelKicker }}
          </div>
          <h3 class="workspace-directory-toolbar__filter-panel-title">
            {{ filterPanelTitle }}
          </h3>
        </div>
        <button
          type="button"
          class="workspace-directory-toolbar__filter-reset"
          :disabled="resetDisabled"
          @click="emit('resetFilters')"
        >
          {{ resetLabel }}
        </button>
      </div>

      <slot
        name="filter-panel"
        :close="closeFilter"
      />
    </div>
  </div>
</template>

<style scoped>
.workspace-directory-toolbar {
  --workspace-toolbar-surface: var(--color-bg-surface);
  --workspace-toolbar-surface-subtle: var(--color-bg-elevated);
  --workspace-toolbar-control-border: var(--color-border-default);
  --workspace-toolbar-control-border-strong: color-mix(in srgb, var(--color-border-default) 80%, var(--color-text-primary));
  --workspace-toolbar-control-background: var(--workspace-toolbar-surface);
  --workspace-toolbar-control-text: var(--color-text-primary);
  --workspace-toolbar-control-muted: var(--color-text-muted);
  --workspace-toolbar-control-shadow: var(--color-shadow-soft);
  
  --workspace-toolbar-menu-surface: var(--color-bg-elevated);
  --workspace-toolbar-menu-border: var(--color-border-default);

  position: relative;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--workspace-directory-toolbar-gap-bottom, 1.5rem);
}

:global([data-theme='dark']) .workspace-directory-toolbar {
  --workspace-toolbar-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --workspace-toolbar-surface-subtle: color-mix(in srgb, var(--color-bg-elevated) 84%, var(--color-bg-base));
  --workspace-toolbar-control-border: color-mix(in srgb, var(--color-border-default) 72%, transparent);
  --workspace-toolbar-control-border-strong: color-mix(in srgb, var(--color-primary) 32%, var(--color-border-default));
  --workspace-toolbar-control-background: var(--workspace-toolbar-surface);
  --workspace-toolbar-menu-surface: var(--workspace-toolbar-surface-subtle);
  --workspace-toolbar-menu-border: color-mix(in srgb, var(--color-border-default) 76%, transparent);
}

.workspace-directory-toolbar__main,
.workspace-directory-toolbar__meta {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.workspace-directory-toolbar__search {
  position: relative;
}

.workspace-directory-toolbar__search-icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--workspace-toolbar-control-muted);
}

.workspace-directory-toolbar__search-input {
  width: 20rem;
  min-height: var(--ui-control-height-md);
  padding: 0 1rem 0 2.25rem;
  font-size: var(--font-size-13);
  font-weight: 500;
  border: 1px solid var(--workspace-toolbar-control-border);
  border-radius: var(--ui-control-radius-md);
  background: var(--workspace-toolbar-control-background);
  color: var(--workspace-toolbar-control-text);
  box-shadow: var(--workspace-toolbar-control-shadow);
  outline: none;
  transition: all 0.2s ease;
}

.workspace-directory-toolbar__search-input:hover {
  border-color: var(--workspace-toolbar-control-border-strong);
}

.workspace-directory-toolbar__search-input:focus {
  border-color: var(--color-primary);
  background: var(--workspace-toolbar-control-background);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 12%, transparent);
}

.workspace-directory-toolbar__filter-toggle,
.workspace-directory-toolbar__sort-button,
.workspace-directory-toolbar__count-pill {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-height: var(--ui-control-height-md);
  padding: 0 var(--space-4);
  border: 1px solid var(--workspace-toolbar-control-border);
  border-radius: var(--ui-control-radius-md);
  background: var(--workspace-toolbar-control-background);
  font-size: var(--font-size-12);
  font-weight: 800;
  color: var(--workspace-toolbar-control-text);
  box-shadow: var(--workspace-toolbar-control-shadow);
  transition: all 0.2s ease;
  cursor: pointer;
}

.workspace-directory-toolbar__filter-toggle--active {
  border-color: var(--color-primary);
  background: var(--color-primary-soft);
  color: var(--color-primary);
}

.workspace-directory-toolbar__sort-button:hover,
.workspace-directory-toolbar__filter-toggle:hover {
  border-color: var(--workspace-toolbar-control-border-strong);
  color: var(--color-primary);
}

.workspace-directory-toolbar__count-pill:hover {
  border-color: var(--workspace-toolbar-control-border-strong);
}

.workspace-directory-toolbar__sort {
  position: relative;
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.workspace-directory-toolbar__sort-caption {
  font-size: var(--font-size-10);
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--workspace-toolbar-control-muted);
}

.workspace-directory-toolbar__sort-label {
  font-weight: 900;
  letter-spacing: 0.03em;
}

.workspace-directory-toolbar__count-value {
  font-family: var(--font-family-mono);
  font-weight: 900;
  color: var(--color-primary);
}

.workspace-directory-toolbar__filter-panel,
.workspace-directory-toolbar__sort-menu {
  border: 1px solid var(--workspace-toolbar-menu-border);
  background: var(--workspace-toolbar-menu-surface);
  box-shadow:
    0 24px 60px color-mix(in srgb, var(--color-shadow-strong) 18%, transparent),
    0 10px 24px color-mix(in srgb, var(--color-shadow-soft) 16%, transparent);
}

.workspace-directory-toolbar__filter-panel {
  position: absolute;
  top: calc(100% + 0.5rem);
  left: 0;
  z-index: 40;
  border-radius: 1.25rem;
  padding: 1.5rem;
}

.workspace-directory-toolbar__filter-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.25rem;
}

.workspace-directory-toolbar__filter-panel-title {
  font-size: var(--font-size-14);
  font-weight: 900;
  color: var(--workspace-toolbar-control-text);
}

.workspace-directory-toolbar__filter-reset {
  font-size: var(--font-size-11);
  font-weight: 800;
  color: var(--color-primary);
  background: transparent;
  border: none;
  cursor: pointer;
}

.workspace-directory-toolbar__filter-reset:hover {
  text-decoration: underline;
}

.workspace-directory-toolbar__sort-menu {
  position: absolute;
  right: 0;
  top: calc(100% + 0.5rem);
  z-index: 40;
  width: 14rem;
  border-radius: 1rem;
  overflow: hidden;
}

.workspace-directory-toolbar__menu-title {
  padding: 0.85rem 1.25rem 0.6rem;
  font-size: var(--font-size-10);
  font-weight: 900;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-text-muted);
  background: var(--color-bg-base);
  border-bottom: 1px solid var(--color-border-default);
}

.workspace-directory-toolbar__menu-item {
  display: flex;
  width: 100%;
  align-items: center;
  padding: 0.75rem 1.25rem;
  font-size: var(--font-size-12);
  font-weight: 700;
  color: var(--color-text-secondary);
  background: transparent;
  border: none;
  transition: all 0.2s ease;
  cursor: pointer;
}

.workspace-directory-toolbar__menu-item:hover,
.workspace-directory-toolbar__menu-item--active {
  background: var(--color-primary-soft);
  color: var(--color-primary);
}

.workspace-directory-toolbar__menu-item-content {
  display: inline-flex;
  align-items: center;
  gap: var(--space-3);
}

@media (max-width: 767px) {
  .workspace-directory-toolbar {
    align-items: stretch;
  }

  .workspace-directory-toolbar__main,
  .workspace-directory-toolbar__meta {
    width: 100%;
    flex-wrap: wrap;
  }

  .workspace-directory-toolbar__search {
    flex: 1 1 100%;
  }

  .workspace-directory-toolbar__search-input {
    width: 100%;
  }

  .workspace-directory-toolbar__count-pill {
    margin-left: auto;
  }
}
</style>
