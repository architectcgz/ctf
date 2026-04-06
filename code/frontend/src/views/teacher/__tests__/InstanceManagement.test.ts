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

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

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

    vi.stubGlobal('confirm', vi.fn(() => true))

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'teacher-1',
      username: 'teacher',
      role: 'teacher',
      class_name: 'Class A',
    }, 'token')
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

    expect(teacherApiMocks.getTeacherInstances).toHaveBeenCalledWith({
      class_name: 'Class A',
      keyword: undefined,
      student_no: undefined,
    })
    expect(wrapper.findAll('.teacher-summary-item')).toHaveLength(3)
    expect(wrapper.find('.teacher-directory-head').exists()).toBe(true)
    expect(wrapper.text()).toContain('Web SQLi 101')
    expect(wrapper.text()).toContain('@alice')
    expect(wrapper.find('.teacher-directory-row-title').attributes('title')).toBe('Alice')
    expect(wrapper.find('.teacher-directory-row-points').attributes('title')).toBe('Web SQLi 101')
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

    expect(teacherApiMocks.getTeacherInstances).toHaveBeenLastCalledWith({
      class_name: 'Class A',
      keyword: 'ali',
      student_no: 'S-1001',
    })

    await wrapper.find('[data-instance-id="inst-1"]').trigger('click')
    await flushPromises()

    expect(teacherApiMocks.destroyTeacherInstance).toHaveBeenCalledWith('inst-1')
    expect(wrapper.text()).not.toContain('Web SQLi 101')
  })

  it('应该为教师实例列表长文本保留省略样式与完整提示', () => {
    expect(instanceManagementSource).toMatch(/class="teacher-directory-row-title"[\s\S]*:title="item\.student_name \|\| item\.student_username"/s)
    expect(instanceManagementSource).toMatch(/class="teacher-directory-row-points"[\s\S]*:title="item\.challenge_title"/s)
    expect(instanceManagementSource).toMatch(/class="teacher-directory-row-copy"[\s\S]*:title="/s)
    expect(instanceManagementSource).toMatch(/\.teacher-directory-row-title\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s)
    expect(instanceManagementSource).toMatch(/\.teacher-directory-row-points\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s)
    expect(instanceManagementSource).toMatch(/\.teacher-directory-row-copy\s*\{[^}]*display:\s*-webkit-box;[^}]*-webkit-line-clamp:\s*2;[^}]*overflow:\s*hidden;/s)
  })
})
