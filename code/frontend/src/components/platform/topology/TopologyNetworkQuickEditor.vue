<script setup lang="ts">
type EditableNetworkDraft = {
  uid: string
  key: string
  name: string
  internal: boolean
}

type NetworkPatch = Partial<Pick<EditableNetworkDraft, 'key' | 'name' | 'internal'>>

const props = defineProps<{
  networks: EditableNetworkDraft[]
}>()

const emit = defineEmits<{
  updateNetwork: [payload: { uid: string; patch: NetworkPatch }]
}>()

function updateNetwork(uid: string, patch: NetworkPatch) {
  emit('updateNetwork', { uid, patch })
}
</script>

<template>
  <div class="rounded-2xl border border-border bg-elevated p-4">
    <div class="text-sm font-semibold text-text-primary">网络快速编辑</div>
    <div class="mt-3 space-y-3">
      <div
        v-for="network in props.networks"
        :key="network.uid"
        class="grid gap-3 rounded-xl border border-border bg-surface p-3 md:grid-cols-[0.9fr_1fr_auto]"
      >
        <input
          :value="network.key"
          type="text"
          class="w-full rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
          placeholder="network key"
          @input="
            updateNetwork(network.uid, { key: ($event.target as HTMLInputElement).value })
          "
        />
        <input
          :value="network.name"
          type="text"
          class="w-full rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
          placeholder="网络名称"
          @input="
            updateNetwork(network.uid, { name: ($event.target as HTMLInputElement).value })
          "
        />
        <label
          class="flex items-center gap-2 rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary"
        >
          <input
            :checked="network.internal"
            type="checkbox"
            class="h-4 w-4 rounded border-border bg-transparent"
            @change="
              updateNetwork(network.uid, {
                internal: ($event.target as HTMLInputElement).checked,
              })
            "
          />
          internal
        </label>
      </div>
    </div>
  </div>
</template>
