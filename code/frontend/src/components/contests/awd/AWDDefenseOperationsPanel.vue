<script setup lang="ts">
import { computed } from 'vue'
import {
  ExternalLink,
  RefreshCw,
  RotateCcw,
  ShieldCheck,
  Terminal,
} from 'lucide-vue-next'

import type {
  AWDDefenseSSHAccessData,
  ContestAWDWorkspaceServiceData,
  ID,
} from '@/api/contracts'
import type { AWDDefenseServiceCard } from '@/features/contest-awd-workspace'
import AWDDefenseConnectionPanel from './AWDDefenseConnectionPanel.vue'

const props = defineProps<{
  serviceCard: AWDDefenseServiceCard | null
  service: ContestAWDWorkspaceServiceData | null
  serviceTitle: string
  openingServiceKey: string
  openingSshKey: string
  actionPending: boolean
  loading: boolean
  access?: AWDDefenseSSHAccessData
  copiedCommand: boolean
  copiedConfig: boolean
}>()

const emit = defineEmits<{
  openService: [serviceId: ID]
  requestSsh: [serviceId: ID]
  restartService: [serviceId: ID]
  refresh: []
  copyCommand: [serviceId: string]
  copyConfig: [serviceId: string]
}>()

const serviceId = computed(() => props.serviceCard?.serviceId || '')
const receivedAttackCount = computed(() => props.service?.attack_received ?? 0)
const hasReceivedAttacks = computed(() => receivedAttackCount.value > 0)
const canOpenService = computed(() =>
  Boolean(
    props.serviceCard?.canOpenService &&
      props.serviceCard.instanceId &&
      props.openingServiceKey !== props.serviceCard.instanceId
  )
)
const canRequestSSH = computed(() =>
  Boolean(props.serviceCard?.canRequestSSH && props.openingSshKey !== serviceId.value)
)
const canRestart = computed(() =>
  Boolean(props.serviceCard?.canRestart && !props.actionPending)
)

function emitOpenService(): void {
  if (!serviceId.value) return
  emit('openService', serviceId.value)
}

function emitRequestSsh(): void {
  if (!serviceId.value) return
  emit('requestSsh', serviceId.value)
}

function emitRestartService(): void {
  if (!serviceId.value) return
  emit('restartService', serviceId.value)
}
</script>

<template>
  <section class="defense-ops" aria-label="Web 防守工作台">
    <header class="defense-ops__header">
      <div>
        <div class="defense-ops__eyebrow">Web 防守</div>
        <h4 class="defense-ops__title">{{ serviceTitle || '未选择服务' }}</h4>
      </div>
      <button
        class="defense-ops__refresh"
        type="button"
        :disabled="loading"
        @click="emit('refresh')"
      >
        <RefreshCw class="h-3.5 w-3.5" :class="{ 'animate-spin': loading }" />
        <span>刷新</span>
      </button>
    </header>

    <div v-if="!serviceCard" class="defense-ops__empty">请选择一个防守服务。</div>
    <div v-else class="defense-ops__body">
      <section class="defense-ops__block">
        <div class="defense-ops__block-title">
          <ShieldCheck class="h-3.5 w-3.5" />
          <span>风险</span>
        </div>
        <div class="defense-ops__status-grid">
          <div>
            <span>服务</span>
            <strong>{{ serviceCard.serviceStatusLabel }}</strong>
          </div>
          <div>
            <span>实例</span>
            <strong>{{ serviceCard.instanceStatusLabel }}</strong>
          </div>
        </div>
        <div class="defense-ops__chips">
          <span
            v-for="reason in serviceCard.riskReasons"
            :key="reason"
            class="defense-ops__chip"
          >
            {{ reason }}
          </span>
          <span v-if="!serviceCard.riskReasons.length" class="defense-ops__chip">暂无告警</span>
        </div>
      </section>

      <section class="defense-ops__block">
        <div class="defense-ops__block-title">
          <Terminal class="h-3.5 w-3.5" />
          <span>风险片段</span>
        </div>
        <div class="defense-ops__fragment">
          <div>
            <span class="defense-ops__fragment-title">{{
              hasReceivedAttacks ? '最近受到攻击' : '等待片段'
            }}</span>
            <span class="defense-ops__fragment-meta">{{
              hasReceivedAttacks ? `${receivedAttackCount} 次` : '当前服务暂无可展示片段'
            }}</span>
          </div>
          <span class="defense-ops__fragment-badge">只读</span>
        </div>
      </section>

      <section class="defense-ops__block">
        <div class="defense-ops__block-title">
          <RotateCcw class="h-3.5 w-3.5" />
          <span>验证</span>
        </div>
        <div class="defense-ops__actions" role="group" :aria-label="`${serviceCard.title} 验证操作`">
          <button
            class="defense-ops__button"
            type="button"
            :disabled="!canOpenService"
            @click="emitOpenService"
          >
            <ExternalLink class="h-3.5 w-3.5" />
            <span>{{ openingServiceKey === serviceCard.instanceId ? '打开中' : '打开服务' }}</span>
          </button>
          <button
            class="defense-ops__button"
            type="button"
            :disabled="!canRestart"
            @click="emitRestartService"
          >
            <RotateCcw class="h-3.5 w-3.5" />
            <span>{{ actionPending ? '重启中' : '重启服务' }}</span>
          </button>
          <button
            class="defense-ops__button defense-ops__button--primary"
            type="button"
            :disabled="!canRequestSSH"
            @click="emitRequestSsh"
          >
            <Terminal class="h-3.5 w-3.5" />
            <span>{{ openingSshKey === serviceId ? '生成中' : 'SSH' }}</span>
          </button>
        </div>
      </section>

      <AWDDefenseConnectionPanel
        :access="access"
        :service-id="serviceId"
        :copied-command="copiedCommand"
        :copied-config="copiedConfig"
        @copy-command="emit('copyCommand', $event)"
        @copy-config="emit('copyConfig', $event)"
      />
    </div>
  </section>
