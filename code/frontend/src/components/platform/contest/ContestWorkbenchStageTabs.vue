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
}

function handleStageSelect(stage: ContestWorkbenchStage): void {
  if (stage.disabled) return
  props.selectStage(stage.key)
}
</script>

<template>
  <nav
    class="studio-tabs"
    role="tablist"
    aria-label="竞赛工作台阶段切换"
  >
    <div class="studio-tabs-list">
      <button
        v-for="stage in stages"
        :id="`contest-workbench-stage-tab-${stage.key}`"
        :key="stage.key"
        type="button"
        class="studio-tab-item"
        role="tab"
        :aria-selected="activeStage === stage.key"
        :aria-disabled="stage.disabled ? 'true' : 'false'"
        :tabindex="activeStage === stage.key ? 0 : -1"
        :class="{
          'is-active': activeStage === stage.key,
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
        <div
          v-if="activeStage === stage.key"
          class="tab-active-indicator"
        />
      </button>
    </div>
  </nav>
</template>

<style scoped>
.studio-tabs {
  background: var(--color-bg-base);
  border-bottom: 1px solid var(--color-border-default);
  padding: 0 2rem;
}

.studio-tabs-list {
  display: flex;
  gap: 2rem;
  height: 3.5rem;
}

.studio-tab-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.65rem;
  height: 100%;
  padding: 0 0.25rem;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s ease;
}

.studio-tab-item:hover:not(.is-disabled) {
  color: var(--color-text-primary);
}

.studio-tab-item.is-active {
  color: var(--color-primary);
}

.tab-icon {
  width: 1rem;
  height: 1rem;
}

.tab-active-indicator {
  position: absolute;
  bottom: -1px;
  left: 0;
  right: 0;
  height: 2px;
  background: var(--color-primary);
  box-shadow: 0 0 8px var(--color-primary);
}

.is-disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
</style>
