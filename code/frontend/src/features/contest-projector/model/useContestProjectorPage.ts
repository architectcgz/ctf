import { computed, onMounted, onUnmounted, ref } from 'vue'

import { useToast } from '@/composables/useToast'
import type { ContestProjectorFocusPanel } from './projectorTypes'

import { useContestProjectorData } from './useContestProjectorData'
import { useContestProjectorDerived } from './useContestProjectorDerived'
import {
  formatProjectorTime,
  getContestStatusLabel,
  getRoundStatusLabel,
} from './projectorFormatters'

export function useContestProjectorPage() {
  const toast = useToast()
  const projectorStageRef = ref<HTMLElement | null>(null)
  const fullscreenActive = ref(false)
  const focusedPanel = ref<ContestProjectorFocusPanel | null>(null)

  const data = useContestProjectorData()
  const derived = useContestProjectorDerived({
    scoreboardRows: data.scoreboardRows,
    services: data.services,
    attacks: data.attacks,
    trafficSummary: data.trafficSummary,
  })

  const contestTitle = computed(
    () => data.selectedContest.value?.title ?? data.scoreboard.value?.contest.title ?? '未选择赛事'
  )
  const contestStatusLabel = computed(() =>
    getContestStatusLabel(data.selectedContest.value?.status ?? data.scoreboard.value?.contest.status)
  )
  const roundLabel = computed(
    () =>
      `R${data.selectedRound.value?.round_number ?? '--'} · ${getRoundStatusLabel(data.selectedRound.value?.status)}`
  )
  const topTeamName = computed(() => derived.topThreeRows.value[0]?.team_name ?? '--')
  const successfulAttackCount = computed(
    () => data.roundSummary.value?.metrics?.successful_attack_count ?? 0
  )
  const trafficRequestCount = computed(() => data.trafficSummary.value?.total_request_count ?? 0)
  const abnormalServiceCount = computed(
    () => derived.serviceStatusCounts.value.down + derived.serviceStatusCounts.value.compromised
  )

  function syncFullscreenState(): void {
    fullscreenActive.value = document.fullscreenElement === projectorStageRef.value
  }

  function focusPanel(panel: ContestProjectorFocusPanel): void {
    focusedPanel.value = panel
  }

  function closeFocusPanel(): void {
    focusedPanel.value = null
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
    } catch {
      toast.error('切换全屏失败')
    }
  }

  onMounted(() => {
    document.addEventListener('fullscreenchange', syncFullscreenState)
    void data.loadContests()
    data.startAutoRefresh()
  })

  onUnmounted(() => {
    document.removeEventListener('fullscreenchange', syncFullscreenState)
    data.stopAutoRefresh()
  })

  return {
    projectorStageRef,
    fullscreenActive,
    focusedPanel,
    contestTitle,
    contestStatusLabel,
    roundLabel,
    topTeamName,
    successfulAttackCount,
    trafficRequestCount,
    abnormalServiceCount,
    focusPanel,
    closeFocusPanel,
    toggleFullscreen,
    formatProjectorTime,
    ...data,
    ...derived,
  }
}
