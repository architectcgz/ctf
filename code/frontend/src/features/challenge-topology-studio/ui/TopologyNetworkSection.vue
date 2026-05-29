<script setup lang="ts">
import { Plus, Trash2 } from 'lucide-vue-next'

import SectionCard from '@/components/common/SectionCard.vue'
import type { TopologyNetworkDraft } from '@/features/challenge-topology-studio/model'

type NetworkPatch = Partial<Pick<TopologyNetworkDraft, 'key' | 'name' | 'cidr' | 'internal'>>

const props = defineProps<{
  networks: TopologyNetworkDraft[]
  addButtonClass: string
}>()

const emit = defineEmits<{
  addNetwork: []
  removeNetwork: [uid: string]
  updateNetwork: [payload: { uid: string; patch: NetworkPatch }]
}>()

function updateNetwork(uid: string, patch: NetworkPatch) {
  emit('updateNetwork', { uid, patch })
}
</script>

<template>
  <SectionCard
    title="网络分段"
    subtitle="一个节点可以挂多个网络，运行时会创建多个 Docker Network。"
  >
    <div class="space-y-3">
      <div
        v-for="network in props.networks"
        :key="network.uid"
        class="grid gap-3 rounded-2xl border border-border bg-elevated p-4 md:grid-cols-[0.9fr_1fr_0.9fr_auto_auto]"
      >
        <input
          :value="network.key"
          type="text"
          class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
          placeholder="network key"
          @input="updateNetwork(network.uid, { key: ($event.target as HTMLInputElement).value })"
        />
        <input
          :value="network.name"
          type="text"
          class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
          placeholder="网络名称"
          @input="updateNetwork(network.uid, { name: ($event.target as HTMLInputElement).value })"
        />
        <input
          :value="network.cidr"
          type="text"
          class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
          placeholder="CIDR（可选）"
          @input="updateNetwork(network.uid, { cidr: ($event.target as HTMLInputElement).value })"
        />
        <label
          class="flex items-center gap-3 rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary"
        >
          <input
            :checked="network.internal"
            type="checkbox"
            class="h-4 w-4 rounded border-border bg-transparent"
            @change="
              updateNetwork(network.uid, { internal: ($event.target as HTMLInputElement).checked })
            "
          />
          internal
        </label>
        <button
          type="button"
          class="ui-btn ui-btn--danger topology-action-btn topology-action-btn--icon"
          :disabled="props.networks.length <= 1"
          @click="emit('removeNetwork', network.uid)"
        >
          <Trash2 class="h-4 w-4" />
        </button>
      </div>
    </div>

    <template #footer>
      <button type="button" :class="addButtonClass" @click="emit('addNetwork')">
        <Plus class="h-4 w-4" />
        添加网络
      </button>
    </template>
  </SectionCard>
</template>
