<script setup lang="ts">
import {
  Activity,
  Boxes,
  ClipboardCheck,
  Settings2,
  UsersRound,
} from 'lucide-vue-next'
import type { Component } from 'vue'
import type { ContestWorkbenchStage, ContestWorkbenchStageKey } from '@/composables/useContestWorkbench'

const props = defineProps<{
  stages: ContestWorkbenchStage[]
  activeStage: ContestWorkbenchStageKey
  selectStage: (stage: ContestWorkbenchStageKey) => void
}>()

const stageIcons: Record<string, Component> = {
  basics: Settings2,
  pool: Boxes,
  teams: UsersRound,
  preflight: ClipboardCheck,
  operations: Activity,
  instances: Boxes,
}

function handleStageSelect(stage: ContestWorkbenchStage): void {
  if (stage.disabled) return
  props.selectStage(stage.key)
}
</script>

<template>
  <nav
    class="top-tabs studio-tabs-container"
    role="tablist"
    aria-label="竞赛工作台阶段切换"
  >
    <button
      v-for="stage in stages"
      :id="`contest-workbench-stage-tab-${stage.key}`"
      :key="stage.key"
      type="button"
      class="top-tab"
      role="tab"
      :aria-selected="activeStage === stage.key"
      :aria-disabled="stage.disabled ? 'true' : 'false'"
      :tabindex="activeStage === stage.key ? 0 : -1"
      :class="{
        active: activeStage === stage.key,
        'is-disabled': stage.disabled,
      }"
      :disabled="stage.disabled"
      @click="handleStageSelect(stage)"
    >
      <component
        :is="stageIcons[stage.key] || Settings2"
        class="tab-icon"
      />
      <span class="tab-label">{{ stage.label }}</span>
    </button>
  </nav>
</template>

<style scoped>
.studio-tabs-container {
  background: var(--color-bg-surface);
  margin-top: 0;
  padding: 0 var(--space-workspace-side-padding);
  border-top: 0;
  border-bottom-color: var(--workspace-line-soft);
}

.top-tab {
  gap: 0.65rem;
  transition:
    color 180ms cubic-bezier(0.25, 1, 0.5, 1),
    border-color 180ms cubic-bezier(0.25, 1, 0.5, 1),
    background 180ms cubic-bezier(0.25, 1, 0.5, 1);
}

.top-tab.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

.tab-icon {
  width: 1rem;
  height: 1rem;
  transition:
    color 180ms cubic-bezier(0.25, 1, 0.5, 1),
    transform 180ms cubic-bezier(0.25, 1, 0.5, 1);
}

.top-tab.active .tab-icon {
  transform: translateY(calc(var(--space-0-5) * -1));
}

.is-disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

@media (prefers-reduced-motion: reduce) {
  .top-tab,
  .tab-icon {
    transition: none;
  }

  .top-tab.active .tab-icon {
    transform: none;
  }
}
</style>
