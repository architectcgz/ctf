import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import UserProfile from '../UserProfile.vue'
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
      },
      'token'
    )

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
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('Class A')
    expect(wrapper.text()).toContain('生成个人报告')

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
})
