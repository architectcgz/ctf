<script setup lang="ts">
import { Trash2 } from 'lucide-vue-next'

import SectionCard from '@/shared/ui/common/SectionCard.vue'

type NodeOption = {
  key: string
  label: string
}

const props = withDefaults(
  defineProps<{
    entryNodeKey: string
    nodeOptions: NodeOption[]
    showDeleteAction?: boolean
    deleteDisabled?: boolean
  }>(),
  {
    showDeleteAction: false,
    deleteDisabled: false,
  }
)

const emit = defineEmits<{
  updateEntryNodeKey: [value: string]
  deleteTopology: []
}>()
</script>

<template>
  <SectionCard title="入口节点" subtitle="实例访问入口和当前草稿的保存范围。">
    <div class="grid gap-4 md:grid-cols-[1fr_auto]">
      <label class="space-y-2">
        <span class="text-sm text-text-secondary">入口节点</span>
        <select
          :value="props.entryNodeKey"
          class="w-full rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary outline-none transition focus:border-primary"
          @change="emit('updateEntryNodeKey', ($event.target as HTMLSelectElement).value)"
        >
          <option v-for="node in props.nodeOptions" :key="node.key" :value="node.key">
            {{ node.label }} ({{ node.key }})
          </option>
        </select>
      </label>

      <button
        v-if="props.showDeleteAction"
        type="button"
        class="ui-btn ui-btn--danger self-end"
        :disabled="props.deleteDisabled"
        @click="emit('deleteTopology')"
      >
        <Trash2 class="h-4 w-4" />
        删除已保存拓扑
      </button>
    </div>
  </SectionCard>
</template>
