<script setup lang="ts">
import type { ContestWorkbenchStage, ContestWorkbenchStageKey } from '@/composables/useContestWorkbench'

const props = defineProps<{
  stages: ContestWorkbenchStage[]
  activeStage: ContestWorkbenchStageKey
  selectStage: (stage: ContestWorkbenchStageKey) => void
}>()

const tabButtonRefs: Partial<Record<ContestWorkbenchStageKey, HTMLButtonElement | null>> = {}

function getEnabledStageKeys(): ContestWorkbenchStageKey[] {
  return props.stages.filter((stage) => !stage.disabled).map((stage) => stage.key)
}

function setTabButtonRef(stage: ContestWorkbenchStageKey, element: HTMLButtonElement | null): void {
  tabButtonRefs[stage] = element
}

function focusStage(stage: ContestWorkbenchStageKey): void {
  tabButtonRefs[stage]?.focus()
}

function handleStageSelect(stage: ContestWorkbenchStage): void {
  if (stage.disabled) {
    return
  }
  props.selectStage(stage.key)
}

function handleStageKeydown(event: KeyboardEvent, currentStage: ContestWorkbenchStage): void {
  if (currentStage.disabled) {
    return
  }

  if (
    event.key !== 'ArrowRight' &&
    event.key !== 'ArrowLeft' &&
    event.key !== 'Home' &&
    event.key !== 'End'
  ) {
    return
  }

  const enabledStageKeys = getEnabledStageKeys()
  if (enabledStageKeys.length === 0) {
    return
  }

  event.preventDefault()

  if (event.key === 'Home') {
    const firstStage = enabledStageKeys[0]
    if (!firstStage) {
      return
    }
    props.selectStage(firstStage)
    focusStage(firstStage)
    return
  }

  if (event.key === 'End') {
    const lastStage = enabledStageKeys[enabledStageKeys.length - 1]
    if (!lastStage) {
      return
    }
    props.selectStage(lastStage)
    focusStage(lastStage)
    return
  }

  const currentIndex = enabledStageKeys.indexOf(currentStage.key)
  if (currentIndex === -1) {
    return
  }

  const direction = event.key === 'ArrowRight' ? 1 : -1
  const nextIndex = (currentIndex + direction + enabledStageKeys.length) % enabledStageKeys.length
  const nextStage = enabledStageKeys[nextIndex]
  if (!nextStage) {
    return
  }

  props.selectStage(nextStage)
  focusStage(nextStage)
}
</script>

<template>
  <nav class="top-tabs contest-workbench-stage-rail" role="tablist" aria-label="竞赛工作台阶段切换">
    <button
      v-for="stage in stages"
      :id="`contest-workbench-stage-tab-${stage.key}`"
      :key="stage.key"
      :ref="(element) => setTabButtonRef(stage.key, element as HTMLButtonElement | null)"
      type="button"
      role="tab"
      class="top-tab"
      :class="{
        active: activeStage === stage.key,
        'contest-workbench-stage-rail__tab--disabled': stage.disabled,
      }"
      :aria-selected="activeStage === stage.key ? 'true' : 'false'"
      :aria-controls="`contest-workbench-panel-${stage.key}`"
      :aria-disabled="stage.disabled ? 'true' : 'false'"
      :tabindex="!stage.disabled && activeStage === stage.key ? 0 : -1"
      :disabled="stage.disabled"
      @click="handleStageSelect(stage)"
      @keydown="handleStageKeydown($event, stage)"
    >
      {{ stage.label }}
    </button>
  </nav>
</template>

<style scoped>
.contest-workbench-stage-rail__tab--disabled {
  cursor: not-allowed;
  opacity: 0.5;
}
</style>
