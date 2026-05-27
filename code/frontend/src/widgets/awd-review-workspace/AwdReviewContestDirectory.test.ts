import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import type { AwdReviewContestItemData } from '@/api/contracts'
import AwdReviewContestDirectory from './AwdReviewContestDirectory.vue'

function createContests(): AwdReviewContestItemData[] {
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
    statusFilter: '' as '' | AwdReviewContestItemData['status'],
    keywordFilter: '',
    contestStatusLabel: () => '进行中',
  }
}

describe('AwdReviewContestDirectory', () => {
  it('应透传筛选更新、重试和进入复盘事件', async () => {
    const wrapper = mount(AwdReviewContestDirectory, {
      props: {
        ...createProps(),
        contests: [],
        hasContests: false,
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

  it('应透传目录行进入复盘事件', async () => {
    const wrapper = mount(AwdReviewContestDirectory, {
      props: createProps(),
    })

    await wrapper.find('button.teacher-directory-row').trigger('click')

    expect(wrapper.emitted('openContest')).toEqual([['contest-1']])
  })

  it('应透传分页切换事件', async () => {
    const wrapper = mount(AwdReviewContestDirectory, {
      props: createProps(),
    })

    const paginationButtons = wrapper.findAll('.page-pagination-controls__button')
    await paginationButtons[1].trigger('click')

    expect(wrapper.emitted('changePage')).toEqual([[2]])
  })
})
