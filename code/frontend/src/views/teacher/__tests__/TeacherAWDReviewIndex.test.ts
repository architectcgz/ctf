import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

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

  it('筛选区源码不应继续保留表单提交和应用筛选按钮', () => {
    expect(teacherAwdReviewIndexSource).not.toContain('@submit.prevent="loadContests"')
    expect(teacherAwdReviewIndexSource).not.toContain('应用筛选')
    expect(teacherAwdReviewIndexSource).not.toContain('赛事筛选')
    expect(teacherAwdReviewIndexSource).not.toContain(
      '支持按状态或关键字快速定位要进入的 AWD 赛事。'
    )
  })

  it('加载骨架应通过语义类承接，不再直接写圆角和背景混色', () => {
    expect(teacherAwdReviewIndexSource).toContain('awd-review-loading-card')
    expect(teacherAwdReviewIndexSource).not.toContain('rounded-[22px]')
    expect(teacherAwdReviewIndexSource).not.toContain(
      'bg-[color-mix(in_srgb,var(--journal-surface-subtle)_92%,transparent)]'
    )
  })

  it('管理员打开 AWD 目录并进入复盘时应使用后台路由', async () => {
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

    wrapper.findAll('button').find((button) => button.text().includes('平台概览'))?.trigger('click')
    wrapper.findAll('button').find((button) => button.text().includes('进入复盘'))?.trigger('click')
    await flushPromises()

    expect(pushMock).toHaveBeenCalledWith({ name: 'AdminDashboard' })
    expect(pushMock).toHaveBeenCalledWith({
      name: 'AdminAWDReviewDetail',
      params: { contestId: 'contest-1' },
    })
  })
})
