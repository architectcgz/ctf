<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    page: number
    totalPages: number
    total: number
    totalLabel?: string
    disabled?: boolean
    prevButtonId?: string
    nextButtonId?: string
    showJump?: boolean
  }>(),
  {
    totalLabel: '',
    disabled: false,
    prevButtonId: '',
    nextButtonId: '',
    showJump: false,
  }
)

const emit = defineEmits<{
  changePage: [page: number]
}>()

const safePage = computed(() => Math.max(1, Math.floor(props.page || 1)))
const safeTotalPages = computed(() => Math.max(1, Math.floor(props.totalPages || 1)))
const jumpPage = ref(String(safePage.value))
const summaryText = computed(() => props.totalLabel.trim() || `共 ${props.total} 条`)

watch(
  () => safePage.value,
  (page) => {
    jumpPage.value = String(page)
  },
  { immediate: true }
)

function emitPageChange(nextPage: number): boolean {
  if (props.disabled) {
    return false
  }

  const normalizedPage = Math.floor(nextPage)
  if (
    !Number.isFinite(normalizedPage) ||
    normalizedPage < 1 ||
    normalizedPage > safeTotalPages.value ||
    normalizedPage === safePage.value
  ) {
    return false
  }

  emit('changePage', normalizedPage)
  jumpPage.value = String(normalizedPage)
  return true
}

function submitJumpPage(): void {
  const nextPage = Number.parseInt(jumpPage.value, 10)
  const changed = emitPageChange(nextPage)

  if (!changed) {
    jumpPage.value = String(safePage.value)
  }
}
</script>

<template>
  <div class="page-pagination-controls" :class="{ 'is-disabled': disabled }">
    <span class="page-pagination-controls__summary">{{ summaryText }}</span>

    <div class="page-pagination-controls__actions">
      <button
        :id="prevButtonId || undefined"
        type="button"
        class="page-pagination-controls__button"
        :disabled="disabled || safePage <= 1"
        @click="emitPageChange(safePage - 1)"
      >
        上一页
      </button>

      <span class="page-pagination-controls__status">{{ safePage }} / {{ safeTotalPages }}</span>

      <button
        :id="nextButtonId || undefined"
        type="button"
        class="page-pagination-controls__button"
        :disabled="disabled || safePage >= safeTotalPages"
        @click="emitPageChange(safePage + 1)"
      >
        下一页
      </button>

      <form v-if="showJump" class="page-pagination-controls__jump" @submit.prevent="submitJumpPage">
        <label class="page-pagination-controls__jump-label">
          <span>跳至</span>
          <input
            v-model="jumpPage"
            class="page-pagination-controls__input"
            type="number"
            inputmode="numeric"
            min="1"
            :max="safeTotalPages"
            :disabled="disabled"
            aria-label="跳转到页码"
          />
        </label>
        <span class="page-pagination-controls__jump-suffix">页</span>
        <button
          type="submit"
          class="page-pagination-controls__button page-pagination-controls__button--submit"
          :disabled="disabled || safeTotalPages <= 1"
        >
          跳转
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.page-pagination-controls {
  --page-pagination-ink: var(--journal-muted, var(--color-text-secondary));
  --page-pagination-border: var(
    --pagination-control-border,
    var(
      --admin-control-border,
      color-mix(in srgb, var(--journal-border, var(--color-border-default)) 80%, transparent)
    )
  );
  --page-pagination-surface: color-mix(
    in srgb,
    var(--journal-surface, var(--color-bg-surface)) 88%,
    transparent
  );
  width: 100%;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem 1rem;
  color: var(--page-pagination-ink);
}

.page-pagination-controls__summary {
  color: var(--page-pagination-ink);
}

.page-pagination-controls__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 0.65rem;
}

.page-pagination-controls__status {
  min-width: 4.75rem;
  text-align: center;
  font-variant-numeric: tabular-nums;
}

.page-pagination-controls__jump {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}

.page-pagination-controls__jump-label {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--page-pagination-ink);
}

.page-pagination-controls__input,
.page-pagination-controls__button {
  min-height: 34px;
  border-radius: 10px;
  border: 1px solid var(--page-pagination-border);
  transition:
    border-color 0.2s ease,
    transform 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease,
    box-shadow 0.2s ease;
}

.page-pagination-controls__input {
  width: 4.5rem;
  padding: 0 0.75rem;
  background: var(--page-pagination-surface);
  color: var(--journal-ink, var(--color-text-primary));
}

.page-pagination-controls__button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 0.95rem;
  background: transparent;
  color: var(--journal-ink, var(--color-text-primary));
}

.page-pagination-controls__button:hover:not(:disabled) {
  border-color: color-mix(
    in srgb,
    var(--journal-accent, var(--color-primary)) 42%,
    var(--page-pagination-border)
  );
  color: var(--journal-accent, var(--color-primary));
  transform: translateY(-1px);
}

.page-pagination-controls__button:focus-visible,
.page-pagination-controls__input:focus-visible {
  outline: none;
  border-color: color-mix(
    in srgb,
    var(--journal-accent, var(--color-primary)) 54%,
    var(--page-pagination-border)
  );
  box-shadow: 0 0 0 3px
    color-mix(in srgb, var(--journal-accent, var(--color-primary)) 14%, transparent);
}

.page-pagination-controls__button:disabled,
.page-pagination-controls__input:disabled,
.page-pagination-controls.is-disabled {
  cursor: not-allowed;
}

.page-pagination-controls__button:disabled,
.page-pagination-controls__input:disabled {
  opacity: 0.45;
}

.page-pagination-controls__input::-webkit-outer-spin-button,
.page-pagination-controls__input::-webkit-inner-spin-button {
  margin: 0;
  -webkit-appearance: none;
}

.page-pagination-controls__input[type='number'] {
  -moz-appearance: textfield;
}

@media (max-width: 720px) {
  .page-pagination-controls {
    align-items: flex-start;
    flex-direction: column;
  }

  .page-pagination-controls__actions {
    justify-content: flex-start;
  }
}
</style>
