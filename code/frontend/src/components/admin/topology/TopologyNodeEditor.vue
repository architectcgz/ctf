<script setup lang="ts">
import { Plus, Server, Trash2 } from 'lucide-vue-next'

import type { AdminImageListItem } from '@/api/contracts'

import {
  createEmptyEnvEntryDraft,
  type TopologyNetworkDraft,
  type TopologyNodeDraft,
} from './topologyDraft'

const node = defineModel<TopologyNodeDraft>({ required: true })

defineProps<{
  index: number
  images: AdminImageListItem[]
  networks: TopologyNetworkDraft[]
  removable: boolean
  selected?: boolean
}>()

const emit = defineEmits<{
  remove: []
}>()

function updateNumberField(
  field: 'service_port' | 'cpu_quota' | 'memory_mb' | 'pids_limit',
  value: string
) {
  node.value[field] = value.trim() === '' ? null : Number(value)
}

function toggleNetwork(key: string, checked: boolean) {
  const next = checked
    ? Array.from(new Set([...node.value.network_keys, key]))
    : node.value.network_keys.filter((item) => item !== key)

  node.value.network_keys = next.length > 0 ? next : [key]
}

function addEnvEntry() {
  node.value.env_entries = [...node.value.env_entries, createEmptyEnvEntryDraft()]
}

function removeEnvEntry(uid: string) {
  node.value.env_entries = node.value.env_entries.filter((item) => item.uid !== uid)
}
</script>

