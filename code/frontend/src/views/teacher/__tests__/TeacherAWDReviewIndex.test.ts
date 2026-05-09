import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import TeacherAWDReviewIndex from '../TeacherAWDReviewIndex.vue'
import teacherAwdReviewIndexSource from '../TeacherAWDReviewIndex.vue?raw'
import teacherAwdReviewIndexWorkspaceSource from '@/widgets/teacher-awd-review/TeacherAWDReviewIndexWorkspace.vue?raw'
import teacherAwdReviewContestDirectorySource from '@/widgets/teacher-awd-review/TeacherAWDReviewContestDirectory.vue?raw'
import teacherAwdReviewDirectorySectionSource from '@/widgets/teacher-awd-review/TeacherAWDReviewDirectorySection.vue?raw'

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

describe('TeacherAWDReviewIndex', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    setActivePinia(createPinia())
    pushMock.mockReset()
    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())

    teacherApiMocks.listTeacherAWDReviews.mockResolvedValue({
      list: [
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
      ],
      total: 2,
      page: 1,
      page_size: 20,
      summary: {
        running_count: 1,
        export_ready_count: 1,
      },
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应加载 AWD 赛事目录并渲染进入复盘入口', async () => {
    const wrapper = mount(TeacherAWDReviewIndex)

    await flushPromises()

    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalledWith({
      status: undefined,
      keyword: undefined,
      page: 1,
      page_size: 20,
    }, {
      signal: expect.any(AbortSignal),
    })
    expect(wrapper.text()).toContain('AWD复盘')
    expect(wrapper.text()).toContain('春季 AWD 联训')
    expect(wrapper.text()).toContain('进入复盘')
  })

  it('页面应通过 feature model 获取筛选与摘要状态，不再直接耦合 teacher api', () => {
    expect(teacherAwdReviewIndexSource).toContain("useTeacherAwdReviewIndex } from '@/features/teacher-awd-review'")
    expect(teacherAwdReviewIndexSource).toContain(
      "import { TeacherAWDReviewIndexWorkspace } from '@/widgets/teacher-awd-review'"
    )
    expect(teacherAwdReviewIndexSource).not.toContain("from '@/api/teacher'")
    expect(teacherAwdReviewIndexSource).not.toContain('const statusOptions = [')
    expect(teacherAwdReviewIndexSource).not.toContain('function contestStatusLabel')
    expect(teacherAwdReviewIndexSource).not.toContain('router.push({ name: \'TeacherDashboard\' })')
    expect(teacherAwdReviewIndexSource).not.toContain('contests.filter((item) => item.status ===')
    expect(teacherAwdReviewIndexSource).not.toContain('contests.filter((item) => item.export_ready)')
  })

  it('应在停止输入后自动筛选，不再依赖显式筛选按钮', async () => {
    const wrapper = mount(TeacherAWDReviewIndex)

    await flushPromises()

    const statusSelect = wrapper.find('select')
    const keywordInput = wrapper.find('input[placeholder="搜索赛事标题"]')

    expect(statusSelect.exists()).toBe(true)
    expect(keywordInput.exists()).toBe(true)
    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalledTimes(1)

    await statusSelect.setValue('ended')
    await keywordInput.setValue('期末')

    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalledTimes(1)

    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalledTimes(2)
    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenLastCalledWith({
      status: 'ended',
      keyword: '期末',
      page: 1,
      page_size: 20,
    }, {
      signal: expect.any(AbortSignal),
    })
    expect(wrapper.text()).not.toContain('应用筛选')
  })

  it('头部概览按钮应返回教学概览', async () => {
    const wrapper = mount(TeacherAWDReviewIndex)

    await flushPromises()

    const overviewButton = wrapper.get('button.teacher-btn--ghost')

    expect(overviewButton.text()).toContain('教学概览')

    await overviewButton.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({ name: 'TeacherDashboard' })
  })

  it('筛选区应保持平铺，不应继续在页面局部做成独立卡片壳', () => {
    expect(teacherAwdReviewIndexWorkspaceSource).toContain('<TeacherAWDReviewContestDirectory')
    expect(teacherAwdReviewContestDirectorySource).toContain('<TeacherAWDReviewIndexFilters')
    expect(teacherAwdReviewContestDirectorySource).toContain('<TeacherAWDReviewDirectorySection')
    expect(teacherAwdReviewDirectorySectionSource).toContain(
      'class="workspace-directory-section teacher-directory-section"'
    )
    expect(teacherAwdReviewDirectorySectionSource).toContain('class="list-heading"')
    expect(teacherAwdReviewContestDirectorySource).not.toContain('teacher-controls-title')
    expect(teacherAwdReviewContestDirectorySource).not.toContain('teacher-controls-copy')
    expect(teacherAwdReviewDirectorySectionSource).not.toMatch(
      /\.teacher-controls\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )
    expect(teacherAwdReviewDirectorySectionSource).not.toMatch(
      /\.teacher-controls\s*\{[\s\S]*background:\s*color-mix\(in srgb,\s*var\(--journal-surface-subtle\)\s*84%,\s*transparent\);/s
    )
    expect(teacherAwdReviewDirectorySectionSource).not.toMatch(
      /\.teacher-controls\s*\{[\s\S]*box-shadow:\s*0 10px 24px var\(--color-shadow-soft\);/s
    )
  })

  it('赛事概览条不应继续保留多余的底部分隔线', () => {
    expect(teacherAwdReviewIndexWorkspaceSource).toContain('<TeacherAWDReviewSummaryPanel')
  })

  it('平台 AWD 复盘页头部应切到 workspace 语义，不再保留 teacher journal eyebrow', () => {
    expect(teacherAwdReviewIndexWorkspaceSource).toContain('<TeacherAWDReviewWorkspaceHeader')
    expect(teacherAwdReviewIndexWorkspaceSource).toContain('TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY')
    expect(teacherAwdReviewIndexWorkspaceSource).toContain(
      ':overline="TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY.overline"'
    )
    expect(teacherAwdReviewIndexWorkspaceSource).toContain(
      ':title="TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY.title"'
    )
    expect(teacherAwdReviewIndexWorkspaceSource).toContain('header-class="awd-review-index-header"')
    expect(teacherAwdReviewIndexWorkspaceSource).toContain(
      'overline-class="awd-review-index-overline"'
    )
    expect(teacherAwdReviewIndexWorkspaceSource).toMatch(
      /\.awd-review-index-overline\s*\{[\s\S]*font-size:\s*var\(--journal-overline-font-size,\s*var\(--font-size-0-70\)\);[\s\S]*letter-spacing:\s*var\(--journal-overline-letter-spacing,\s*0\.2em\);[\s\S]*text-transform:\s*uppercase;[\s\S]*color:\s*var\(--journal-accent,\s*var\(--color-primary\)\);/s
    )
    expect(teacherAwdReviewIndexWorkspaceSource).not.toContain(
      '<div class="teacher-surface-eyebrow journal-eyebrow">AWD Review Workspace</div>'
    )
  })

  it('筛选区源码不应继续保留表单提交和应用筛选按钮', () => {
    expect(teacherAwdReviewIndexWorkspaceSource).not.toContain('@submit.prevent="loadContests"')
    expect(teacherAwdReviewIndexWorkspaceSource).not.toContain('应用筛选')
    expect(teacherAwdReviewIndexWorkspaceSource).not.toContain('赛事筛选')
    expect(teacherAwdReviewIndexWorkspaceSource).not.toContain(
      '支持按状态或关键字快速定位要进入的 AWD 赛事。'
    )
  })
})
