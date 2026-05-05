<script setup lang="ts">
import { ExternalLink, ShieldCheck } from 'lucide-vue-next'

import type { AWDDefenseServiceCard } from '@/features/contest-awd-workspace'

const props = defineProps<{
  services: AWDDefenseServiceCard[]
  selectedServiceId: string
  openingServiceKey: string
  openingSshKey: string
  serviceActionPendingById: Record<string, boolean>
}>()

const emit = defineEmits<{
  selectService: [serviceId: string]
  openDefense: [serviceId: string]
  openService: [serviceId: string]
  requestSsh: [serviceId: string]
  restartService: [serviceId: string]
}>()

function formatServiceRef(serviceId?: string): string {
  return `服务 #${serviceId || '--'}`
}

function getServiceStatusClass(card: AWDDefenseServiceCard): string {
  if (card.riskLevel === 'critical') return 'status-badge status-badge--compromised'
  if (card.serviceStatusLabel === '离线') return 'status-badge status-badge--down'
  if (card.serviceStatusLabel === '正常') return 'status-badge status-badge--up'
  return 'status-badge status-badge--pending'
}

function isActionPending(card: AWDDefenseServiceCard): boolean {
  return Boolean(props.serviceActionPendingById[card.serviceId])
}
</script>

<template>
  <div class="asset-list">
    <div class="asset-header">防守服务</div>
    <div v-if="services.length === 0" class="panel-note">当前竞赛暂无可部署服务。</div>
    <div
      v-for="service in services"
      :key="service.serviceId"
      class="asset-item"
      :class="{ 'is-selected': service.serviceId === selectedServiceId }"
      @click="emit('selectService', service.serviceId)"
    >
      <div class="asset-main">
        <div class="asset-topline">
          <div class="asset-title-stack">
            <span class="asset-title">{{ service.title }}</span>
            <span class="asset-ref">{{ formatServiceRef(service.serviceId) }}</span>
          </div>
          <span :class="getServiceStatusClass(service)">
            {{ service.serviceStatusLabel }}
          </span>
        </div>
        <div class="asset-meta">
          {{ service.instanceStatusLabel }}
        </div>
        <div v-if="service.riskReasons.length > 0" class="asset-risk">
          <span v-for="reason in service.riskReasons" :key="reason">{{ reason }}</span>
        </div>
      </div>
      <div class="asset-actions" role="group" :aria-label="`${service.title} 防守操作`">
        <button
          class="asset-btn"
          type="button"
          @click.stop="emit('openDefense', service.serviceId)"
        >
          <ShieldCheck class="h-3 w-3" />
          <span>防守</span>
        </button>
        <button
          v-if="service.instanceId"
          :disabled="!service.canOpenService || openingServiceKey === service.instanceId"
          class="asset-btn"
          type="button"
          @click.stop="emit('openService', service.serviceId)"
        >
          <ExternalLink class="h-3 w-3" />
        </button>
        <button
          v-if="service.instanceId"
          :disabled="!service.canRequestSSH || openingSshKey === service.serviceId"
          class="asset-btn"
          type="button"
          @click.stop="emit('requestSsh', service.serviceId)"
        >
          SSH
        </button>
        <button
          :disabled="!service.canRestart || isActionPending(service)"
          class="asset-btn asset-btn--primary"
          type="button"
          @click.stop="emit('restartService', service.serviceId)"
        >
          {{ isActionPending(service) ? '重启中' : '重启' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.asset-header {
  font-size: var(--font-size-11);
  font-weight: 900;
  color: var(--color-text-muted);
  letter-spacing: 0.1em;
  margin-bottom: var(--space-3);
  text-transform: uppercase;
}

.asset-item {
  background: color-mix(in srgb, var(--color-bg-elevated) 72%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-border-default) 84%, transparent);
  padding: var(--space-3);
  border-radius: var(--ui-control-radius-sm);
  margin-bottom: var(--space-2);
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
  cursor: pointer;
  transition:
    border-color var(--ui-motion-fast),
    background var(--ui-motion-fast);
}

.asset-item.is-selected {
  border-color: color-mix(in srgb, var(--color-primary) 46%, transparent);
  background: color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-elevated));
}

.asset-item:hover {
  border-color: color-mix(in srgb, var(--color-primary) 30%, transparent);
}

.asset-main {
  min-width: 0;
  flex: 1;
}

.asset-topline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.asset-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-shrink: 0;
}

.asset-title-stack {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  min-width: 0;
}

.asset-title {
  font-size: var(--font-size-13);
  font-weight: 800;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.asset-ref {
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
}

.asset-meta {
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-11);
  margin-top: var(--space-2);
}

.asset-risk {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1);
  margin-top: var(--space-2);
}

.asset-risk span {
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-warning) 12%, transparent);
  color: var(--color-warning);
  padding: var(--space-0-5) var(--space-2);
  font-size: var(--font-size-11);
  font-weight: 800;
}

.status-badge {
  font-size: var(--font-size-11);
  font-weight: 900;
  padding: var(--space-0-5) var(--space-2);
  border-radius: 999px;
  white-space: nowrap;
}

.status-badge--up {
  background: var(--color-success-soft);
  color: var(--color-success);
}

.status-badge--down {
  background: var(--color-danger-soft);
  color: var(--color-danger);
}

.status-badge--compromised {
  background: var(--color-warning-soft);
  color: var(--color-warning);
}

.status-badge--pending {
  background: color-mix(in srgb, var(--color-text-secondary) 12%, transparent);
  color: var(--color-text-secondary);
}

.asset-btn {
  min-width: var(--ui-control-height-sm);
  height: var(--ui-control-height-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
  padding: 0 var(--space-2);
  border-radius: var(--ui-control-radius-sm);
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border-default);
  cursor: pointer;
  font-size: var(--font-size-11);
  font-weight: 800;
  transition: all 0.2s ease;
}

.asset-btn:hover {
  color: var(--color-primary);
  border-color: var(--color-primary);
}

.asset-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.asset-btn--primary {
  width: auto;
  padding: 0 var(--space-3);
  font-size: var(--font-size-11);
  font-weight: 900;
  background: var(--color-primary-soft);
  color: var(--color-primary);
  border-color: color-mix(in srgb, var(--color-primary) 20%, transparent);
}

.asset-btn--primary:hover {
  background: var(--color-primary);
  color: var(--color-bg-base);
}

@media (max-width: 42rem) {
  .asset-item {
    align-items: stretch;
    flex-direction: column;
  }

  .asset-actions {
    justify-content: flex-start;
    flex-wrap: wrap;
  }
}
</style>
