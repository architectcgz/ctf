<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'

import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import ContestProjectorEvents from '@/components/platform/contest/projector/ContestProjectorEvents.vue'
import ContestProjectorHero from '@/components/platform/contest/projector/ContestProjectorHero.vue'
import ContestProjectorLeaderboard from '@/components/platform/contest/projector/ContestProjectorLeaderboard.vue'
import ContestProjectorServiceMatrix from '@/components/platform/contest/projector/ContestProjectorServiceMatrix.vue'
import ContestProjectorToolbar from '@/components/platform/contest/projector/ContestProjectorToolbar.vue'
import ContestProjectorTraffic from '@/components/platform/contest/projector/ContestProjectorTraffic.vue'
import {
  formatProjectorTime,
  getContestStatusLabel,
  getRoundStatusLabel,
} from '@/components/platform/contest/projector/contestProjectorFormatters'
import { useContestProjectorData } from '@/composables/useContestProjectorData'
import { useContestProjectorDerived } from '@/composables/useContestProjectorDerived'
import { useToast } from '@/composables/useToast'

const toast = useToast()
const projectorStageRef = ref<HTMLElement | null>(null)
const fullscreenActive = ref(false)

const {
  scoreboard,
  services,
  attacks,
  roundSummary,
  trafficSummary,
  loadingContests,
  loadingScoreboard,
  loadError,
  lastUpdatedLabel,
  projectorContests,
  selectedContest,
  selectedContestId,
  scoreboardRows,
  selectedRound,
  loadContests,
  selectContest,
  startAutoRefresh,
  stopAutoRefresh,
} = useContestProjectorData()

const {
  topThreeRows,
  leaderboardRows,
  firstBlood,
  latestAttackEvents,
  serviceStatusCounts,
  serviceHealthRate,
  serviceMatrixRows,
  attackLeaders,
  trafficTrendBars,
  hotVictims,
} = useContestProjectorDerived({
  scoreboardRows,
  services,
  attacks,
  trafficSummary,
})

const contestTitle = computed(() => selectedContest.value?.title ?? scoreboard.value?.contest.title ?? '未选择赛事')
const contestStatusLabel = computed(() =>
  getContestStatusLabel(selectedContest.value?.status ?? scoreboard.value?.contest.status)
)
const roundLabel = computed(
  () => `R${selectedRound.value?.round_number ?? '--'} · ${getRoundStatusLabel(selectedRound.value?.status)}`
)
const topTeamName = computed(() => topThreeRows.value[0]?.team_name ?? '--')
const successfulAttackCount = computed(() => roundSummary.value?.metrics?.successful_attack_count ?? 0)
const trafficRequestCount = computed(() => trafficSummary.value?.total_request_count ?? 0)
const abnormalServiceCount = computed(
  () => serviceStatusCounts.value.down + serviceStatusCounts.value.compromised
)

function syncFullscreenState(): void {
  fullscreenActive.value = document.fullscreenElement === projectorStageRef.value
}

async function toggleFullscreen(): Promise<void> {
  try {
    if (fullscreenActive.value) {
      await document.exitFullscreen()
      return
    }

    const target = projectorStageRef.value
    if (!target?.requestFullscreen) {
      toast.error('当前浏览器不支持全屏展示')
      return
    }
    await target.requestFullscreen()
  } catch (error) {
    toast.error('切换全屏失败')
  }
}

onMounted(() => {
  document.addEventListener('fullscreenchange', syncFullscreenState)
  void loadContests()
  startAutoRefresh()
})

onUnmounted(() => {
  document.removeEventListener('fullscreenchange', syncFullscreenState)
  stopAutoRefresh()
})
</script>

