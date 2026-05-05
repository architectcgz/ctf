<script setup lang="ts">
import { computed } from 'vue'
import { Copy } from 'lucide-vue-next'

import type { AWDDefenseSSHAccessData } from '@/api/contracts'
import { formatTime } from '@/utils/format'

const props = defineProps<{
  access?: AWDDefenseSSHAccessData
  serviceId: string
  copiedCommand: boolean
}>()

const emit = defineEmits<{
  copyCommand: [serviceId: string]
}>()

const command = computed(() => props.access?.command || '')
const expiresAtLabel = computed(() =>
  props.access?.expires_at ? `票据将在 ${formatTime(props.access.expires_at)} 过期` : ''
)
</script>

<template>
  <div v-if="access" class="asset-ssh">
    <div class="asset-ssh__topline">
      <div>
        <div class="asset-ssh__label">SSH 连接</div>
        <code class="asset-ssh__command">{{ command }}</code>
      </div>
      <button
        v-if="command"
        class="asset-ssh__copy asset-ssh__copy--primary"
        type="button"
        @click="emit('copyCommand', serviceId)"
      >
        <Copy class="h-3 w-3" />
        <span>{{ copiedCommand ? '已复制' : '复制 SSH 命令' }}</span>
      </button>
    </div>
    <dl class="asset-ssh__meta">
      <div>
        <dt>主机</dt>
        <dd><code>{{ access.host }}</code></dd>
      </div>
      <div>
        <dt>端口</dt>
        <dd><code>{{ access.port }}</code></dd>
      </div>
      <div>
        <dt>用户</dt>
        <dd><code>{{ access.username }}</code></dd>
      </div>
    </dl>
    <div class="asset-ssh__secret">
      <span>密码</span>
      <code>{{ access.password }}</code>
    </div>
    <div v-if="expiresAtLabel" class="asset-ssh__expires">
      {{ expiresAtLabel }}
    </div>
  </div>
</template>

<style scoped>
.asset-ssh {
  margin-top: var(--space-3);
  border: 1px solid color-mix(in srgb, var(--color-primary) 24%, transparent);
  border-radius: var(--ui-control-radius-md);
  background: color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-surface));
  max-height: min(60vh, calc(var(--space-12) * 8));
  overflow: auto;
  padding: var(--space-2);
}

.asset-ssh__topline {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-2);
  align-items: start;
}

.asset-ssh__label {
  font-size: var(--font-size-11);
  font-weight: 900;
  letter-spacing: 0.1em;
  color: var(--color-primary);
  text-transform: uppercase;
}

.asset-ssh__command {
  display: block;
  margin-top: var(--space-1);
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-11);
  line-height: 1.45;
  max-width: 100%;
  overflow-x: auto;
  white-space: nowrap;
}

.asset-ssh__secret {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  margin-top: var(--space-2);
  padding: var(--space-2);
  border-radius: var(--ui-control-radius-sm);
  background: color-mix(in srgb, var(--color-bg-surface) 70%, transparent);
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 800;
}

.asset-ssh__secret code {
  min-width: 0;
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.asset-ssh__meta {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-2);
  margin-top: var(--space-2);
}

.asset-ssh__meta div {
  min-width: 0;
  padding: var(--space-2);
  border-radius: var(--ui-control-radius-sm);
  background: color-mix(in srgb, var(--color-bg-surface) 70%, transparent);
}

.asset-ssh__meta dt {
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 800;
}

.asset-ssh__meta dd {
  margin: var(--space-1) 0 0;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-11);
}

.asset-ssh__expires {
  margin-top: var(--space-2);
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 800;
}

.asset-ssh__copy {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  border: 1px solid color-mix(in srgb, var(--color-primary) 20%, transparent);
  border-radius: var(--ui-control-radius-sm);
  background: var(--color-bg-surface);
  color: var(--color-primary);
  padding: var(--space-1) var(--space-2);
  font-size: var(--font-size-11);
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

@media (max-width: 42rem) {
  .asset-ssh__meta {
    grid-template-columns: 1fr;
  }
}
</style>
