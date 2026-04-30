<script setup lang="ts">
import type { AdminContestAWDServiceData, AWDCheckerType } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

defineProps<{
  loading: boolean
  services: AdminContestAWDServiceData[]
  selectedServiceId: string
  getCheckerTypeLabel: (value?: AWDCheckerType) => string
  getValidationLabel: (value?: AdminContestAWDServiceData['validation_state']) => string
}>()

const emit = defineEmits<{
  select: [service: AdminContestAWDServiceData]
}>()
</script>

<template>
  <aside class="awd-config-page__services" aria-label="AWD 服务目录">
    <header class="list-heading awd-config-section-head">
      <div class="awd-config-section-head__main">
        <div class="journal-note-label">Service Directory</div>
        <h2 class="list-heading__title">服务目录</h2>
      </div>
      <span class="awd-config-section-count">{{ services.length }}</span>
    </header>

    <AppEmpty
      v-if="!loading && services.length === 0"
      title="暂无 AWD 服务"
      description="请先回到题目编排关联 AWD 题目。"
      icon="ShieldCheck"
      class="awd-config-page__empty"
    />

    <div v-else class="awd-service-list awd-config-page__service-scroll" role="list">
      <button
        v-for="service in services"
        :key="service.id"
        type="button"
        class="awd-service-row"
        :class="{ 'is-active': selectedServiceId === service.id }"
        role="listitem"
        @click="emit('select', service)"
      >
        <span class="awd-service-row__index">#{{ service.order }}</span>
        <span class="awd-service-row__main">
          <strong :title="service.display_name">{{ service.display_name }}</strong>
          <small>{{ service.category || '通用' }}</small>
        </span>
        <span class="awd-service-row__meta">
          <span class="awd-service-row__checker">{{ getCheckerTypeLabel(service.checker_type) }}</span>
          <span class="validation-pill" :class="service.validation_state || 'pending'">
            {{ getValidationLabel(service.validation_state) }}
          </span>
        </span>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.awd-config-page__services {
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: var(--space-5);
  border-right: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
}

.awd-config-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}

.awd-config-section-head__main {
  min-width: 0;
}

.awd-config-section-count {
  flex: none;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  border-radius: var(--ui-badge-radius-soft);
  padding: var(--space-1) var(--space-2);
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 700;
}

.awd-service-list {
  display: grid;
  gap: var(--space-1);
}

.awd-config-page__service-scroll {
  min-height: 0;
  overflow: auto;
  padding-right: var(--space-1);
}

.awd-service-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-3);
  width: 100%;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 52%, transparent);
  border-radius: var(--ui-control-radius);
  background: color-mix(in srgb, var(--color-bg-surface) 72%, transparent);
  padding: var(--space-2) var(--space-3);
  text-align: left;
  color: var(--color-text-primary);
  cursor: pointer;
  transition:
    background var(--ui-motion-fast),
    border-color var(--ui-motion-fast),
    transform var(--ui-motion-fast);
}

.awd-service-row__index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 2.75rem;
  border-radius: var(--ui-badge-radius-soft);
  padding: var(--space-1) var(--space-2);
  background: color-mix(in srgb, var(--color-bg-elevated) 88%, transparent);
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 700;
}

.awd-service-row:hover,
.awd-service-row:focus-visible,
.awd-service-row.is-active {
  border-color: color-mix(in srgb, var(--color-primary) 34%, transparent);
  background: color-mix(in srgb, var(--color-primary-soft) 46%, var(--color-bg-surface));
  transform: translateY(-1px);
  outline: none;
}

.awd-service-row__main,
.awd-service-row__meta {
  min-width: 0;
  display: grid;
  gap: 0;
}

.awd-service-row__main strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-14);
  font-weight: 700;
}

.awd-service-row__main small,
.awd-service-row__checker {
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
}

.awd-service-row__meta {
  justify-items: end;
  gap: var(--space-1);
}

.validation-pill {
  width: fit-content;
  border-radius: var(--ui-badge-radius-soft);
  padding: var(--space-1) var(--space-2);
  background: var(--color-bg-elevated);
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 800;
}

.validation-pill.passed {
  background: var(--color-success-soft);
  color: var(--color-success);
}

.validation-pill.failed,
.validation-pill.stale {
  background: var(--color-warning-soft);
  color: var(--color-warning);
}

.awd-config-page__empty {
  margin: var(--space-8);
}

@media (max-width: 1023px) {
  .awd-config-page__services {
    border-right: 0;
    border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  }
}

@media (max-width: 767px) {
  .awd-service-row {
    grid-template-columns: minmax(0, 1fr);
    align-items: start;
  }

  .awd-service-row__meta {
    justify-items: start;
  }
}
</style>
