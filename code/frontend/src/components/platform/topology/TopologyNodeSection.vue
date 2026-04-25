<script setup lang="ts">
import { defineAsyncComponent } from 'vue'
import { Plus } from 'lucide-vue-next'

import SectionCard from '@/components/common/SectionCard.vue'
import type { AdminImageListItem } from '@/api/contracts'
import type { TopologyNetworkDraft, TopologyNodeDraft } from './topologyDraft'

const TopologyNodeEditor = defineAsyncComponent(() => import('./TopologyNodeEditor.vue'))

const props = defineProps<{
  nodes: TopologyNodeDraft[]
  images: AdminImageListItem[]
  networks: TopologyNetworkDraft[]
  selectedNodeKey: string | null
  addButtonClass: string
}>()

const emit = defineEmits<{
  addNode: []
  removeNode: [uid: string]
  updateNode: [payload: { uid: string; node: TopologyNodeDraft }]
}>()
</script>

<template>
  <SectionCard title="节点编排" subtitle="节点支持单独镜像、资源限制、网络归属和环境变量。">
    <div class="space-y-4">
      <TopologyNodeEditor
        v-for="(node, index) in props.nodes"
        :key="node.uid"
        :data-node-editor="node.key"
        :model-value="node"
        :index="index"
        :images="props.images"
        :networks="props.networks"
        :removable="props.nodes.length > 1"
        :selected="props.selectedNodeKey === node.key"
        @update:model-value="emit('updateNode', { uid: node.uid, node: $event })"
        @remove="emit('removeNode', node.uid)"
      />
    </div>

    <template #footer>
      <button type="button" :class="addButtonClass" @click="emit('addNode')">
        <Plus class="h-4 w-4" />
        添加节点
      </button>
    </template>
  </SectionCard>
</template>