</template>

<style scoped>
.defense-ops {
  margin-top: var(--space-4);
  border-top: 1px solid var(--color-border-default);
  padding-top: var(--space-4);
}

.defense-ops__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.defense-ops__eyebrow {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 900;
}

.defense-ops__title {
  color: var(--color-text-primary);
  font-size: var(--font-size-14);
  font-weight: 800;
  margin: var(--space-1) 0 0;
}

.defense-ops__refresh,
.defense-ops__button {
  display: inline-flex;
  align-items: center;
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-md);
  font-weight: 800;
}

.defense-ops__refresh {
  background: var(--color-bg-elevated);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  gap: var(--space-1);
  padding: var(--space-2) var(--space-3);
}

.defense-ops__refresh:disabled,
.defense-ops__button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.defense-ops__empty {
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  margin-top: var(--space-3);
}

.defense-ops__body {
  display: grid;
  gap: var(--space-3);
  margin-top: var(--space-3);
}

.defense-ops__block {
  background: color-mix(in srgb, var(--color-bg-elevated) 76%, transparent);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-lg);
  padding: var(--space-3);
}

.defense-ops__block-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--color-text-primary);
  font-size: var(--font-size-12);
  font-weight: 900;
}

.defense-ops__status-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-2);
  margin-top: var(--space-3);
}

.defense-ops__status-grid div {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.defense-ops__status-grid span,
.defense-ops__fragment-meta {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 800;
}

.defense-ops__status-grid strong {
  color: var(--color-text-primary);
  font-size: var(--font-size-12);
}

.defense-ops__chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1);
  margin-top: var(--space-3);
}

.defense-ops__chip,
.defense-ops__fragment-badge {
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-warning) 12%, transparent);
  color: var(--color-warning);
  font-size: var(--font-size-11);
  font-weight: 900;
  padding: var(--space-1) var(--space-2);
}

.defense-ops__fragment {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-top: var(--space-3);
}

.defense-ops__fragment-title {
  display: block;
  color: var(--color-text-primary);
  font-size: var(--font-size-12);
  font-weight: 900;
}

.defense-ops__actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-2);
  margin-top: var(--space-3);
}

.defense-ops__button {
  justify-content: center;
  gap: var(--space-1);
  min-height: var(--control-height-sm);
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  padding: 0 var(--space-2);
}

.defense-ops__button--primary {
  background: var(--color-primary-soft);
  border-color: color-mix(in srgb, var(--color-primary) 22%, transparent);
  color: var(--color-primary);
}

@media (max-width: 42rem) {
  .defense-ops__actions,
  .defense-ops__status-grid {
    grid-template-columns: 1fr;
  }
}
</style>
