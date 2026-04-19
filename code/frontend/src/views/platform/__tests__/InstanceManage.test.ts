import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import PlatformInstanceManagement from '../InstanceManage.vue'
import adminInstanceManageSource from '../InstanceManage.vue?raw'
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

describe('PlatformInstanceManagement', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()
    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())

    teacherApiMocks.getClasses.mockResolvedValue([
      { name: 'Class A', student_count: 1 },
      { name: 'Class B', student_count: 1 },
    ])
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
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
      },
      'token'
    )
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应使用后台工作台目录组件而不是教师端实例目录壳层', async () => {
    expect(adminInstanceManageSource).toContain(
      "from '@/components/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(adminInstanceManageSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(adminInstanceManageSource).toContain(
      "from '@/components/common/WorkspaceDirectoryPagination.vue'"
    )
    expect(adminInstanceManageSource).toContain(
      'class="workspace-shell admin-instance-manage-shell"'
    )
    expect(adminInstanceManageSource).toContain('<WorkspaceDirectoryToolbar')
    expect(adminInstanceManageSource).toContain('<WorkspaceDataTable')
    expect(adminInstanceManageSource).toContain('<WorkspaceDirectoryPagination')
    expect(adminInstanceManageSource).toMatch(
      /\.admin-instance-manage-directory\s*\{[\s\S]*display:\s*grid;[\s\S]*gap:\s*var\(--space-4\);/s
    )
    expect(adminInstanceManageSource).toMatch(
      /\.admin-instance-manage-directory :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:\s*0;/s
    )
    expect(adminInstanceManageSource).not.toContain('teacher-management-shell')
    expect(adminInstanceManageSource).not.toContain('teacher-directory-row')

    const wrapper = mount(PlatformInstanceManagement)
    await flushPromises()

    expect(wrapper.text()).toContain('实例管理')
    expect(wrapper.text()).toContain('Web SQLi 101')
    expect(wrapper.text()).toContain('Alice')
    expect(wrapper.text()).toContain('运行中')
    expect(wrapper.text()).toContain('访问地址')
    expect(wrapper.text()).toContain('销毁实例')
  })

  it('应支持筛选并销毁实例', async () => {
    const wrapper = mount(PlatformInstanceManagement)
    await flushPromises()

    const searchInput = wrapper.get('input[placeholder="检索学生、用户名或题目名称..."]')
    await searchInput.setValue('ali')
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(teacherApiMocks.getTeacherInstances).toHaveBeenLastCalledWith(
      {
        class_name: undefined,
        keyword: 'ali',
        student_no: undefined,
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

  it('卸载页面前的延迟筛选不应在离开后继续发请求', async () => {
    const wrapper = mount(PlatformInstanceManagement)
    await flushPromises()

    const searchInput = wrapper.get('input[placeholder="检索学生、用户名或题目名称..."]')
    await searchInput.setValue('alice')

    wrapper.unmount()

    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(teacherApiMocks.getTeacherInstances).toHaveBeenCalledTimes(1)
  })
})
