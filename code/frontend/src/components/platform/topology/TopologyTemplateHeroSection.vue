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
