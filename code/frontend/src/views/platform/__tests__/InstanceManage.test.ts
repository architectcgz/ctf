import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import PlatformInstanceManagement from '../InstanceManage.vue'
import adminInstanceManageSource from '../InstanceManage.vue?raw'
import instanceManageHeroPanelSource from '@/components/platform/instance/InstanceManageHeroPanel.vue?raw'
import instanceManageWorkspacePanelSource from '@/components/platform/instance/InstanceManageWorkspacePanel.vue?raw'

const pushMock = vi.fn()

const teacherApiMocks = vi.hoisted(() => ({
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
    pushMock.mockReset()
    teacherApiMocks.getTeacherInstances.mockReset()
    teacherApiMocks.destroyTeacherInstance.mockReset()
    confirmMock.mockReset()

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
      {
        id: 'inst-2',
        student_id: 'stu-2',
        student_name: 'Bob',
        student_username: 'bob',
        student_no: 'S-1002',
        class_name: 'Class B',
        challenge_id: 'challenge-2',
        challenge_title: 'Pwn Stack 201',
        status: 'expired',
        access_url: '',
        expires_at: '2026-03-09T09:00:00Z',
        remaining_time: 0,
        extend_count: 0,
        max_extends: 3,
        created_at: '2026-03-09T08:30:00Z',
      },
    ])
    teacherApiMocks.destroyTeacherInstance.mockResolvedValue(undefined)
    confirmMock.mockResolvedValue(true)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应保留当前后台实例页样式并复用 teacher 实例接口', async () => {
    expect(adminInstanceManageSource).toContain("from '@/api/teacher'")
    expect(adminInstanceManageSource).toContain("from '@/composables/useDestructiveConfirm'")
    expect(adminInstanceManageSource).not.toContain("from '@/api/admin'")
    expect(adminInstanceManageSource).not.toContain("from '@/composables/useAdminDestructiveConfirm'")
    expect(adminInstanceManageSource).toContain(
      "import InstanceManageWorkspacePanel from '@/components/platform/instance/InstanceManageWorkspacePanel.vue'"
    )
    expect(adminInstanceManageSource).toContain(
      "import InstanceManageHeroPanel from '@/components/platform/instance/InstanceManageHeroPanel.vue'"
    )
    expect(adminInstanceManageSource).toContain('<InstanceManageHeroPanel')
    expect(adminInstanceManageSource).toContain('<InstanceManageWorkspacePanel')
    expect(instanceManageHeroPanelSource).toContain('返回概览')
    expect(instanceManageHeroPanelSource).toContain('刷新列表')
    expect(instanceManageHeroPanelSource).toContain('class="ui-btn ui-btn--primary"')
    expect(instanceManageHeroPanelSource).toContain(
      'class="admin-summary-grid admin-instance-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(instanceManageWorkspacePanelSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(instanceManageWorkspacePanelSource).toContain(
      "from '@/components/common/WorkspaceDirectoryPagination.vue'"
    )
    expect(instanceManageWorkspacePanelSource).toContain('<WorkspaceDataTable')
    expect(instanceManageWorkspacePanelSource).toContain('<WorkspaceDirectoryPagination')
    expect(instanceManageWorkspacePanelSource).toContain('class="instance-status-pill"')
    expect(adminInstanceManageSource).not.toContain('bg-green-100 text-green-700')
    expect(adminInstanceManageSource).not.toContain('bg-slate-100 text-slate-600')

    const wrapper = mount(PlatformInstanceManagement)
    await flushPromises()

    expect(teacherApiMocks.getTeacherInstances).toHaveBeenCalledWith({
      class_name: undefined,
      keyword: undefined,
      student_no: undefined,
    })
    expect(wrapper.text()).toContain('实例管理')
    expect(wrapper.text()).toContain('Web SQLi 101')
    expect(wrapper.text()).toContain('Alice')
    expect(wrapper.text()).toContain('http://127.0.0.1:30001')
    expect(wrapper.text()).toContain('运行中')
    expect(wrapper.text()).toContain('已过期')
    expect(wrapper.text()).toContain('销毁')
  })

  it('应支持销毁实例并更新列表', async () => {
    const wrapper = mount(PlatformInstanceManagement)
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((node) => node.text().includes('销毁'))
      ?.trigger('click')
    await flushPromises()

    expect(confirmMock).toHaveBeenCalled()
    expect(teacherApiMocks.destroyTeacherInstance).toHaveBeenCalledWith('inst-1')
    expect(wrapper.text()).not.toContain('Web SQLi 101')
    expect(wrapper.text()).toContain('Pwn Stack 201')
  })

  it('应支持返回概览页', async () => {
    const wrapper = mount(PlatformInstanceManagement)
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((node) => node.text().includes('返回概览'))
      ?.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({ name: 'PlatformOverview' })
  })
})
