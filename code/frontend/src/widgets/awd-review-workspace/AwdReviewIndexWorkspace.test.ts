import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import type { AwdReviewContestItemData } from '@/api/contracts'
import AwdReviewIndexWorkspace from './AwdReviewIndexWorkspace.vue'

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
    dashboardRoute: {
      name: 'TeacherDashboard',
    } as const,
    buildContestRoute: (contestId: string) =>
      ({
        name: 'TeacherAWDReviewDetail',
        params: { contestId },
      }) as const,
    contestSummary: {
      totalCount: 1,
      runningCount: 1,
      exportReadyCount: 0,
    },
    statusFilter: '' as '' | AwdReviewContestItemData['status'],
    keywordFilter: '',
    contestStatusLabel: () => '进行中',
  }
}

describe('AwdReviewIndexWorkspace', () => {
  it('应透传头部与目录 route target，并转发刷新事件', async () => {
    const wrapper = mount(AwdReviewIndexWorkspace, {
      props: createProps(),
      global: {
        stubs: {
          AppRouteLink: {
            name: 'AppRouteLink',
            props: ['to'],
            template:
              '<a :data-route-name="to.name" :data-contest-id="to.params?.contestId"><slot /></a>',
          },
        },
      },
    })

    const refreshButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('刷新目录'))
    expect(refreshButton).toBeTruthy()
    await refreshButton!.trigger('click')

    const routeLinks = wrapper.findAll('a')
    expect(routeLinks[0].attributes('data-route-name')).toBe('TeacherDashboard')
    expect(routeLinks[1].attributes('data-route-name')).toBe('TeacherAWDReviewDetail')
    expect(routeLinks[1].attributes('data-contest-id')).toBe('contest-1')
    expect(wrapper.emitted('refresh')).toBeTruthy()
  })

  it('应转发分页切换事件', async () => {
    const wrapper = mount(AwdReviewIndexWorkspace, {
      props: createProps(),
      global: {
        stubs: {
          AppRouteLink: {
            template: '<a><slot /></a>',
          },
        },
      },
    })

    const nextPageButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('下一页'))
    expect(nextPageButton).toBeTruthy()
    await nextPageButton!.trigger('click')

    expect(wrapper.emitted('changePage')).toEqual([[2]])
  })

  it('应转发筛选输入与错误态重试事件', async () => {
    const wrapper = mount(AwdReviewIndexWorkspace, {
      props: {
        ...createProps(),
        hasContests: false,
        contests: [],
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
})