<template>
  <section class="contest-projector-shell workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col">
    <div class="workspace-grid">
      <main class="content-pane contest-projector-content">
        <section
          ref="projectorStageRef"
          class="projector-stage workspace-directory-section"
        >
          <ContestProjectorToolbar
            :contests="projectorContests"
            :selected-contest-id="selectedContestId"
            :last-updated-label="lastUpdatedLabel"
            :fullscreen-active="fullscreenActive"
            :loading-contests="loadingContests"
            :loading-scoreboard="loadingScoreboard"
            @refresh="void loadContests()"
            @toggle-fullscreen="void toggleFullscreen()"
            @select-contest="(contestId) => void selectContest(contestId)"
          />

          <AppLoading v-if="loadingContests">
            正在同步大屏赛事...
          </AppLoading>

          <AppEmpty
            v-else-if="loadError"
            title="大屏展示暂时不可用"
            :description="loadError"
            icon="AlertTriangle"
            class="py-20"
          >
            <template #action>
              <button
                type="button"
                class="ui-btn ui-btn--ghost"
                @click="void loadContests()"
              >
                重试加载
              </button>
            </template>
          </AppEmpty>

          <AppEmpty
            v-else-if="projectorContests.length === 0"
            title="暂无可展示的 AWD 赛事"
            description="大屏展示只面向进行中、冻结或已结束的 AWD 赛事。"
            icon="FileChartColumnIncreasing"
            class="py-20"
          />

          <section
            v-else
            class="scoreboard-projector"
          >
            <ContestProjectorHero
              :title="contestTitle"
              :contest-status-label="contestStatusLabel"
              :round-label="roundLabel"
              :team-count="scoreboardRows.length"
              :leader-name="topTeamName"
              :service-health-rate="serviceHealthRate"
              :successful-attack-count="successfulAttackCount"
              :traffic-request-count="trafficRequestCount"
              :abnormal-service-count="abnormalServiceCount"
            />

            <div
              v-if="loadingScoreboard"
              class="scoreboard-loading"
            >
              <AppLoading>正在刷新榜单...</AppLoading>
            </div>

            <div
              v-else
              class="projector-board"
            >
              <div class="arena-grid">
                <ContestProjectorLeaderboard
                  :top-three-rows="topThreeRows"
                  :leaderboard-rows="leaderboardRows"
                  :scoreboard-rows-length="scoreboardRows.length"
                />

                <section class="battlefield-panel">
                  <ContestProjectorServiceMatrix
                    :rows="serviceMatrixRows"
                    :up-count="serviceStatusCounts.up"
                    :down-count="serviceStatusCounts.down"
                    :compromised-count="serviceStatusCounts.compromised"
                  />

                  <ContestProjectorTraffic
                    :summary="trafficSummary"
                    :trend-bars="trafficTrendBars"
                    :hot-victims="hotVictims"
                  />
                </section>

                <ContestProjectorEvents
                  :first-blood="firstBlood"
                  :attack-leaders="attackLeaders"
                  :latest-attack-events="latestAttackEvents"
                />
              </div>

              <div class="projector-footer">
                <span>结束 {{ formatProjectorTime(selectedContest?.ends_at ?? scoreboard?.contest.ends_at) }}</span>
                <span>最新流量 {{ formatProjectorTime(trafficSummary?.latest_event_at) }}</span>
                <span>轮次创建 {{ formatProjectorTime(selectedRound?.created_at) }}</span>
              </div>
            </div>
          </section>
        </section>
      </main>
    </div>
  </section>
</template>

<style scoped>
.contest-projector-shell {
  --workspace-shell-bg: var(--journal-surface);
  --workspace-shell-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
}

.contest-projector-content,
.projector-stage,
.scoreboard-projector,
.projector-board,
.battlefield-panel {
  display: flex;
  flex-direction: column;
}

.contest-projector-content {
  padding: 0;
}

.projector-stage {
  --workspace-directory-section-padding: var(--space-4) var(--space-5);
  gap: var(--space-4);
  background: transparent;
}

.projector-stage:fullscreen {
  min-height: 100vh;
  overflow: auto;
  background: var(--journal-surface);
  padding: var(--space-5);
}

.scoreboard-projector,
.projector-board,
.battlefield-panel {
  gap: var(--space-4);
}

.scoreboard-projector {
  position: relative;
  min-height: 38rem;
}

.scoreboard-loading {
  display: flex;
  min-height: 20rem;
  align-items: center;
  justify-content: center;
}

.arena-grid {
  display: grid;
  grid-template-columns: minmax(18rem, 1.05fr) minmax(26rem, 1.65fr) minmax(18rem, 1fr);
  gap: var(--space-4);
  align-items: stretch;
}

.projector-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  border-top: 1px solid var(--color-border-subtle);
  padding-top: var(--space-3);
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 800;
}

@media (max-width: 1280px) {
  .arena-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 900px) {
  .projector-stage {
    --workspace-directory-section-padding: var(--space-4);
  }

  .projector-footer {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
