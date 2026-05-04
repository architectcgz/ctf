<script setup lang="ts">
import { ExternalLink } from 'lucide-vue-next'

import type { AWDDefenseSSHAccessData } from '@/api/contracts'
import type { AWDDefenseServiceCard } from '@/features/contest-awd-workspace'
import AWDDefenseConnectionPanel from './AWDDefenseConnectionPanel.vue'

const props = defineProps<{
  services: AWDDefenseServiceCard[]
  selectedServiceId: string
  openingServiceKey: string
  openingSshKey: string
  serviceActionPendingById: Record<string, boolean>
  accessByServiceId: Record<string, AWDDefenseSSHAccessData>
  copiedCommandKey: string
  copiedConfigKey: string
}>()

const emit = defineEmits<{
  selectService: [serviceId: string]
  openService: [serviceId: string]
  requestSsh: [serviceId: string]
  restartService: [serviceId: string]
  copyCommand: [serviceId: string]
  copyConfig: [serviceId: string]
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
  <div class="asset-list mt-4">
    <div class="asset-header">战队服务</div>
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
        <div class="asset-meta font-mono text-[10px]">
          {{ service.instanceStatusLabel }}
        </div>
        <div v-if="service.riskReasons.length > 0" class="asset-risk">
          <span v-for="reason in service.riskReasons" :key="reason">{{ reason }}</span>
        </div>
        <AWDDefenseConnectionPanel
          :access="accessByServiceId[service.serviceId]"
          :service-id="service.serviceId"
          :copied-command="copiedCommandKey === service.serviceId"
          :copied-config="copiedConfigKey === service.serviceId"
          @copy-command="emit('copyCommand', $event)"
          @copy-config="emit('copyConfig', $event)"
        />
      </div>
      <div class="asset-actions" role="group" :aria-label="`${service.title} 防守操作`">
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
          {{ openingSshKey === service.serviceId ? '...' : 'SSH' }}
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
  font-size: 10px;
  font-weight: 900;
  color: var(--color-text-muted);
  letter-spacing: 0.1em;
  margin-bottom: 0.75rem;
}

.asset-item {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-default);
  padding: 1rem;
  border-radius: 0.75rem;
  margin-bottom: 0.75rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
  cursor: pointer;
}

.asset-item.is-selected {
  border-color: color-mix(in srgb, var(--color-primary) 46%, transparent);
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
  gap: 0.2rem;
}

.asset-title {
  font-size: 13px;
  font-weight: 800;
  color: var(--color-text-primary);
}

.asset-ref {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
}

.asset-meta {
  color: var(--color-text-muted);
  margin-top: 0.35rem;
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
  padding: 0.15rem 0.45rem;
  font-size: 10px;
  font-weight: 800;
}

.status-badge {
  font-size: 10px;
  font-weight: 900;
  padding: 0.2rem 0.6rem;
  border-radius: 99px;
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
  width: 2.25rem;
  height: 2.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border-default);
  cursor: pointer;
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
  padding: 0 1rem;
  font-size: 11px;
  font-weight: 900;
  background: var(--color-primary-soft);
  color: var(--color-primary);
  border-color: color-mix(in srgb, var(--color-primary) 20%, transparent);
}

.asset-btn--primary:hover {
  background: var(--color-primary);
  color: var(--color-bg-base);
}
</style>
