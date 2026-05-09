import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import PlatformAWDReviewIndex from '../AWDReviewIndex.vue'
import platformAwdReviewIndexSource from '../AWDReviewIndex.vue?raw'
import awdReviewHeroPanelSource from '@/components/platform/awd-review/AwdReviewHeroPanel.vue?raw'
import awdReviewDirectoryPanelSource from '@/components/platform/awd-review/AwdReviewDirectoryPanel.vue?raw'
import { useAuthStore } from '@/stores/auth'

const pushMock = vi.fn()

const teacherApiMocks = vi.hoisted(() => ({
  listTeacherAWDReviews: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

const combinedSource = [
  platformAwdReviewIndexSource,
  awdReviewHeroPanelSource,
  awdReviewDirectoryPanelSource,
].join('\n')

describe('PlatformAWDReviewIndex', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    setActivePinia(createPinia())
    useAuthStore().user = { id: 'admin-1', role: 'admin' } as never
    pushMock.mockReset()
    teacherApiMocks.listTeacherAWDReviews.mockReset()
    teacherApiMocks.listTeacherAWDReviews.mockImplementation(async (params) => {
      const contests = [
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
        {
          id: 'contest-2',
          title: '期末 AWD 复盘',
          mode: 'awd',
          status: 'ended',
          current_round: 8,
          round_count: 8,
          team_count: 10,
          export_ready: true,
        },
      ]

      const filtered = contests.filter((item) => {
        const matchesStatus = !params?.status || item.status === params.status
        const matchesKeyword = !params?.keyword || item.title.includes(params.keyword)
        return matchesStatus && matchesKeyword
      })

      return {
        list: filtered,
        total: filtered.length,
        page: params?.page ?? 1,
        page_size: params?.page_size ?? 20,
        summary: {
          running_count: filtered.filter((item) => item.status === 'running').length,
          export_ready_count: filtered.filter((item) => item.export_ready).length,
        },
      }
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应使用平台工作台目录壳层而不是教师目录模板', async () => {
    expect(combinedSource).toContain(
      "from '@/components/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(combinedSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(platformAwdReviewIndexSource).toContain(
      'class="workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero admin-awd-review-shell flex min-h-full flex-1 flex-col"'
    )
    expect(awdReviewHeroPanelSource).toContain(
      'class="admin-summary-grid admin-awd-review-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(combinedSource).toContain(
      'class="workspace-directory-section admin-awd-review-directory"'
    )
    expect(combinedSource).toContain(
      'class="workspace-directory-list admin-awd-review-table"'
    )
    expect(combinedSource).not.toContain('teacher-management-shell')
    expect(combinedSource).not.toContain('teacher-directory-row')

    const wrapper = mount(PlatformAWDReviewIndex)
    await flushPromises()

    expect(wrapper.text()).toContain('AWD复盘')
    expect(wrapper.text()).toContain('春季 AWD 联训')
    expect(wrapper.text()).toContain('赛事目录')
    expect(wrapper.text()).toContain('进入复盘')
  })

  it('应支持自动筛选并跳转到平台复盘详情', async () => {
    const wrapper = mount(PlatformAWDReviewIndex)
    await flushPromises()

    const searchInput = wrapper.get('input[placeholder="搜索赛事标题"]')
    await searchInput.setValue('期末')

    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalledTimes(1)

    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalledTimes(2)
    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenLastCalledWith({
      status: undefined,
      keyword: '期末',
      page: 1,
      page_size: 20,
    }, {
      signal: expect.any(AbortSignal),
    })

    await wrapper
      .findAll('button')
      .find((node) => node.text().includes('进入复盘'))
      ?.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'PlatformAwdReviewDetail',
      params: { contestId: 'contest-2' },
    })
  })

  it('顶部返回按钮应回到平台概览', async () => {
    const wrapper = mount(PlatformAWDReviewIndex)
    await flushPromises()

    const overviewButton = wrapper.get('button.ui-btn--ghost')
    expect(overviewButton.text()).toContain('返回平台概览')

    await overviewButton.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({ name: 'PlatformOverview' })
  })

  it('路由壳页应通过 feature 动作处理返回，而不是模板内直接 push', () => {
    expect(platformAwdReviewIndexSource).toContain('openPlatformOverview')
    expect(platformAwdReviewIndexSource).not.toContain("router.push({ name: 'PlatformOverview' })")
  })
})
