import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import TeacherAWDReviewWorkspace from './TeacherAWDReviewWorkspace.vue'
import type { TeacherAWDReviewArchiveData } from '@/api/contracts'

function createReview(): TeacherAWDReviewArchiveData {
  return {
    generated_at: '2026-01-01T00:00:00Z',
    scope: {
      snapshot_type: 'live',
      requested_by: 1,
      requested_id: 'contest-1',
    },
    contest: {
      id: 'contest-1',
      title: '春季 AWD 联训',
      mode: 'awd',
      status: 'running',
      current_round: 1,
      round_count: 4,
      team_count: 6,
      export_ready: true,
    },
    overview: {
      round_count: 4,
      team_count: 6,
      service_count: 12,
      attack_count: 8,
      traffic_count: 20,
    },
    rounds: [],
  }
}

function createProps() {
  return {
    polling: false,
    loading: false,
    error: null as string | null,
    review: createReview(),
    exporting: null as 'archive' | 'report' | null,
    activeContestTitle: '春季 AWD 联训',
    activeSummaryTitle: '整场总览',
    summaryStats: {
      roundCount: 4,
      teamCount: 6,
      serviceCount: 12,
      attackCount: 8,
      trafficCount: 20,
    },
    timelineRounds: [],
    selectedRoundNumber: undefined as number | undefined,
    selectedRound: undefined,
    selectedTeam: null,
    selectedTeamServices: [],
    selectedTeamAttacks: [],
    selectedTeamTraffic: [],
    canExportReport: true,
    contestStatusLabel: () => '进行中',
    formatServiceRef: (id?: string) => `Service #${id || '--'}`,
  }
}

describe('TeacherAWDReviewWorkspace', () => {
  it('应转发顶部动作事件', async () => {
    const wrapper = mount(TeacherAWDReviewWorkspace, {
      props: createProps(),
      global: {
        stubs: {
          TeacherAWDReviewRoundSelector: { template: '<div />' },
          TeacherAWDReviewAnalysisSection: { template: '<div />' },
          TeacherAWDReviewEvidenceGrid: { template: '<div />' },
          TeacherAWDReviewTeamDrawer: { template: '<div />' },
        },
      },
    })

    const [backButton, archiveButton, reportButton] = wrapper.findAll('button')

    await backButton.trigger('click')
    await archiveButton.trigger('click')
    await reportButton.trigger('click')

    expect(wrapper.emitted('openIndex')).toBeTruthy()
    expect(wrapper.emitted('exportArchive')).toBeTruthy()
    expect(wrapper.emitted('exportReport')).toBeTruthy()
  })

  it('应转发轮次切换、队伍关闭与重试加载事件', async () => {
    const wrapper = mount(TeacherAWDReviewWorkspace, {
      props: {
        ...createProps(),
        review: null,
        error: '加载失败',
      },
      global: {
        stubs: {
          TeacherAWDReviewRoundSelector: {
            template: '<button data-testid="round-selector" @click="$emit(\'set-round\', 2)">round</button>',
          },
          TeacherAWDReviewTeamDrawer: {
            template: '<button data-testid="team-drawer" @click="$emit(\'close\')">drawer</button>',
          },
        },
      },
    })

    await wrapper.get('[data-testid="round-selector"]').trigger('click')
    await wrapper.get('[data-testid="team-drawer"]').trigger('click')
    const reloadButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('重新加载'))

    expect(reloadButton).toBeTruthy()
    await reloadButton!.trigger('click')

    expect(wrapper.emitted('setRound')).toEqual([[2]])
    expect(wrapper.emitted('closeTeam')).toBeTruthy()
    expect(wrapper.emitted('loadReview')).toBeTruthy()
  })
})
