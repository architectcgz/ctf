<script setup lang="ts">
import TopologyStatusNotes from './TopologyStatusNotes.vue'
import TopologySummaryGrid from './TopologySummaryGrid.vue'

type TopologySummary = {
  networks: number
  nodes: number
  links: number
  policies: number
}

type TopologyStatusCard = {
  eyebrow: string
  title: string
  subtitle: string
}

defineProps<{
  heroEyebrow: string
  heroTitle: string
  heroDescription: string
  topologySummary: TopologySummary
  statusCard: TopologyStatusCard
  secondaryCard: TopologyStatusCard
}>()
</script>

<template>
  <section class="topology-hero-grid grid gap-4 xl:grid-cols-[1.04fr_0.96fr]">
    <div class="topology-hero-lead topology-hero-lead--library">
      <div class="topology-hero-kicker">
        <span>{{ heroEyebrow }}</span>
        <span class="topology-hero-badge">真实接口</span>
      </div>
      <h1 class="topology-hero-title">
        {{ heroTitle }}
      </h1>
      <p class="topology-hero-description">
        {{ heroDescription }}
      </p>

      <TopologySummaryGrid :summary="topologySummary" mode="template-library" />
    </div>

    <TopologyStatusNotes
      mode="template-library"
      :status-card="statusCard"
      :secondary-card="secondaryCard"
    />
  </section>
</template>

<style scoped>
.topology-hero-lead--library {
  padding: 0;
}

.topology-hero-kicker {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.topology-hero-badge {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 18%, transparent);
  border-radius: 0.6rem;
  background: color-mix(in srgb, var(--journal-accent) 7%, transparent);
  padding: var(--space-1) var(--space-2);
  color: var(--journal-accent);
}

.topology-hero-description {
  max-width: 46rem;
}

:deep(.topology-hero-aside--library > section) {
  padding-left: 0;
  background: transparent;
}

:deep(.topology-hero-aside--library > section h2) {
  font-size: var(--font-size-1-45);
}

:deep(.topology-hero-aside--library > section p) {
  color: var(--journal-muted);
}
</style>
