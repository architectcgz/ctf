<script setup lang="ts">
import type { AdminContestChallengeViewData, AWDReadinessData, ContestDetailData } from '@/api/contracts'
import type { ContestFieldLocks, ContestFormDraft, PlatformContestStatus } from '@/composables/usePlatformContests'
import type { ContestWorkbenchStageKey } from '@/composables/useContestWorkbench'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import AWDChallengeConfigPanel from '@/components/platform/contest/AWDChallengeConfigPanel.vue'
import AWDOperationsPanel from '@/components/platform/contest/AWDOperationsPanel.vue'
import ContestAwdPreflightPanel from '@/components/platform/contest/ContestAwdPreflightPanel.vue'
import ContestChallengeOrchestrationPanel from '@/components/platform/contest/ContestChallengeOrchestrationPanel.vue'
import PlatformContestFormPanel from '@/components/platform/contest/PlatformContestFormPanel.vue'

defineProps<{
  loadError: string
  formDraft: ContestFormDraft | null
  contest: ContestDetailData | null
  activeStage: ContestWorkbenchStageKey
  saving: boolean
  statusOptions: Array<{ label: string; value: PlatformContestStatus }>
  fieldLocks: ContestFieldLocks
  loadingAwdStageData: boolean
  awdChallengeLinks: AdminContestChallengeViewData[]
  activeAwdChallengeId: string | null
  awdConfigFocusSource: 'pool' | 'preflight' | null
  canNavigatePreviousAwdChallenge: boolean
  canNavigateNextAwdChallenge: boolean
  awdPreflightLoadError: string
  awdReadiness: AWDReadinessData | null
}>()

const emit = defineEmits<{
  (event: 'go-back'): void
  (event: 'update:draft', value: ContestFormDraft): void
  (event: 'save', draft: ContestFormDraft): void
  (event: 'refresh-awd-workbench'): void
  (event: 'open:awd-config-from-pool', challenge: AdminContestChallengeViewData): void
  (event: 'open:awd-config-from-operations', challengeId: string): void
  (event: 'create:awd-challenge'): void
  (event: 'edit:awd-challenge', challenge: AdminContestChallengeViewData): void
  (event: 'previous:awd-challenge'): void
  (event: 'next:awd-challenge'): void
  (event: 'retry:preflight'): void
  (event: 'navigate:awd-challenge-from-preflight', challengeId: string): void
  (event: 'navigate:stage', stage: ContestWorkbenchStageKey): void
  (event: 'open:preflight-override'): void
}>()
</script>

