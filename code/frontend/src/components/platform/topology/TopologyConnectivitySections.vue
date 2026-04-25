<script setup lang="ts">
import { Link2, ShieldBan, Trash2 } from 'lucide-vue-next'

import SectionCard from '@/components/common/SectionCard.vue'
import type { TopologyLinkDraft, TopologyPolicyDraft } from './topologyDraft'

type NodeOption = {
  key: string
  label: string
}

type LinkPatch = Partial<Pick<TopologyLinkDraft, 'from_node_key' | 'to_node_key'>>
type PolicyPatch = Partial<
  Pick<TopologyPolicyDraft, 'source_node_key' | 'target_node_key' | 'action'>
>

const props = defineProps<{
  links: TopologyLinkDraft[]
  policies: TopologyPolicyDraft[]
  nodeOptions: NodeOption[]
  addButtonClass: string
}>()

const emit = defineEmits<{
  addLink: []
  removeLink: [uid: string]
  updateLink: [payload: { uid: string; patch: LinkPatch }]
  addPolicy: []
  removePolicy: [uid: string]
  updatePolicy: [payload: { uid: string; patch: PolicyPatch }]
}>()

function updateLink(uid: string, patch: LinkPatch) {
  emit('updateLink', { uid, patch })
}

function updatePolicy(uid: string, patch: PolicyPatch) {
  emit('updatePolicy', { uid, patch })
}
</script>

<template>
  <SectionCard title="拓扑连线" subtitle="用于表达逻辑依赖关系，不直接等同于运行时 ACL。">
    <div
      v-if="props.links.length === 0"
      class="rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted"
    >
      暂无逻辑连线
    </div>
    <div v-else class="space-y-3">
      <div
        v-for="link in props.links"
        :key="link.uid"
        class="grid gap-3 rounded-2xl border border-border bg-elevated p-4 md:grid-cols-[1fr_1fr_auto]"
      >
        <select
          :value="link.from_node_key"
          class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
          @change="
            updateLink(link.uid, { from_node_key: ($event.target as HTMLSelectElement).value })
          "
        >
          <option value="">选择源节点</option>
          <option v-for="node in props.nodeOptions" :key="node.key" :value="node.key">
            {{ node.label }}
          </option>
        </select>
        <select
          :value="link.to_node_key"
          class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
          @change="
            updateLink(link.uid, { to_node_key: ($event.target as HTMLSelectElement).value })
          "
        >
          <option value="">选择目标节点</option>
          <option v-for="node in props.nodeOptions" :key="node.key" :value="node.key">
            {{ node.label }}
          </option>
        </select>
        <button
          type="button"
          class="ui-btn ui-btn--danger topology-action-btn topology-action-btn--icon"
          @click="emit('removeLink', link.uid)"
        >
          <Trash2 class="h-4 w-4" />
        </button>
      </div>
    </div>

    <template #footer>
      <button type="button" :class="addButtonClass" @click="emit('addLink')">
        <Link2 class="h-4 w-4" />
        添加连线
      </button>
    </template>
  </SectionCard>

  <SectionCard
    title="链路策略"
    subtitle="当前前端只开放粗粒度节点 allow/deny，细粒度端口策略尚未支持。"
  >
    <div
      v-if="props.policies.length === 0"
      class="rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted"
    >
      暂无链路策略
    </div>
    <div v-else class="space-y-3">
      <div
        v-for="policy in props.policies"
        :key="policy.uid"
        class="grid gap-3 rounded-2xl border border-border bg-elevated p-4 md:grid-cols-[1fr_1fr_0.7fr_auto]"
      >
        <select
          :value="policy.source_node_key"
          class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
          @change="
            updatePolicy(policy.uid, {
              source_node_key: ($event.target as HTMLSelectElement).value,
            })
          "
        >
          <option value="">选择源节点</option>
          <option v-for="node in props.nodeOptions" :key="node.key" :value="node.key">
            {{ node.label }}
          </option>
        </select>
        <select
          :value="policy.target_node_key"
          class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
          @change="
            updatePolicy(policy.uid, {
              target_node_key: ($event.target as HTMLSelectElement).value,
            })
          "
        >
          <option value="">选择目标节点</option>
          <option v-for="node in props.nodeOptions" :key="node.key" :value="node.key">
            {{ node.label }}
          </option>
        </select>
        <select
          :value="policy.action"
          class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
          @change="
            updatePolicy(policy.uid, {
              action: ($event.target as HTMLSelectElement).value as TopologyPolicyDraft['action'],
            })
          "
        >
          <option value="allow">allow</option>
          <option value="deny">deny</option>
        </select>
        <button
          type="button"
          class="ui-btn ui-btn--danger topology-action-btn topology-action-btn--icon"
          @click="emit('removePolicy', policy.uid)"
        >
          <Trash2 class="h-4 w-4" />
        </button>
      </div>
    </div>

    <template #footer>
      <button type="button" :class="addButtonClass" @click="emit('addPolicy')">
        <ShieldBan class="h-4 w-4" />
        添加策略
      </button>
    </template>
  </SectionCard>
</template>
