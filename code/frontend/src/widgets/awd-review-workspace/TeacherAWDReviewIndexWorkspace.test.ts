import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import TeacherAWDReviewIndexWorkspace from './TeacherAWDReviewIndexWorkspace.vue'
import type { TeacherAWDReviewContestItemData } from '@/api/contracts'

function createContests(): TeacherAWDReviewContestItemData[] {
  return [
    {
      id: 'contest-1',
      title: '春季 AWD 联训',
      mode: 'awd',
      status: 'running',
      current_round: 2,
      round_count: 6,
      team_count: 8,
      export_ready: false,
    },
  ]
}

function createProps() {
  return {
    loading: false,
    error: null as string | null,
    contests: createContests(),
    total: 21,
    page: 1,
    totalPages: 2,
    hasContests: true,
    statusOptions: [
      { value: '', label: '全部状态' },
      { value: 'running', label: '进行中' },
    ] as const,
    contestSummary: {
      totalCount: 1,
      runningCount: 1,
      exportReadyCount: 0,
    },
    statusFilter: '' as '' | TeacherAWDReviewContestItemData['status'],
    keywordFilter: '',
    contestStatusLabel: () => '进行中',
  }
}

describe('TeacherAWDReviewIndexWorkspace', () => {
  it('应转发顶部动作和进入复盘事件', async () => {
    const wrapper = mount(TeacherAWDReviewIndexWorkspace, {
      props: createProps(),
    })

    const [dashboardButton, refreshButton] = wrapper.findAll('button')

    await dashboardButton.trigger('click')
    await refreshButton.trigger('click')
    await wrapper.find('button.teacher-directory-row').trigger('click')

    expect(wrapper.emitted('openDashboard')).toBeTruthy()
    expect(wrapper.emitted('refresh')).toBeTruthy()
    expect(wrapper.emitted('openContest')).toEqual([['contest-1']])
  })

  it('应转发分页切换事件', async () => {
    const wrapper = mount(TeacherAWDReviewIndexWorkspace, {
      props: createProps(),
    })

    const paginationButtons = wrapper.findAll('.page-pagination-controls__button')
    await paginationButtons[1].trigger('click')

    expect(wrapper.emitted('changePage')).toEqual([[2]])
  })

  it('应转发筛选输入与错误态重试事件', async () => {
    const wrapper = mount(TeacherAWDReviewIndexWorkspace, {
      props: {
        ...createProps(),
        hasContests: false,
        contests: [],
        error: '加载失败',
      },
    })

    await wrapper.find('select').setValue('running')
    await wrapper.find('input[placeholder="搜索赛事标题"]').setValue('期末')
    const reloadButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('重新加载'))

    expect(reloadButton).toBeTruthy()
    await reloadButton!.trigger('click')

    expect(wrapper.emitted('updateStatusFilter')).toEqual([['running']])
    expect(wrapper.emitted('updateKeywordFilter')).toEqual([['期末']])
    expect(wrapper.emitted('reload')).toBeTruthy()
  })
})
