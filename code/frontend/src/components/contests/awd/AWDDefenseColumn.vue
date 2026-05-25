<script setup lang="ts">
import { ShieldAlert } from 'lucide-vue-next'

import type { AWDDefenseSSHAccessData, ID } from '@/api/contracts'
import type { AWDDefenseServiceCard } from '@/features/contest-awd-workspace'
import AWDDefenseAlertsPanel from './AWDDefenseAlertsPanel.vue'
import AWDDefenseOperationsPanel from './AWDDefenseOperationsPanel.vue'
import AWDDefenseServiceList from './AWDDefenseServiceList.vue'

defineProps<{
  alerts: Array<{
    challengeId: string
    challengeTitle: string
    statusLabel: string
    tone: 'danger' | 'warning'
    issues: string[]
  }>
  services: AWDDefenseServiceCard[]
  selectedServiceId: string
  openingServiceKey: string
  openingSshKey: string
  serviceActionPendingById: Record<string, boolean>
  serviceCard: AWDDefenseServiceCard | null
  serviceTitle: string
  actionPending: boolean
  loading: boolean
  access?: AWDDefenseSSHAccessData
  copiedCommand: boolean
  copiedPassword: boolean
}>()

const emit = defineEmits<{
  selectService: [serviceId: string]
  openService: [serviceId: ID]
  requestSsh: [serviceId: ID]
  restartService: [serviceId: ID]
  refresh: []
  copyCommand: [serviceId: string]
  copyPassword: [serviceId: string]
}>()
</script>

<template>
  <aside class="war-room-col column-defense">
    <section class="defense-panel">
      <header class="defense-panel__header">
        <ShieldAlert class="defense-panel__icon h-4 w-4" />
        <h3 class="defense-panel__title">我的防守</h3>
      </header>

      <div class="defense-panel__content">
        <AWDDefenseAlertsPanel :alerts="alerts" />

        <AWDDefenseServiceList
          :services="services"
          :selected-service-id="selectedServiceId"
          :opening-service-key="openingServiceKey"
          :opening-ssh-key="openingSshKey"
          :service-action-pending-by-id="serviceActionPendingById"
          @select-service="emit('selectService', $event)"
          @open-service="emit('openService', $event)"
          @request-ssh="emit('requestSsh', $event)"
          @restart-service="emit('restartService', $event)"
        />
        <AWDDefenseOperationsPanel
          :service-card="serviceCard"
          :service-title="serviceTitle"
          :opening-service-key="openingServiceKey"
          :opening-ssh-key="openingSshKey"
          :action-pending="actionPending"
          :loading="loading"
          :access="access"
          :copied-command="copiedCommand"
          :copied-password="copiedPassword"
          @open-service="emit('openService', $event)"
          @request-ssh="emit('requestSsh', $event)"
          @restart-service="emit('restartService', $event)"
          @refresh="emit('refresh')"
          @copy-command="emit('copyCommand', $event)"
          @copy-password="emit('copyPassword', $event)"
        />
      </div>
    </section>
  </aside>
</template>

<style scoped>
.column-defense {
  grid-area: defense;
}

.defense-panel {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  border-radius: 1rem;
  box-shadow: var(--color-shadow-soft);
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.defense-panel__header {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-border-subtle);
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: var(--color-bg-elevated);
}

.defense-panel__icon {
  color: var(--color-warning);
}

.defense-panel__title {
  font-size: var(--font-size-12);
  font-weight: 900;
  letter-spacing: 0.15em;
  color: var(--color-text-primary);
  margin: 0;
}

.defense-panel__content {
  flex: 1;
  overflow-y: auto;
  padding: 1.25rem;
}

.defense-panel__content::-webkit-scrollbar {
  width: 4px;
}

.defense-panel__content::-webkit-scrollbar-thumb {
  background: var(--color-border-default);
  border-radius: 10px;
}
</style>
