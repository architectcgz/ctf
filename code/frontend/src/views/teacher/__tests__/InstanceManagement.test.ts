import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { ElButton, ElTable, ElTableColumn } from 'element-plus'

import InstanceManagement from '../InstanceManagement.vue'
import instanceManagementSource from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue?raw'
import { useAuthStore } from '@/stores/auth'

const pushMock = vi.fn()

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
  getTeacherInstances: vi.fn(),
  destroyTeacherInstance: vi.fn(),
}))

const confirmMock = vi.hoisted(() => vi.fn())

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)
vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: confirmMock,
}))

describe('InstanceManagement', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()
    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())

    teacherApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 1 }])
    teacherApiMocks.getTeacherInstances.mockResolvedValue([
      {
        id: 'inst-1',
        student_id: 'stu-1',
        student_name: 'Alice',
        student_username: 'alice',
        student_no: 'S-1001',
        class_name: 'Class A',
        challenge_id: 'challenge-1',
        challenge_title: 'Web SQLi 101',
        status: 'running',
        access_url: 'http://127.0.0.1:30001',
        expires_at: '2026-03-09T10:30:00Z',
        remaining_time: 1200,
        extend_count: 1,
        max_extends: 3,
        created_at: '2026-03-09T09:30:00Z',
      },
    ])
    teacherApiMocks.destroyTeacherInstance.mockResolvedValue(undefined)
    confirmMock.mockReset()
    confirmMock.mockResolvedValue(true)

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'teacher-1',
        username: 'teacher',
        role: 'teacher',
        class_name: 'Class A',
      })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应该按教师所属班级加载实例', async () => {
    const wrapper = mount(InstanceManagement, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
      },
    })

    await flushPromises()

    expect(teacherApiMocks.getTeacherInstances).toHaveBeenCalledWith(
      {
        class_name: 'Class A',
        keyword: undefined,
        student_no: undefined,
      },
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      })
    )
    expect(wrapper.findAll('.progress-card.metric-panel-card')).toHaveLength(3)
    expect(wrapper.find('.teacher-directory-head').exists()).toBe(true)
    expect(wrapper.text()).toContain('Web SQLi 101')
    expect(wrapper.text()).toContain('@alice')
    expect(wrapper.find('.teacher-directory-row-title').attributes('title')).toBe('Alice')
    expect(wrapper.find('.teacher-directory-row-challenge').attributes('title')).toBe(
      'Web SQLi 101'
    )
    expect(wrapper.find('.teacher-directory-row-extends').text()).toBe('1 / 3')
    expect(wrapper.find('.teacher-directory-row-remaining').text()).toBe('00:20:00')
    expect(wrapper.find('.teacher-directory-row-url').text()).toContain('http://127.0.0.1:30001')
    expect(wrapper.text()).not.toContain('重置筛选')
    expect(wrapper.findAll('button').some((node) => node.text().includes('查询实例'))).toBe(false)
    expect(wrapper.text()).not.toContain('实例筛选')
    expect(wrapper.text()).not.toContain('支持按班级、用户名或学号关键字筛选，也可用学号精确筛选。')
  })

  it('应该支持输入后自动筛选并销毁实例', async () => {
    const wrapper = mount(InstanceManagement, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
      },
    })
    await flushPromises()

    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('ali')
    await inputs[1].setValue('S-1001')
    expect(teacherApiMocks.getTeacherInstances).toHaveBeenCalledTimes(1)
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(teacherApiMocks.getTeacherInstances).toHaveBeenLastCalledWith(
      {
        class_name: 'Class A',
        keyword: 'ali',
        student_no: 'S-1001',
      },
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      })
    )

    await wrapper.find('[data-instance-id="inst-1"]').trigger('click')
    await flushPromises()

    expect(confirmMock).toHaveBeenCalled()
    expect(teacherApiMocks.destroyTeacherInstance).toHaveBeenCalledWith('inst-1')
    expect(wrapper.text()).not.toContain('Web SQLi 101')
  })

  it('取消危险确认后不应继续销毁实例', async () => {
    confirmMock.mockResolvedValue(false)

    const wrapper = mount(InstanceManagement, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
      },
    })
    await flushPromises()

    await wrapper.find('[data-instance-id="inst-1"]').trigger('click')
    await flushPromises()

    expect(confirmMock).toHaveBeenCalled()
    expect(teacherApiMocks.destroyTeacherInstance).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Web SQLi 101')
  })

  it('管理员从实例管理返回概览时应回到后台概览页', async () => {
    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
        class_name: 'Class A',
      })

    const wrapper = mount(InstanceManagement, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
      },
    })
    await flushPromises()

    wrapper.findComponent({ name: 'TeacherInstanceManagementPage' }).vm.$emit('openDashboard')

    expect(pushMock).toHaveBeenCalledWith({ name: 'PlatformOverview' })
  })

  it('应该支持实例目录分页切换', async () => {
    teacherApiMocks.getTeacherInstances.mockResolvedValue(
      Array.from({ length: 21 }, (_, index) => ({
        id: `inst-${index + 1}`,
        student_id: `stu-${index + 1}`,
        student_name: `Student ${index + 1}`,
        student_username: `student-${index + 1}`,
        student_no: `S-${String(index + 1).padStart(4, '0')}`,
        class_name: 'Class A',
        challenge_id: `challenge-${index + 1}`,
        challenge_title: `Challenge ${index + 1}`,
        status: 'running',
        access_url: `http://127.0.0.1:30${String(index + 1).padStart(3, '0')}`,
        expires_at: '2026-03-09T10:30:00Z',
        remaining_time: 1200,
        extend_count: 1,
        max_extends: 3,
        created_at: '2026-03-09T09:30:00Z',
      }))
    )

    const wrapper = mount(InstanceManagement, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
      },
    })
    await flushPromises()

    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(20)
    expect(wrapper.find('.teacher-directory-pagination').text()).toContain('共 21 条实例')
    expect(wrapper.find('.teacher-directory-pagination').text()).toContain('1 / 2')
    expect(wrapper.text()).toContain('Challenge 20')
    expect(wrapper.text()).not.toContain('Challenge 21')

    const paginationButtons = wrapper.findAll('.page-pagination-controls__button')
    await paginationButtons[1].trigger('click')
    await flushPromises()

    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(1)
    expect(wrapper.find('.teacher-directory-pagination').text()).toContain('2 / 2')
    expect(wrapper.text()).toContain('Challenge 21')
    expect(wrapper.text()).not.toContain('Challenge 20')
  })

  it('应该为教师实例列表长文本保留省略样式与完整提示', () => {
    expect(instanceManagementSource).toContain(
      'class="workspace-directory-section teacher-directory-section"'
    )
    expect(instanceManagementSource).toContain('class="list-heading"')
    expect(instanceManagementSource).toContain(
      'class="teacher-summary metric-panel-default-surface"'
    )
    expect(instanceManagementSource).toContain(
      'class="teacher-summary-grid progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(instanceManagementSource).toContain('class="progress-card metric-panel-card"')
    expect(instanceManagementSource).not.toContain(
      'class="teacher-summary-item progress-card metric-panel-card"'
    )
    expect(instanceManagementSource).toContain('class="progress-card-label metric-panel-label"')
    expect(instanceManagementSource).not.toContain(
      'class="teacher-summary-label progress-card-label metric-panel-label"'
    )
    expect(instanceManagementSource).toContain('class="progress-card-value metric-panel-value"')
    expect(instanceManagementSource).not.toContain(
      'class="teacher-summary-value progress-card-value metric-panel-value"'
    )
    expect(instanceManagementSource).toContain('class="progress-card-hint metric-panel-helper"')
    expect(instanceManagementSource).not.toContain(
      'class="teacher-summary-helper progress-card-hint metric-panel-helper"'
    )
    expect(instanceManagementSource).not.toContain('teacher-controls-title')
    expect(instanceManagementSource).not.toContain('teacher-controls-copy')
    expect(instanceManagementSource).toContain('用户关键字')
    expect(instanceManagementSource).toContain('按用户名或学号搜索')
    expect(instanceManagementSource).toContain('<span>学生</span>')
    expect(instanceManagementSource).toContain('<span>题目</span>')
    expect(instanceManagementSource).not.toContain('<span>学生 / 题目</span>')
    expect(instanceManagementSource).toContain('<span>创建时间</span>')
    expect(instanceManagementSource).toContain('<span>到期时间</span>')
    expect(instanceManagementSource).toContain('<span>延期</span>')
    expect(instanceManagementSource).toContain('<span>剩余时间</span>')
    expect(instanceManagementSource).toContain('<span>访问地址</span>')
    expect(instanceManagementSource).toContain('class="teacher-directory-row-challenge"')
    expect(instanceManagementSource).toContain('class="teacher-directory-row-created"')
    expect(instanceManagementSource).toContain('class="teacher-directory-row-expires-at"')
    expect(instanceManagementSource).toContain('class="teacher-directory-row-extends"')
    expect(instanceManagementSource).toContain('class="teacher-directory-row-remaining"')
    expect(instanceManagementSource).toContain('class="teacher-directory-row-url"')
    expect(instanceManagementSource).toContain(
      'class="teacher-directory-pagination workspace-directory-pagination"'
    )
    expect(instanceManagementSource).not.toContain('重置筛选')
    expect(instanceManagementSource).not.toContain('查询实例')
    expect(instanceManagementSource).not.toContain('创建于 {{ formatDateTime(item.created_at) }}')
    expect(instanceManagementSource).not.toContain('到期 {{ formatDateTime(item.expires_at) }}')
    expect(instanceManagementSource).not.toContain(
      '延期 {{ item.extend_count }} / {{ item.max_extends }}'
    )
    expect(instanceManagementSource).not.toContain(
      '剩余 {{ formatRemainingTime(item.remaining_time) }}'
    )
    expect(instanceManagementSource).not.toContain(
      'class="teacher-directory-chip teacher-directory-chip-muted"'
    )
    expect(instanceManagementSource).not.toContain('class="teacher-directory-row-metrics"')
    expect(instanceManagementSource).toContain('teacher-directory-state-chip--success')
    expect(instanceManagementSource).not.toContain('border-[var(--color-success)]')
    expect(instanceManagementSource).not.toContain('bg-[var(--color-success)]')
    expect(instanceManagementSource).not.toContain('text-[var(--color-success)]')
    expect(instanceManagementSource).not.toContain('border-[var(--color-primary)]')
    expect(instanceManagementSource).not.toContain('bg-[var(--color-primary)]')
    expect(instanceManagementSource).not.toContain('text-[var(--color-primary)]')
    expect(instanceManagementSource).toMatch(
      /class="teacher-directory-row-title"[\s\S]*:title="item\.student_name \|\| item\.student_username"/s
    )
    expect(instanceManagementSource).toMatch(
      /class="teacher-directory-row-challenge"[\s\S]*:title="item\.challenge_title"/s
    )
    expect(instanceManagementSource).toMatch(/class="teacher-directory-row-copy"[\s\S]*:title="/s)
    expect(instanceManagementSource).toMatch(
      /\.teacher-directory-row-title\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
    expect(instanceManagementSource).toMatch(
      /\.teacher-directory-row-challenge\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
    expect(instanceManagementSource).toMatch(
      /\.teacher-directory-row-copy\s*\{[^}]*display:\s*-webkit-box;[^}]*-webkit-line-clamp:\s*2;[^}]*overflow:\s*hidden;/s
    )
  })
})
