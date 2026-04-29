<script setup lang="ts">
import type {
  AdminContestChallengeViewData,
  AWDReadinessData,
  ContestDetailData,
} from '@/api/contracts'
import type {
  ContestFieldLocks,
  ContestFormDraft,
  PlatformContestStatus,
} from '@/composables/usePlatformContests'
import type { ContestWorkbenchStageKey } from '@/composables/useContestWorkbench'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import AWDChallengeConfigPanel from '@/components/platform/contest/AWDChallengeConfigPanel.vue'
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
  awdChallengePoolCreateRequestKey: number
  awdPreflightLoadError: string
  awdReadiness: AWDReadinessData | null
}>()

const emit = defineEmits<{
  (event: 'go-back'): void
  (event: 'update:draft', value: ContestFormDraft): void
  (event: 'save', draft: ContestFormDraft): void
  (event: 'refresh-awd-workbench'): void
  (event: 'edit:awd-challenge', challenge: AdminContestChallengeViewData): void
  (event: 'retry:preflight'): void
  (event: 'navigate:awd-challenge-from-preflight', challengeId: string): void
  (event: 'navigate:stage', stage: ContestWorkbenchStageKey): void
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
        <Transition
          name="studio-stage"
          mode="out-in"
        >
          <section
            v-if="activeStage === 'basics'"
            key="basics"
            class="workspace-directory-section contest-edit-section studio-stage-panel"
          >
            <div class="studio-pane studio-pane--full">
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
            v-else-if="activeStage === 'pool'"
            key="pool"
            class="studio-pane studio-stage-panel"
          >
            <ContestChallengeOrchestrationPanel
              :contest-id="contest.id"
              :contest-mode="contest.mode"
              :challenge-links="contest.mode === 'awd' ? awdChallengeLinks : undefined"
              :loading-external="loadingAwdStageData"
              :create-dialog-request-key="awdChallengePoolCreateRequestKey"
              @updated="emit('refresh-awd-workbench')"
            />
          </div>

          <div
            v-else-if="contest.mode === 'awd' && activeStage === 'awd-config'"
            key="awd-config"
            class="studio-pane studio-stage-panel"
          >
            <template v-if="loadingAwdStageData && awdChallengeLinks.length === 0">
              <AppLoading>正在同步 AWD 配置...</AppLoading>
            </template>
            <AWDChallengeConfigPanel
              v-else
              :challenge-links="awdChallengeLinks"
              @edit="emit('edit:awd-challenge', $event)"
            />
          </div>

          <div
            v-else-if="contest.mode === 'awd' && activeStage === 'preflight'"
            key="preflight"
            class="studio-pane studio-stage-panel"
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
            />
          </div>

        </Transition>
      </template>
    </div>
  </div>
</template>

<style scoped>
.studio-canvas {
  --contest-studio-canvas-surface: color-mix(
    in srgb,
    var(--journal-surface, var(--color-bg-surface)) 88%,
    var(--color-bg-base)
  );
  --contest-studio-canvas-surface-soft: color-mix(
    in srgb,
    var(--journal-surface-subtle, var(--color-bg-elevated)) 84%,
    var(--color-bg-base)
  );
  --contest-studio-pane-background:
    radial-gradient(
      circle at top left,
      color-mix(in srgb, var(--workspace-brand, var(--color-primary)) 8%, transparent),
      transparent 34%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--contest-studio-canvas-surface) 96%, transparent),
      var(--contest-studio-canvas-surface-soft)
    );

  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  position: relative;
  background: var(--contest-studio-pane-background);
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
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 42%, transparent),
      transparent 44%
    ),
    color-mix(in srgb, var(--contest-studio-canvas-surface) 74%, transparent);
  border-top: 1px solid color-mix(in srgb, var(--workspace-line-soft) 72%, transparent);
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--color-bg-surface) 84%, transparent);
}

.studio-stage-panel {
  transform-origin: top center;
  will-change: opacity, transform;
}

.studio-form-canvas {
  width: 100%;
  max-width: none;
  border: none;
  background: transparent;
  padding: var(--space-6) var(--space-7);
  box-shadow: none;
}

.studio-stage-enter-active {
  transition:
    opacity 280ms cubic-bezier(0.22, 1, 0.36, 1),
    transform 280ms cubic-bezier(0.22, 1, 0.36, 1);
}

.studio-stage-leave-active {
  transition:
    opacity 180ms cubic-bezier(0.25, 1, 0.5, 1),
    transform 180ms cubic-bezier(0.25, 1, 0.5, 1);
}

.studio-stage-enter-from {
  opacity: 0;
  transform: translate3d(0, var(--space-3), 0);
}

.studio-stage-leave-to {
  opacity: 0;
  transform: translate3d(0, calc(var(--space-2) * -1), 0);
}

@media (prefers-reduced-motion: reduce) {
  .studio-stage-panel {
    will-change: auto;
  }

  .studio-stage-enter-active,
  .studio-stage-leave-active {
    transition: none;
  }

  .studio-stage-enter-from,
  .studio-stage-leave-to {
    opacity: 1;
    transform: none;
  }
}
</style>