<template>
  <section
    class="rounded-[24px] border bg-surface/70 p-4 shadow-[0_18px_40px_var(--color-shadow-soft)] transition"
    :class="selected ? 'border-primary ring-1 ring-primary/35' : 'border-border'"
  >
    <div class="flex items-start justify-between gap-3">
      <div class="flex items-center gap-3">
        <div
          class="flex h-11 w-11 items-center justify-center rounded-2xl border border-primary/20 bg-primary/10 text-primary"
        >
          <Server class="h-5 w-5" />
        </div>
        <div>
          <div class="text-xs font-semibold uppercase tracking-[0.22em] text-text-muted">
            Node {{ index + 1 }}
          </div>
          <div class="text-lg font-semibold text-text-primary">
            {{ node.name || node.key || '未命名节点' }}
          </div>
        </div>
      </div>

      <button
        v-if="removable"
        type="button"
        class="ui-btn ui-btn--danger topology-node-editor__danger-btn"
        @click="emit('remove')"
      >
        <Trash2 class="h-4 w-4" />
        删除节点
      </button>
    </div>

    <div class="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <label class="ui-field topology-node-editor__field">
        <span class="ui-field__label">节点 Key</span>
        <span class="ui-control-wrap">
          <input
            v-model="node.key"
            type="text"
            class="ui-control"
            placeholder="例如 web"
          />
        </span>
      </label>

      <label class="ui-field topology-node-editor__field">
        <span class="ui-field__label">显示名称</span>
        <span class="ui-control-wrap">
          <input
            v-model="node.name"
            type="text"
            class="ui-control"
            placeholder="例如 Web 应用"
          />
        </span>
      </label>

      <label class="ui-field topology-node-editor__field">
        <span class="ui-field__label">镜像</span>
        <span class="ui-control-wrap">
          <select v-model="node.image_id" class="ui-control">
            <option value="">复用题目主镜像</option>
            <option v-for="image in images" :key="image.id" :value="image.id">
              {{ image.name }}:{{ image.tag }}
            </option>
          </select>
        </span>
      </label>

      <label class="ui-field topology-node-editor__field">
        <span class="ui-field__label">服务端口</span>
        <span class="ui-control-wrap">
          <input
            :value="node.service_port ?? ''"
            type="number"
            min="1"
            max="65535"
            class="ui-control"
            placeholder="例如 8080"
            @input="updateNumberField('service_port', ($event.target as HTMLInputElement).value)"
          />
        </span>
      </label>

      <label class="ui-field topology-node-editor__field">
        <span class="ui-field__label">节点层级</span>
        <span class="ui-control-wrap">
          <select v-model="node.tier" class="ui-control">
            <option value="public">public</option>
            <option value="service">service</option>
            <option value="internal">internal</option>
          </select>
        </span>
      </label>

      <label
        class="topology-node-editor__check-row"
      >
        <input
          v-model="node.inject_flag"
          type="checkbox"
          class="topology-node-editor__checkbox"
        />
        启用 Flag 注入
      </label>
    </div>

    <div class="mt-5 space-y-3">
      <div class="text-sm font-medium text-text-primary">所属网络</div>
      <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        <label
          v-for="network in networks"
          :key="network.uid"
          class="topology-node-editor__check-row"
        >
          <input
            :checked="node.network_keys.includes(network.key)"
            type="checkbox"
            class="topology-node-editor__checkbox"
            @change="toggleNetwork(network.key, ($event.target as HTMLInputElement).checked)"
          />
          <div class="min-w-0">
            <div class="truncate font-medium">{{ network.name || network.key }}</div>
            <div class="truncate text-xs text-text-muted">{{ network.key }}</div>
          </div>
        </label>
      </div>
    </div>

    <div class="mt-5 space-y-3">
      <div class="flex items-center justify-between gap-3">
        <div class="text-sm font-medium text-text-primary">环境变量</div>
        <button
          type="button"
          class="ui-btn ui-btn--secondary topology-node-editor__secondary-btn"
          @click="addEnvEntry"
        >
          <Plus class="h-4 w-4" />
          添加变量
        </button>
      </div>

      <div
        v-if="node.env_entries.length === 0"
        class="rounded-xl border border-dashed border-border px-4 py-4 text-sm text-text-muted"
      >
        暂无环境变量
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="entry in node.env_entries"
          :key="entry.uid"
          class="topology-node-editor__env-row"
        >
          <span class="ui-control-wrap">
            <input v-model="entry.key" type="text" class="ui-control" placeholder="变量名" />
          </span>
          <span class="ui-control-wrap">
            <input v-model="entry.value" type="text" class="ui-control" placeholder="变量值" />
          </span>
          <button
            type="button"
            class="ui-btn ui-btn--danger topology-node-editor__env-remove"
            @click="removeEnvEntry(entry.uid)"
          >
            <Trash2 class="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>

    <div class="mt-5">
      <div class="mb-3 text-sm font-medium text-text-primary">资源限制</div>
      <div class="grid gap-4 md:grid-cols-3">
        <label class="ui-field topology-node-editor__field">
          <span class="ui-field__label">CPU Quota</span>
          <span class="ui-control-wrap">
            <input
              :value="node.cpu_quota ?? ''"
              type="number"
              min="0"
              step="0.1"
              class="ui-control"
              placeholder="例如 1"
              @input="updateNumberField('cpu_quota', ($event.target as HTMLInputElement).value)"
            />
          </span>
        </label>

        <label class="ui-field topology-node-editor__field">
          <span class="ui-field__label">Memory MB</span>
          <span class="ui-control-wrap">
            <input
              :value="node.memory_mb ?? ''"
              type="number"
              min="64"
              class="ui-control"
              placeholder="例如 256"
              @input="updateNumberField('memory_mb', ($event.target as HTMLInputElement).value)"
            />
          </span>
        </label>

        <label class="ui-field topology-node-editor__field">
          <span class="ui-field__label">Pids Limit</span>
          <span class="ui-control-wrap">
            <input
              :value="node.pids_limit ?? ''"
              type="number"
              min="1"
              class="ui-control"
              placeholder="例如 256"
              @input="updateNumberField('pids_limit', ($event.target as HTMLInputElement).value)"
            />
          </span>
        </label>
      </div>
    </div>
  </section>
</template>

<style scoped>
.topology-node-editor__field {
  --ui-field-gap: var(--space-2);
}

.topology-node-editor__check-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-height: 3rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 84%, transparent);
  border-radius: var(--ui-control-radius-md);
  background: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  padding: 0.75rem 0.875rem;
  font-size: var(--font-size-13);
  color: var(--color-text-primary);
}

.topology-node-editor__checkbox {
  width: 1rem;
  height: 1rem;
  flex-shrink: 0;
}

.topology-node-editor__secondary-btn,
.topology-node-editor__danger-btn {
  --ui-btn-padding: 0.55rem 0.9rem;
}

.topology-node-editor__env-row {
  display: grid;
  gap: var(--space-3);
  border: 1px solid color-mix(in srgb, var(--color-border-default) 84%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  padding: 0.75rem;
}

.topology-node-editor__env-remove {
  align-self: center;
  justify-self: stretch;
}

@media (min-width: 768px) {
  .topology-node-editor__env-row {
    grid-template-columns: minmax(0, 0.95fr) minmax(0, 1.05fr) auto;
  }

  .topology-node-editor__env-remove {
    justify-self: center;
  }
}
</style>
