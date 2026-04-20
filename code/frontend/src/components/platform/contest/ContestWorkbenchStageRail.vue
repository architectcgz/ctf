<script setup lang="ts">
import {
  Activity,
  Boxes,
  ClipboardCheck,
  Settings2,
  UsersRound,
  CheckCircle2,
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
  <nav class="studio-rail-v2" role="tablist">
    <div class="studio-nav-group pt-6">
      <button
        v-for="stage in stages"
        :key="stage.key"
        type="button"
        class="nav-item"
        :class="{
          'is-active': activeStage === stage.key,
          'is-disabled': stage.disabled,
        }"
        @click="handleStageSelect(stage)"
      >
        <div class="nav-item-inner">
          <div class="icon-stack">
            <component :is="stageIcons[stage.key] || Settings2" class="main-icon" />
          </div>
          <span class="label-text">{{ stage.label }}</span>
          
          <!-- 状态指示：如果是活跃阶段则显示发光点，如果是已完成建议显示 Check (逻辑可后续扩展) -->
          <div v-if="activeStage === stage.key" class="status-dot-active" />
        </div>
      </button>
    </div>
  </nav>
</template>

<style scoped>
.studio-rail-v2 {
  width: 15rem;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: transparent;
  backdrop-filter: none;
  -webkit-backdrop-filter: none;
  border-right: none;
  z-index: 10;
}

.studio-rail-brand {
  padding: 2rem 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.85rem;
}

.brand-hex {
  width: 2.25rem;
  height: 2.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-hover));
  border-radius: 0.75rem;
  box-shadow: 0 8px 16px color-mix(in srgb, var(--color-primary) 25%, transparent);
}

.brand-info {
  display: flex;
  flex-direction: column;
}

.brand-name {
  font-size: 0.95rem;
  font-weight: 900;
  color: var(--journal-ink);
  letter-spacing: -0.02em;
}

.brand-tag {
  font-size: 10px;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--journal-muted);
  opacity: 0.7;
}

.studio-nav-group {
  flex: 1;
  padding: 0 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.nav-item {
  width: 100%;
  border: none;
  background: transparent;
  padding: 0;
  cursor: pointer;
  outline: none;
}

.nav-item-inner {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.85rem;
  padding: 0.85rem 1rem;
  border-radius: 1rem;
  color: var(--journal-muted);
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.nav-item:hover:not(.is-disabled) .nav-item-inner {
  background: color-mix(in srgb, var(--journal-surface-subtle) 80%, var(--color-bg-base));
  color: var(--journal-ink);
}

.nav-item.is-active .nav-item-inner {
  background: var(--journal-surface);
  color: var(--color-primary);
  font-weight: 700;
  box-shadow: 
    0 1px 2px rgba(0, 0, 0, 0.05),
    0 0 0 1px rgba(255, 255, 255, 0.8);
}

[data-theme='dark'] .nav-item.is-active .nav-item-inner {
  box-shadow: 
    0 4px 12px rgba(0, 0, 0, 0.2),
    0 0 0 1px rgba(255, 255, 255, 0.05);
}

.icon-stack {
  width: 1.25rem;
  height: 1.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.main-icon {
  width: 100%;
  height: 100%;
  transition: transform 0.3s ease;
}

.nav-item.is-active .main-icon {
  transform: scale(1.15);
}

.label-text {
  font-size: 14px;
  letter-spacing: -0.01em;
}

.status-dot-active {
  position: absolute;
  left: 0.15rem;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 1rem;
  background: var(--color-primary);
  border-radius: 0 4px 4px 0;
  box-shadow: 0 0 10px var(--color-primary);
}

.is-disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
</style>
