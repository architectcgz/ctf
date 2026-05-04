<script setup lang="ts">
import { computed } from 'vue'
import { Copy } from 'lucide-vue-next'

import type { AWDDefenseSSHAccessData } from '@/api/contracts'
import {
  buildOpenSSHConfig,
  getVSCodeSSHCommand,
} from '@/features/contest-awd-workspace'
import { formatTime } from '@/utils/format'

const props = defineProps<{
  access?: AWDDefenseSSHAccessData
  serviceId: string
  copiedCommand: boolean
  copiedConfig: boolean
}>()

const emit = defineEmits<{
  copyCommand: [serviceId: string]
  copyConfig: [serviceId: string]
}>()

const command = computed(() => getVSCodeSSHCommand(props.access))
const openSSHConfig = computed(() => buildOpenSSHConfig(props.access?.ssh_profile))
const expiresAtLabel = computed(() =>
  props.access?.expires_at ? `票据将在 ${formatTime(props.access.expires_at)} 过期` : ''
)
</script>

<template>
  <div v-if="access" class="asset-ssh">
    <div class="asset-ssh__topline">
      <div>
        <div class="asset-ssh__label">VS Code Remote-SSH</div>
        <code class="asset-ssh__command">{{ command }}</code>
      </div>
      <button
        v-if="command"
        class="asset-ssh__copy asset-ssh__copy--primary"
        type="button"
        @click="emit('copyCommand', serviceId)"
      >
        <Copy class="h-3 w-3" />
        <span>{{ copiedCommand ? '已复制' : '复制 VS Code 命令' }}</span>
      </button>
    </div>
    <div class="asset-ssh__secret">
      <span>密码</span>
      <code>{{ access.password }}</code>
    </div>
    <div v-if="expiresAtLabel" class="asset-ssh__expires">
      {{ expiresAtLabel }}
    </div>
    <details v-if="openSSHConfig" class="asset-ssh__details">
      <summary>OpenSSH 配置</summary>
      <pre class="asset-ssh__config">{{ openSSHConfig }}</pre>
      <button
        class="asset-ssh__copy"
        type="button"
        @click="emit('copyConfig', serviceId)"
      >
        <Copy class="h-3 w-3" />
        <span>{{ copiedConfig ? '已复制' : '复制配置' }}</span>
      </button>
    </details>
  </div>
</template>

<style scoped>
.asset-ssh {
  margin-top: var(--space-3);
  border: 1px solid color-mix(in srgb, var(--color-primary) 24%, transparent);
  border-radius: 0.625rem;
  background: color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-surface));
  padding: var(--space-2);
}

.asset-ssh__topline {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-2);
  align-items: start;
}

.asset-ssh__label {
  font-size: 9px;
  font-weight: 900;
  letter-spacing: 0.1em;
  color: var(--color-primary);
}

.asset-ssh__command {
  display: block;
  margin-top: var(--space-1);
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-size: 11px;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.asset-ssh__secret {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  margin-top: var(--space-2);
  padding: var(--space-2);
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--color-bg-surface) 70%, transparent);
  color: var(--color-text-secondary);
  font-size: 10px;
  font-weight: 800;
}

.asset-ssh__secret code {
  min-width: 0;
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  overflow-wrap: anywhere;
}

.asset-ssh__expires {
  margin-top: var(--space-2);
  color: var(--color-text-muted);
  font-size: 10px;
  font-weight: 800;
}

.asset-ssh__details {
  margin-top: var(--space-2);
}

.asset-ssh__details summary {
  cursor: pointer;
  color: var(--color-text-secondary);
  font-size: 10px;
  font-weight: 900;
}

.asset-ssh__config {
  margin-top: var(--space-2);
  padding: var(--space-2);
  white-space: pre-wrap;
  border: 1px solid color-mix(in srgb, var(--color-primary) 18%, transparent);
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--color-bg-surface) 76%, var(--color-bg-base));
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-size: 10px;
  line-height: 1.45;
}

.asset-ssh__copy {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  border: 1px solid color-mix(in srgb, var(--color-primary) 20%, transparent);
  border-radius: 0.5rem;
  background: var(--color-bg-surface);
  color: var(--color-primary);
  padding: var(--space-1) var(--space-2);
  font-size: 10px;
  font-weight: 900;
  cursor: pointer;
}

.asset-ssh__details .asset-ssh__copy {
  margin-top: var(--space-2);
}

.asset-ssh__copy--primary {
  min-height: 2rem;
  white-space: nowrap;
  background: var(--color-primary-soft);
}
</style>
