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
        class="inline-flex items-center gap-2 rounded-xl border border-danger/30 bg-danger/10 px-3 py-2 text-sm text-danger transition hover:bg-danger/15"
        @click="emit('remove')"
      >
        <Trash2 class="h-4 w-4" />
        删除节点
      </button>
    </div>

    <div class="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <label class="space-y-2">
        <span class="text-sm text-text-secondary">节点 Key</span>
        <input
          v-model="node.key"
          type="text"
          class="w-full rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
          placeholder="例如 web"
        />
      </label>

      <label class="space-y-2">
        <span class="text-sm text-text-secondary">显示名称</span>
        <input
          v-model="node.name"
          type="text"
          class="w-full rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
          placeholder="例如 Web 应用"
        />
      </label>

      <label class="space-y-2">
        <span class="text-sm text-text-secondary">镜像</span>
        <select
          v-model="node.image_id"
          class="w-full rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
        >
          <option value="">复用题目主镜像</option>
          <option v-for="image in images" :key="image.id" :value="image.id">
            {{ image.name }}:{{ image.tag }}
          </option>
        </select>
      </label>

      <label class="space-y-2">
        <span class="text-sm text-text-secondary">服务端口</span>
        <input
          :value="node.service_port ?? ''"
          type="number"
          min="1"
          max="65535"
          class="w-full rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
          placeholder="例如 8080"
          @input="updateNumberField('service_port', ($event.target as HTMLInputElement).value)"
        />
      </label>

      <label class="space-y-2">
        <span class="text-sm text-text-secondary">节点层级</span>
        <select
          v-model="node.tier"
          class="w-full rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
        >
          <option value="public">public</option>
          <option value="service">service</option>
          <option value="internal">internal</option>
        </select>
      </label>

      <label
        class="flex items-center gap-3 rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary"
      >
        <input
          v-model="node.inject_flag"
          type="checkbox"
          class="h-4 w-4 rounded border-border bg-transparent"
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
          class="flex items-center gap-3 rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary"
        >
          <input
            :checked="node.network_keys.includes(network.key)"
            type="checkbox"
            class="h-4 w-4 rounded border-border bg-transparent"
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
          class="inline-flex items-center gap-2 rounded-xl border border-border px-3 py-2 text-sm text-text-primary transition hover:border-primary"
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
          class="grid gap-3 rounded-xl border border-border bg-elevated p-3 md:grid-cols-[0.95fr_1.05fr_auto]"
        >
          <input
            v-model="entry.key"
            type="text"
            class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
            placeholder="变量名"
          />
          <input
            v-model="entry.value"
            type="text"
            class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
            placeholder="变量值"
          />
          <button
            type="button"
            class="inline-flex items-center justify-center rounded-xl border border-danger/30 bg-danger/10 px-3 py-2 text-sm text-danger transition hover:bg-danger/15"
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
        <label class="space-y-2">
          <span class="text-sm text-text-secondary">CPU Quota</span>
          <input
            :value="node.cpu_quota ?? ''"
            type="number"
            min="0"
            step="0.1"
            class="w-full rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
            placeholder="例如 1"
            @input="updateNumberField('cpu_quota', ($event.target as HTMLInputElement).value)"
          />
        </label>

        <label class="space-y-2">
          <span class="text-sm text-text-secondary">Memory MB</span>
          <input
            :value="node.memory_mb ?? ''"
            type="number"
            min="64"
            class="w-full rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
            placeholder="例如 256"
            @input="updateNumberField('memory_mb', ($event.target as HTMLInputElement).value)"
          />
        </label>

        <label class="space-y-2">
          <span class="text-sm text-text-secondary">Pids Limit</span>
          <input
            :value="node.pids_limit ?? ''"
            type="number"
            min="1"
            class="w-full rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
            placeholder="例如 256"
            @input="updateNumberField('pids_limit', ($event.target as HTMLInputElement).value)"
          />
        </label>
      </div>
    </div>
  </section>
</template>
