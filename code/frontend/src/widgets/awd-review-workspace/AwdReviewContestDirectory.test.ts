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
    buildContestRoute: (contestId: string) =>
      ({
        name: 'TeacherAWDReviewDetail',
        params: { contestId },
      }) as const,
    statusFilter: '' as '' | AwdReviewContestItemData['status'],
    keywordFilter: '',
    contestStatusLabel: () => '进行中',
  }
}

describe('AwdReviewContestDirectory', () => {
  it('应透传筛选更新、重试事件', async () => {
    const wrapper = mount(AwdReviewContestDirectory, {
      props: {
        ...createProps(),
        contests: [],
        hasContests: false,
        error: '加载失败',
      },
      global: {
        stubs: {
          AppRouteLink: {
            template: '<a><slot /></a>',
          },
        },
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

  it('应把目录行改成显式 route target link', () => {
    const wrapper = mount(AwdReviewContestDirectory, {
      props: createProps(),
      global: {
        stubs: {
          AppRouteLink: {
            name: 'AppRouteLink',
            props: ['to'],
            template:
              '<a class="teacher-directory-row" :data-route-name="to.name" :data-contest-id="to.params?.contestId"><slot /></a>',
          },
        },
      },
    })

    const rowLink = wrapper.get('a.teacher-directory-row')
    expect(rowLink.attributes('data-route-name')).toBe('TeacherAWDReviewDetail')
    expect(rowLink.attributes('data-contest-id')).toBe('contest-1')
  })

  it('应透传分页切换事件', async () => {
    const wrapper = mount(AwdReviewContestDirectory, {
      props: createProps(),
      global: {
        stubs: {
          AppRouteLink: {
            template: '<a><slot /></a>',
          },
        },
      },
    })

    const paginationButtons = wrapper.findAll('.page-pagination-controls__button')
    await paginationButtons[1].trigger('click')

    expect(wrapper.emitted('changePage')).toEqual([[2]])
  })
})
