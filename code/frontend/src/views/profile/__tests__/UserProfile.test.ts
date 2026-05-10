import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import UserProfile from '../UserProfile.vue'
import userProfileSource from '../UserProfile.vue?raw'
import { useAuthStore } from '@/stores/auth'

const authApiMocks = vi.hoisted(() => ({
  getProfile: vi.fn(),
}))

const assessmentApiMocks = vi.hoisted(() => ({
  exportPersonalReport: vi.fn(),
  getReportStatus: vi.fn(),
  downloadReport: vi.fn(),
}))

vi.mock('@/api/auth', () => authApiMocks)
vi.mock('@/api/assessment', () => assessmentApiMocks)

describe('UserProfile', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()

    authApiMocks.getProfile.mockReset()
    assessmentApiMocks.exportPersonalReport.mockReset()
    assessmentApiMocks.getReportStatus.mockReset()
    assessmentApiMocks.downloadReport.mockReset()

    authApiMocks.getProfile.mockResolvedValue({
      id: 'student-1',
      username: 'alice',
      role: 'student',
      class_name: 'Class A',
      name: 'Alice',
    })
    assessmentApiMocks.exportPersonalReport.mockResolvedValue({
      report_id: 'report-1',
      status: 'ready',
      expires_at: '2026-03-08T10:00:00Z',
    })
    assessmentApiMocks.getReportStatus.mockResolvedValue({
      report_id: 'report-1',
      status: 'ready',
      expires_at: '2026-03-08T10:00:00Z',
    })
    assessmentApiMocks.downloadReport.mockResolvedValue({
      blob: new Blob(['demo']),
      filename: 'report.pdf',
    })

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'student-1',
        username: 'alice',
        role: 'student',
        class_name: 'Class A',
        name: 'Alice',
      })

    vi.stubGlobal('URL', {
      createObjectURL: vi.fn(() => 'blob:demo'),
      revokeObjectURL: vi.fn(),
    })
  })

  it('应该展示个人资料并支持生成报告', async () => {
    const originalCreateElement = document.createElement.bind(document)
    const clickMock = vi.fn()
    const createElementSpy = vi
      .spyOn(document, 'createElement')
      .mockImplementation((tagName: string) => {
        if (tagName === 'a') {
          return {
            href: '',
            download: '',
            click: clickMock,
          } as unknown as HTMLAnchorElement
        }
        return originalCreateElement(tagName)
      })

    const wrapper = mount(UserProfile)
    await flushPromises()

    expect(wrapper.element.tagName).toBe('SECTION')
    expect(wrapper.classes()).toContain('journal-shell')
    expect(wrapper.classes()).toContain('journal-hero')
    expect(wrapper.classes()).toContain('min-h-full')
    expect(wrapper.text()).toContain('个人资料')
    expect(wrapper.text()).toContain('查看账号信息、个人报告与最近导出状态。')
    expect(wrapper.get('h1').classes()).toContain('workspace-page-title')
    expect(wrapper.find('.workspace-page-copy').exists()).toBe(true)
    expect(wrapper.find('.profile-topbar').exists()).toBe(true)
    expect(wrapper.find('.profile-summary').exists()).toBe(true)
    expect(wrapper.find('.profile-summary-title').text()).toContain('账号概况')
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('Class A')
    expect(wrapper.text()).toContain('生成个人报告')
    expect(wrapper.text()).toContain('报告状态')

    const createButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('生成个人报告'))
    expect(createButton).toBeTruthy()

    await createButton!.trigger('click')
    await flushPromises()

    expect(assessmentApiMocks.exportPersonalReport).toHaveBeenCalledWith({ format: 'pdf' })
    expect(wrapper.text()).toContain('report-1')

    const downloadButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('下载最近报告'))
    expect(downloadButton).toBeTruthy()

    await downloadButton!.trigger('click')
    await flushPromises()

    expect(assessmentApiMocks.downloadReport).toHaveBeenCalledWith('report-1')
    expect(clickMock).toHaveBeenCalledTimes(1)

    createElementSpy.mockRestore()
  })

  it('路由页应仅负责组合，不直接耦合个人资料请求与导出流程', () => {
    expect(userProfileSource).toContain('useUserProfilePage')
    expect(userProfileSource).not.toContain("from '@/api/auth'")
    expect(userProfileSource).not.toContain("from '@/api/assessment'")
    expect(userProfileSource).toContain('class="workspace-page-header profile-topbar"')
    expect(userProfileSource).toContain('class="profile-topbar-meta"')
    expect(userProfileSource).not.toContain('<PageHeader')
    expect(userProfileSource).toContain('class="profile-summary metric-panel-default-surface"')
    expect(userProfileSource).toContain('class="profile-summary-item progress-card metric-panel-card"')
    expect(userProfileSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(userProfileSource).toContain(
      'class="profile-summary-value progress-card-value metric-panel-value"'
    )
    expect(userProfileSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
    expect(userProfileSource).toContain('<component')
    expect(userProfileSource).not.toContain('class="profile-summary-icon"')
  })

  it('管理员不应该展示个人报告区块', async () => {
    authApiMocks.getProfile.mockResolvedValue({
      id: 'admin-1',
      username: 'admin',
      role: 'admin',
      name: 'Admin',
    })

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
        name: 'Admin',
      })

    const wrapper = mount(UserProfile)
    await flushPromises()

    expect(wrapper.text()).toContain('个人资料')
    expect(wrapper.text()).toContain('查看账号信息与当前账号状态。')
    expect(wrapper.get('h1').classes()).toContain('workspace-page-title')
    expect(wrapper.find('.workspace-page-copy').exists()).toBe(true)
    expect(wrapper.find('.profile-topbar').exists()).toBe(true)
    expect(wrapper.find('.profile-summary').exists()).toBe(true)
    expect(wrapper.text()).toContain('Admin')
    expect(wrapper.text()).not.toContain('生成个人报告')
    expect(wrapper.text()).not.toContain('个人报告')
    expect(wrapper.text()).not.toContain('导出格式')
    expect(wrapper.find('.profile-section--report').exists()).toBe(false)
    expect(wrapper.find('.profile-summary-grid').exists()).toBe(true)
    expect(wrapper.find('.profile-layout').classes()).toContain('profile-layout--single')
    expect(assessmentApiMocks.exportPersonalReport).not.toHaveBeenCalled()
  })

  it('下载最近报告失败时应展示错误提示而不是抛出异常', async () => {
    assessmentApiMocks.downloadReport.mockRejectedValue(new Error('下载失败'))

    const wrapper = mount(UserProfile)
    await flushPromises()

    const createButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('生成个人报告'))
    expect(createButton).toBeTruthy()

    await createButton!.trigger('click')
    await flushPromises()

    const downloadButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('下载最近报告'))
    expect(downloadButton).toBeTruthy()

    await expect(downloadButton!.trigger('click')).resolves.toBeUndefined()
    await flushPromises()

    expect(assessmentApiMocks.downloadReport).toHaveBeenCalledWith('report-1')
    expect(wrapper.text()).toContain('下载失败')
  })

  it('应该移除个人资料页级 shell 上遗留的 journal-eyebrow-text 修饰类', () => {
    expect(userProfileSource).toContain(
      'class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"'
    )
    expect(userProfileSource).not.toContain('journal-eyebrow-text')
  })

  it('应该把个人资料内容区的 soft eyebrow 收敛为局部 section kicker', () => {
    expect(userProfileSource).toContain('<div class="profile-section-kicker">Account</div>')
    expect(userProfileSource).toContain('<div class="profile-section-kicker">Report</div>')
    expect(userProfileSource).not.toContain('<div class="journal-eyebrow journal-eyebrow-soft">Account</div>')
    expect(userProfileSource).not.toContain('<div class="journal-eyebrow journal-eyebrow-soft">Report</div>')
  })

  it('应该把个人资料页残留的骨架圆角与内文字色收敛为语义类', () => {
    expect(userProfileSource).not.toContain('rounded-[24px]')
    expect(userProfileSource).not.toContain('text-[var(--journal-accent)]')
    expect(userProfileSource).not.toContain('text-[var(--journal-ink)]')
    expect(userProfileSource).not.toContain('text-[var(--journal-muted)]')
    expect(userProfileSource).toContain('profile-loading-card')
    expect(userProfileSource).toContain('profile-accent-icon')
    expect(userProfileSource).toContain('profile-format-title')
    expect(userProfileSource).toContain('profile-format-copy')
  })
})
