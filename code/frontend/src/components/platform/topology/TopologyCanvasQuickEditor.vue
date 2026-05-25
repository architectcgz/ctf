<script setup lang="ts">
import { computed } from 'vue'

import type { AdminImageListItem } from '@/api/contracts'
import type { TopologyTier } from '@/api/contracts'

type NodeOption = {
  key: string
  label: string
}

type EditableNetworkDraft = {
  uid: string
  key: string
  name: string
}

type EditableNodeDraft = {
  name: string
  image_id: string
  service_port: number | null
  inject_flag: boolean
  tier: TopologyTier
  network_keys: string[]
}

type EdgeKind = 'link' | 'allow' | 'deny'

type SelectedNodeField = 'name' | 'image_id' | 'tier' | 'inject_flag'

const props = withDefaults(
  defineProps<{
    variant?: 'template' | 'challenge'
    emptyMessage?: string
    selectedNodeDraft: EditableNodeDraft | null
    hasSelectedEdge: boolean
    nodeOptions: NodeOption[]
    networks: EditableNetworkDraft[]
    images?: AdminImageListItem[]
    selectedEdgeSourceKey: string
    selectedEdgeTargetKey: string
    selectedEdgeKind: EdgeKind
  }>(),
  {
    variant: 'challenge',
    emptyMessage: '请选择一个节点或一条边',
    images: () => [],
  }
)

const emit = defineEmits<{
  updateSelectedNodeField: [payload: { field: SelectedNodeField; value: string | boolean }]
  updateSelectedNodeServicePort: [value: string]
  toggleSelectedNodeNetwork: [payload: { networkKey: string; checked: boolean }]
  updateSelectedEdgeSourceKey: [value: string]
  updateSelectedEdgeTargetKey: [value: string]
  updateSelectedEdgeKind: [value: string]
}>()

const isTemplateVariant = computed(() => props.variant === 'template')

const rootClass = computed(() =>
  isTemplateVariant.value
    ? 'template-quick-editor rounded-2xl border border-border bg-elevated p-4'
    : 'rounded-2xl border border-border bg-elevated p-4'
)

const networkListClass = computed(() =>
  isTemplateVariant.value ? 'flex flex-wrap gap-2' : 'grid gap-2 md:grid-cols-2'
)

const networkLabelClass = computed(() =>
  isTemplateVariant.value
    ? 'flex items-center gap-2 rounded-xl border border-border bg-surface px-3 py-2 text-sm text-text-primary transition hover:border-primary'
    : 'flex items-center gap-3 rounded-xl border border-border bg-surface px-3 py-3 text-sm text-text-primary'
)

function updateSelectedNodeField(field: SelectedNodeField, value: string | boolean) {
  emit('updateSelectedNodeField', { field, value })
}
</script>

<template>
  <div :class="rootClass">
    <div class="text-sm font-semibold text-text-primary">画布快速编辑</div>

    <div
      v-if="!props.selectedNodeDraft && !props.hasSelectedEdge"
      class="mt-3 rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted"
    >
      {{ props.emptyMessage }}
    </div>

    <div v-else-if="props.selectedNodeDraft" class="mt-3 space-y-4">
      <div class="grid gap-3 md:grid-cols-2">
        <label class="space-y-2">
          <span class="text-sm text-text-secondary">节点名称</span>
          <input
            :value="props.selectedNodeDraft.name"
            type="text"
            class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
            @input="
              updateSelectedNodeField('name', ($event.target as HTMLInputElement).value)
            "
          />
        </label>
        <label v-if="!isTemplateVariant" class="space-y-2">
          <span class="text-sm text-text-secondary">镜像</span>
          <select
            :value="props.selectedNodeDraft.image_id"
            class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
            @change="
              updateSelectedNodeField('image_id', ($event.target as HTMLSelectElement).value)
            "
          >
            <option value="">复用题目主镜像</option>
            <option v-for="image in props.images" :key="image.id" :value="image.id">
              {{ image.name }}:{{ image.tag }}
            </option>
          </select>
        </label>
        <label v-if="!isTemplateVariant" class="space-y-2">
          <span class="text-sm text-text-secondary">层级</span>
          <select
            :value="props.selectedNodeDraft.tier"
            class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
            @change="
              updateSelectedNodeField('tier', ($event.target as HTMLSelectElement).value)
            "
          >
            <option value="public">public</option>
            <option value="service">service</option>
            <option value="internal">internal</option>
          </select>
        </label>
        <label class="space-y-2">
          <span class="text-sm text-text-secondary">服务端口</span>
          <input
            :value="props.selectedNodeDraft.service_port ?? ''"
            type="number"
            min="1"
            max="65535"
            class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
            @input="
              emit('updateSelectedNodeServicePort', ($event.target as HTMLInputElement).value)
            "
          />
        </label>
      </div>

      <label
        v-if="!isTemplateVariant"
        class="flex items-center gap-3 rounded-xl border border-border bg-surface px-3 py-3 text-sm text-text-primary"
      >
        <input
          :checked="Boolean(props.selectedNodeDraft.inject_flag)"
          type="checkbox"
          class="h-4 w-4 rounded border-border bg-transparent"
          @change="
            updateSelectedNodeField(
              'inject_flag',
              ($event.target as HTMLInputElement).checked
            )
          "
        />
        启用 Flag 注入
      </label>

      <div class="space-y-2">
        <div class="text-sm text-text-secondary">所属网络</div>
        <div :class="networkListClass">
          <label
            v-for="network in props.networks"
            :key="network.uid"
            :class="networkLabelClass"
          >
            <input
              :checked="props.selectedNodeDraft.network_keys.includes(network.key)"
              type="checkbox"
              class="h-4 w-4 rounded border-border bg-transparent"
              @change="
                emit('toggleSelectedNodeNetwork', {
                  networkKey: network.key,
                  checked: ($event.target as HTMLInputElement).checked,
                })
              "
            />
            <span>{{ network.name || network.key }}</span>
          </label>
        </div>
      </div>
    </div>

    <div v-else class="mt-3 space-y-4">
      <div class="grid gap-3 md:grid-cols-2">
        <label class="space-y-2">
          <span class="text-sm text-text-secondary">源节点</span>
          <select
            :value="props.selectedEdgeSourceKey"
            class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
            @change="emit('updateSelectedEdgeSourceKey', ($event.target as HTMLSelectElement).value)"
          >
            <option v-for="node in props.nodeOptions" :key="node.key" :value="node.key">
              {{ node.label }}
            </option>
          </select>
        </label>
        <label class="space-y-2">
          <span class="text-sm text-text-secondary">目标节点</span>
          <select
            :value="props.selectedEdgeTargetKey"
            class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
            @change="emit('updateSelectedEdgeTargetKey', ($event.target as HTMLSelectElement).value)"
          >
            <option v-for="node in props.nodeOptions" :key="node.key" :value="node.key">
              {{ node.label }}
            </option>
          </select>
        </label>
      </div>

      <label class="space-y-2">
        <span class="text-sm text-text-secondary">边类型</span>
        <select
          :value="props.selectedEdgeKind"
          class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
          @change="emit('updateSelectedEdgeKind', ($event.target as HTMLSelectElement).value)"
        >
          <option value="link">logic link</option>
          <option value="allow">allow</option>
          <option value="deny">deny</option>
        </select>
      </label>
    </div>
  </div>
</template>
