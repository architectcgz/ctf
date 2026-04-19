import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import TeacherAWDReviewIndex from '../TeacherAWDReviewIndex.vue'
import teacherAwdReviewIndexSource from '../TeacherAWDReviewIndex.vue?raw'
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

describe('TeacherAWDReviewIndex', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    setActivePinia(createPinia())
    pushMock.mockReset()
    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())

    teacherApiMocks.listTeacherAWDReviews.mockResolvedValue([
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
    ])

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'teacher-1',
        username: 'teacher',
        role: 'teacher',
      },
      'token'
    )
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应加载 AWD 赛事目录并渲染进入复盘入口', async () => {
    const wrapper = mount(TeacherAWDReviewIndex)

    await flushPromises()

    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalled()
    expect(wrapper.text()).toContain('AWD复盘')
    expect(wrapper.text()).toContain('春季 AWD 联训')
    expect(wrapper.text()).toContain('进入复盘')
  })

  it('应在停止输入后自动筛选，不再依赖显式筛选按钮', async () => {
    const wrapper = mount(TeacherAWDReviewIndex)

    await flushPromises()

    const keywordInput = wrapper.find('input[placeholder="搜索赛事标题"]')
    const filterToggle = wrapper.find('.workspace-directory-toolbar__filter-toggle')

    expect(keywordInput.exists()).toBe(true)
    expect(filterToggle.exists()).toBe(true)
    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalledTimes(1)

    await filterToggle.trigger('click')
    const statusSelect = wrapper.find('select')
    expect(statusSelect.exists()).toBe(true)

    await statusSelect.setValue('ended')
    await keywordInput.setValue('期末')

    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalledTimes(1)

    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalledTimes(2)
    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenLastCalledWith({
      status: 'ended',
      keyword: '期末',
    })
    expect(wrapper.text()).not.toContain('应用筛选')
  })

  it('应使用共享目录工具栏与数据表，而不是继续渲染旧的自绘目录行', async () => {
    expect(teacherAwdReviewIndexSource).toContain(
      "from '@/components/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(teacherAwdReviewIndexSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(teacherAwdReviewIndexSource).toContain('<WorkspaceDirectoryToolbar')
    expect(teacherAwdReviewIndexSource).toContain('<WorkspaceDataTable')
    expect(teacherAwdReviewIndexSource).not.toContain('class="teacher-directory-row"')

    const wrapper = mount(TeacherAWDReviewIndex)

    await flushPromises()

    expect(wrapper.findComponent({ name: 'WorkspaceDirectoryToolbar' }).exists()).toBe(true)
    expect(wrapper.findComponent({ name: 'WorkspaceDataTable' }).exists()).toBe(true)
    expect(wrapper.text()).toContain('赛事编号')
    expect(wrapper.text()).toContain('赛事名称')
    expect(wrapper.text()).toContain('轮次')
    expect(wrapper.text()).toContain('队伍')
    expect(wrapper.text()).toContain('状态')
  })

  it('筛选区应保持平铺，不应继续在页面局部做成独立卡片壳', () => {
    expect(teacherAwdReviewIndexSource).toContain(
      'class="workspace-directory-section teacher-directory-section"'
    )
    expect(teacherAwdReviewIndexSource).toContain('class="list-heading"')
    expect(teacherAwdReviewIndexSource).not.toContain('teacher-controls-title')
    expect(teacherAwdReviewIndexSource).not.toContain('teacher-controls-copy')
    expect(teacherAwdReviewIndexSource).not.toMatch(
      /\.teacher-controls\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )
    expect(teacherAwdReviewIndexSource).not.toMatch(
      /\.teacher-controls\s*\{[\s\S]*background:\s*color-mix\(in srgb,\s*var\(--journal-surface-subtle\)\s*84%,\s*transparent\);/s
    )
    expect(teacherAwdReviewIndexSource).not.toMatch(
      /\.teacher-controls\s*\{[\s\S]*box-shadow:\s*0 10px 24px var\(--color-shadow-soft\);/s
    )
  })

  it('赛事概览条不应继续保留多余的底部分隔线', () => {
    expect(teacherAwdReviewIndexSource).toContain(
      'class="teacher-summary teacher-summary--flat metric-panel-default-surface"'
    )
    expect(teacherAwdReviewIndexSource).toMatch(
      /\.teacher-summary--flat\s*\{[\s\S]*border-bottom:\s*0;/s
    )
  })

  it('平台 AWD 复盘页头部应切到 workspace 语义，不再保留 teacher journal eyebrow', () => {
    expect(teacherAwdReviewIndexSource).toContain(
      '<header class="teacher-topbar workspace-tab-heading awd-review-index-header">'
    )
    expect(teacherAwdReviewIndexSource).toContain(
      '<div class="teacher-heading workspace-tab-heading__main">'
    )
    expect(teacherAwdReviewIndexSource).toContain(
      '<div class="workspace-overline awd-review-index-overline">AWD Review</div>'
    )
    expect(teacherAwdReviewIndexSource).toContain(
      '<h1 class="teacher-title workspace-page-title">AWD复盘</h1>'
    )
    expect(teacherAwdReviewIndexSource).toContain('<p class="teacher-copy workspace-page-copy">')
    expect(teacherAwdReviewIndexSource).toMatch(
      /\.awd-review-index-overline\s*\{[\s\S]*font-size:\s*var\(--journal-overline-font-size,\s*var\(--font-size-0-70\)\);[\s\S]*letter-spacing:\s*var\(--journal-overline-letter-spacing,\s*0\.2em\);[\s\S]*text-transform:\s*uppercase;[\s\S]*color:\s*var\(--journal-accent,\s*var\(--color-primary\)\);/s
    )
    expect(teacherAwdReviewIndexSource).not.toContain(
      '<div class="teacher-surface-eyebrow journal-eyebrow">AWD Review Workspace</div>'
    )
  })

  it('筛选区源码不应继续保留表单提交和应用筛选按钮', () => {
    expect(teacherAwdReviewIndexSource).not.toContain('@submit.prevent="loadContests"')
    expect(teacherAwdReviewIndexSource).not.toContain('应用筛选')
    expect(teacherAwdReviewIndexSource).not.toContain('赛事筛选')
    expect(teacherAwdReviewIndexSource).not.toContain(
      '支持按状态或关键字快速定位要进入的 AWD 赛事。'
    )
  })

  it('管理员打开 AWD 目录时应继续停留在后台教学运营路由', async () => {
    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
      },
      'token'
    )

    const wrapper = mount(TeacherAWDReviewIndex)

    await flushPromises()

    expect(wrapper.text()).toContain('平台概览')

    wrapper.findAll('button').find((button) => button.text().includes('平台概览'))?.trigger('click')
    wrapper.findAll('button').find((button) => button.text().includes('进入复盘'))?.trigger('click')
    await flushPromises()

    expect(pushMock).toHaveBeenCalledWith({ name: 'AdminDashboard' })
    expect(pushMock).toHaveBeenCalledWith({
      name: 'AdminAWDReviewDetail',
      params: { contestId: 'contest-1' },
    })
  })

  it('管理员打开 AWD 目录时应切换到管理员根壳，而不是继续使用教师根壳', async () => {
    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
      },
      'token'
    )

    const wrapper = mount(TeacherAWDReviewIndex)

    await flushPromises()

    expect(wrapper.classes()).toContain('workspace-shell')
    expect(wrapper.classes()).toContain('journal-shell-admin')
    expect(wrapper.classes()).toContain('journal-hero')
    expect(wrapper.classes()).not.toContain('teacher-management-shell')
    expect(wrapper.classes()).not.toContain('teacher-surface-hero')
    expect(wrapper.find('.ui-btn').exists()).toBe(true)
    expect(wrapper.find('.teacher-btn').exists()).toBe(false)
  })
})
