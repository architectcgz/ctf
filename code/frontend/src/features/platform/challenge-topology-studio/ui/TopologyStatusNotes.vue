<script setup lang="ts">
import { Blocks, GitBranch, ShieldBan } from 'lucide-vue-next'

interface TopologyStatusCard {
  eyebrow: string
  title: string
  subtitle: string
}

defineProps<{
  mode: 'template-library' | 'challenge'
  statusCard: TopologyStatusCard
  secondaryCard: TopologyStatusCard
}>()
</script>

<template>
  <div
    v-if="mode === 'template-library'"
    class="topology-hero-aside topology-hero-aside--library grid gap-3 md:grid-cols-3 xl:grid-cols-1"
  >
    <section class="template-hero-note template-hero-note--primary">
      <div class="template-metric-icon template-metric-icon--primary">
        <Blocks class="h-5 w-5" />
      </div>
      <div class="template-hero-note__body">
        <div class="template-hero-note__label">
          {{ statusCard.eyebrow }}
        </div>
        <div class="template-hero-note__value">
          {{ statusCard.title }}
        </div>
        <p class="template-hero-note__copy">
          {{ statusCard.subtitle }}
        </p>
      </div>
    </section>

    <section class="template-hero-note template-hero-note--warning">
      <div class="template-metric-icon template-metric-icon--warning">
        <GitBranch class="h-5 w-5" />
      </div>
      <div class="template-hero-note__body">
        <div class="template-hero-note__label">
          {{ secondaryCard.eyebrow }}
        </div>
        <div class="template-hero-note__value">
          {{ secondaryCard.title }}
        </div>
        <p class="template-hero-note__copy">
          {{ secondaryCard.subtitle }}
        </p>
      </div>
    </section>
  </div>

  <section v-else class="topology-status-list">
    <article class="topology-status-note topology-status-note--primary">
      <div class="topology-status-note__icon">
        <Blocks class="h-5 w-5" />
      </div>
      <div class="topology-status-note__body">
        <div class="topology-status-note__eyebrow">
          {{ statusCard.eyebrow }}
        </div>
        <div class="topology-status-note__title">
          {{ statusCard.title }}
        </div>
        <p class="topology-status-note__copy">
          {{ statusCard.subtitle }}
        </p>
      </div>
    </article>

    <article class="topology-status-note topology-status-note--warning">
      <div class="topology-status-note__icon">
        <GitBranch class="h-5 w-5" />
      </div>
      <div class="topology-status-note__body">
        <div class="topology-status-note__eyebrow">
          {{ secondaryCard.eyebrow }}
        </div>
        <div class="topology-status-note__title">
          {{ secondaryCard.title }}
        </div>
        <p class="topology-status-note__copy">
          {{ secondaryCard.subtitle }}
        </p>
      </div>
    </article>

    <article class="topology-status-note topology-status-note--danger">
      <div class="topology-status-note__icon">
        <ShieldBan class="h-5 w-5" />
      </div>
      <div class="topology-status-note__body">
        <div class="topology-status-note__eyebrow">运行时约束</div>
        <div class="topology-status-note__title">粗粒度</div>
        <p class="topology-status-note__copy">当前只支持节点级 allow/deny，不支持端口级 ACL。</p>
      </div>
    </article>
  </section>
</template>

<style scoped>
.topology-hero-aside--library {
  align-self: start;
  border-left: 0;
  padding-left: 0;
}

.template-hero-note {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-3);
  padding: 0 0 0 var(--space-4);
  border-left: 2px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.template-hero-note__body {
  min-width: 0;
}

.template-hero-note__label,
.topology-status-note__eyebrow {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.template-hero-note__value {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-1-10);
  font-weight: 700;
  color: var(--journal-ink);
}

.template-hero-note__copy {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-0-86);
  line-height: 1.6;
  color: var(--journal-muted);
}

.template-metric-icon,
.topology-status-note__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 0.9rem;
}

.template-metric-icon--primary,
.topology-status-note--primary .topology-status-note__icon {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 22%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
}

.template-metric-icon--warning,
.topology-status-note--warning .topology-status-note__icon {
  border: 1px solid color-mix(in srgb, var(--color-warning) 24%, transparent);
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
  color: var(--color-warning);
}

.topology-status-note--danger .topology-status-note__icon {
  border: 1px solid color-mix(in srgb, var(--color-danger) 24%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, transparent);
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.topology-status-list {
  display: grid;
  gap: var(--space-3);
}

.topology-status-note {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-3);
  padding: var(--space-4);
  border: 1px solid var(--journal-border);
  border-radius: 18px;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--topology-panel) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--topology-panel-subtle) 96%, var(--color-bg-base))
  );
  box-shadow: 0 12px 28px var(--color-shadow-soft);
}

.topology-status-note__title {
  margin-top: var(--space-1);
  font-size: var(--font-size-1-10);
  font-weight: 700;
  color: var(--journal-ink);
}

.topology-status-note__copy {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-0-86);
  line-height: 1.65;
  color: var(--journal-muted);
}

@media (max-width: 1023px) {
  .topology-hero-aside--library {
    border-left: 0;
    padding-left: 0;
  }
}
</style>
