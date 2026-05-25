<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next'

defineProps<{
  currentRoundLabel: string
  currentRoundStatusLabel: string
  teamName: string
  teamRank: string | number
  serviceCount: number
  topScore: number
  lastSyncedLabel: string
  loading: boolean
}>()

defineEmits<{
  refresh: []
}>()
</script>

<template>
  <header class="awd-hud-strip">
    <div class="hud-item">
      <div class="hud-label">当前回合</div>
      <div class="hud-value font-mono">
        {{ currentRoundLabel }}
      </div>
      <div class="hud-helper">
        {{ currentRoundStatusLabel }}
      </div>
    </div>
    <div class="hud-item">
      <div class="hud-label">我的战队</div>
      <div class="hud-value">
        {{ teamName }}
      </div>
      <div class="hud-helper">排名 #{{ teamRank }}</div>
    </div>
    <div class="hud-item">
      <div class="hud-label">战队服务</div>
      <div class="hud-value font-mono">
        {{ serviceCount }}
      </div>
      <div class="hud-helper">运行中服务</div>
    </div>
    <div class="hud-item">
      <div class="hud-label">最高分</div>
      <div class="hud-value hud-value--accent font-mono">
        {{ topScore }}
      </div>
      <div class="hud-helper">当前榜首</div>
    </div>
    <div class="hud-actions">
      <button class="hud-refresh-btn" :disabled="loading" @click="$emit('refresh')">
        <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': loading }" />
        <span>{{ lastSyncedLabel }}</span>
      </button>
    </div>
  </header>
</template>

<style scoped>
.awd-hud-strip {
  display: grid;
  grid-template-columns: repeat(4, 1fr) auto;
  gap: var(--space-4);
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  border-radius: 1rem;
  padding: 1.25rem 1.5rem;
  box-shadow: var(--color-shadow-soft);
}

.hud-item {
  display: flex;
  flex-direction: column;
}

.hud-label {
  font-size: var(--font-size-11);
  font-weight: 900;
  color: var(--color-text-muted);
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.hud-value {
  font-size: var(--font-size-24);
  font-weight: 900;
  color: var(--color-text-primary);
  margin: 0.25rem 0;
}

.hud-value--accent {
  color: var(--color-primary);
}

.hud-helper {
  font-size: var(--font-size-11);
  font-weight: 800;
  color: var(--color-primary);
}

.hud-refresh-btn {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  padding: 0 1.25rem;
  border-left: 1px solid var(--color-border-subtle);
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 800;
  cursor: pointer;
  background: transparent;
  transition: all 0.2s ease;
}

.hud-refresh-btn:hover {
  color: var(--color-text-primary);
}

@media (max-width: 1280px) {
  .awd-hud-strip {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .hud-actions {
    grid-column: 1 / -1;
  }

  .hud-refresh-btn {
    border-left: none;
    border-top: 1px solid var(--color-border-subtle);
    min-height: 4rem;
    padding: var(--space-3) 0 0;
  }
}

@media (max-width: 720px) {
  .awd-hud-strip {
    grid-template-columns: 1fr;
  }
}
</style>