<template>
  <div class="studio-canvas">
    <div class="studio-scroll-area">
      <AppEmpty
        v-if="loadError"
        title="竞赛详情加载失败"
        :description="loadError"
        icon="AlertTriangle"
      >
        <template #action>
          <button
            type="button"
            class="ui-btn ui-btn--ghost"
            @click="emit('go-back')"
          >
            返回竞赛目录
          </button>
        </template>
      </AppEmpty>

      <template v-else-if="formDraft && contest">
        <section class="workspace-directory-section contest-edit-section">
          <div
            v-if="activeStage === 'basics'"
            class="studio-pane studio-pane--full fade-in"
          >
            <div class="studio-form-canvas">
              <PlatformContestFormPanel
                :mode="'edit'"
                :draft="formDraft"
                :saving="saving"
                :status-options="statusOptions"
                :field-locks="fieldLocks"
                :show-cancel="false"
                @update:draft="emit('update:draft', $event)"
                @save="emit('save', $event)"
              />
            </div>
          </div>
        </section>

        <div
          v-if="activeStage === 'pool'"
          class="studio-pane fade-in"
        >
          <ContestChallengeOrchestrationPanel
            :contest-id="contest.id"
            :contest-mode="contest.mode"
            :challenge-links="contest.mode === 'awd' ? awdChallengeLinks : undefined"
            :loading-external="loadingAwdStageData"
            @open:awd-config="emit('open:awd-config-from-pool', $event)"
            @updated="emit('refresh-awd-workbench')"
          />
        </div>

        <div
          v-if="contest.mode === 'awd' && activeStage === 'awd-config'"
          class="studio-pane fade-in"
        >
          <template v-if="loadingAwdStageData && awdChallengeLinks.length === 0">
            <AppLoading>正在同步 AWD 配置...</AppLoading>
          </template>
          <AWDChallengeConfigPanel
            v-else
            :challenge-links="awdChallengeLinks"
            :active-challenge-id="activeAwdChallengeId"
            :focus-source="awdConfigFocusSource"
            :can-navigate-previous="canNavigatePreviousAwdChallenge"
            :can-navigate-next="canNavigateNextAwdChallenge"
            @create="emit('create:awd-challenge')"
            @edit="emit('edit:awd-challenge', $event)"
            @previous="emit('previous:awd-challenge')"
            @next="emit('next:awd-challenge')"
          />
        </div>

        <div
          v-if="contest.mode === 'awd' && activeStage === 'preflight'"
          class="studio-pane fade-in"
        >
          <AppEmpty
            v-if="awdPreflightLoadError"
            title="赛前检查暂时不可用"
            :description="awdPreflightLoadError"
            icon="AlertTriangle"
          >
            <template #action>
              <button
                type="button"
                class="ui-btn ui-btn--ghost"
                @click="emit('retry:preflight')"
              >
                重试加载
              </button>
            </template>
          </AppEmpty>
          <ContestAwdPreflightPanel
            v-else
            :readiness="awdReadiness"
            :loading="loadingAwdStageData"
            @navigate:challenge="emit('navigate:awd-challenge-from-preflight', $event)"
            @navigate:stage="emit('navigate:stage', $event)"
            @open:override="emit('open:preflight-override')"
          />
        </div>

        <div
          v-if="contest.mode === 'awd' && activeStage === 'operations'"
          class="studio-pane studio-pane--operations fade-in"
        >
          <header class="stage-pane-header">
            <h2 class="stage-pane-title">
              轮次态势
            </h2>
          </header>
          <AWDOperationsPanel
            :contests="[contest]"
            :selected-contest-id="contest.id"
            :hide-contest-selector="true"
            :hide-studio-link="true"
            :hide-operation-tabs="true"
            operation-panel="inspector"
            runtime-content="round-inspector"
            @open:awd-config="emit('open:awd-config-from-operations', $event)"
          />
        </div>

        <div
          v-if="contest.mode === 'awd' && activeStage === 'instances'"
          class="studio-pane studio-pane--operations fade-in"
        >
          <header class="stage-pane-header">
            <h2 class="stage-pane-title">
              实例编排
            </h2>
          </header>
          <AWDOperationsPanel
            :contests="[contest]"
            :selected-contest-id="contest.id"
            :hide-contest-selector="true"
            :hide-studio-link="true"
            :hide-operation-tabs="true"
            operation-panel="instances"
            runtime-content="instances"
            @open:awd-config="emit('open:awd-config-from-operations', $event)"
          />
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.studio-canvas {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  position: relative;
  background: var(--color-bg-surface);
}

.studio-scroll-area {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0;
  display: flex;
  flex-direction: column;
}

.studio-scroll-area::-webkit-scrollbar {
  width: 6px;
}

.studio-scroll-area::-webkit-scrollbar-thumb {
  background: color-mix(in srgb, var(--workspace-line-soft) 50%, transparent);
  border-radius: 10px;
}

.studio-pane {
  width: 100%;
  flex: 1 0 auto;
}

.studio-pane--operations {
  padding: 2rem;
}

.stage-pane-header {
  margin-bottom: var(--space-5);
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--color-border-subtle);
}

.stage-pane-title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-1-25);
  font-weight: 900;
}

.studio-form-canvas {
  width: 100%;
  max-width: none;
  border: none;
  background: transparent;
  padding: var(--space-6) var(--space-7);
  box-shadow: none;
}

.fade-in {
  animation: studioFadeIn 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes studioFadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
