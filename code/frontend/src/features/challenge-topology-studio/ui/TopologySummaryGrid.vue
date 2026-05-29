<script setup lang="ts">
interface TopologySummary {
  networks: number
  nodes: number
  links: number
  policies: number
}

const props = defineProps<{
  summary: TopologySummary
  mode: 'template-library' | 'challenge'
}>()

const items = [
  { key: 'networks', label: '网络', helper: '当前模板草稿中的网络数量' },
  { key: 'nodes', label: '节点', helper: '当前模板草稿中的节点数量' },
  { key: 'links', label: '连线', helper: '当前模板草稿中的连线数量' },
  { key: 'policies', label: '策略', helper: '当前模板草稿中的策略数量' },
] as const
</script>

<template>
  <div
    :class="
      mode === 'template-library'
        ? 'topology-summary-grid progress-strip metric-panel-grid metric-panel-default-surface'
        : 'topology-summary-grid topology-summary-grid--challenge metric-panel-grid'
    "
  >
    <component
      :is="mode === 'template-library' ? 'div' : 'article'"
      v-for="item in items"
      :key="item.key"
      :class="
        mode === 'template-library'
          ? 'topology-summary-tile progress-card metric-panel-card'
          : 'topology-summary-card metric-panel-card'
      "
    >
      <div
        :class="
          mode === 'template-library'
            ? 'topology-summary-label progress-card-label metric-panel-label'
            : 'topology-summary-label metric-panel-label'
        "
      >
        {{ item.label }}
      </div>
      <div
        :class="
          mode === 'template-library'
            ? 'topology-summary-value progress-card-value metric-panel-value'
            : 'topology-summary-value metric-panel-value'
        "
      >
        {{ props.summary[item.key] }}
      </div>
      <div
        v-if="mode === 'template-library'"
        class="topology-summary-helper progress-card-hint metric-panel-helper"
      >
        {{ item.helper }}
      </div>
    </component>
  </div>
</template>

<style scoped>
.topology-summary-grid {
  --metric-panel-grid-gap: var(--space-3);
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
}

.topology-summary-grid:not(.topology-summary-grid--challenge) {
  margin-top: var(--space-6);
}

.topology-summary-grid--challenge {
  margin-top: var(--space-2);
}

.topology-summary-card {
  border: 1px solid var(--journal-border);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--topology-panel) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--topology-panel-subtle) 96%, var(--color-bg-base))
  );
  box-shadow: 0 10px 24px var(--color-shadow-soft);
}

@media (max-width: 1023px) {
  .topology-summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .topology-summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>
