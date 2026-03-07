import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import ReportExport from '../ReportExport.vue'
import { useAuthStore } from '@/stores/auth'

const { downloadReportMock, exportClassReportMock, getReportStatusMock } = vi.hoisted(() => ({
  exportClassReportMock: vi.fn(),
  downloadReportMock: vi.fn(),
  getReportStatusMock: vi.fn(),
}))

vi.mock('@/api/teacher', () => ({
  exportClassReport: exportClassReportMock,
}))

vi.mock('@/api/assessment', () => ({
  downloadReport: downloadReportMock,
  getReportStatus: getReportStatusMock,
}))

describe('ReportExport', () => {
  let pinia: ReturnType<typeof createPinia>

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    exportClassReportMock.mockReset()
    downloadReportMock.mockReset()
    getReportStatusMock.mockReset()
    localStorage.clear()

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: '1',
        username: 'teacher-a',
        role: 'teacher',
        class_name: 'Class A',
        name: 'Teacher A',
      },
      'token'
    )

    vi.stubGlobal('URL', {
      createObjectURL: vi.fn(() => 'blob:report'),
      revokeObjectURL: vi.fn(),
    })

    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})
  })

  it('应该使用默认班级创建导出任务', async () => {
    exportClassReportMock.mockResolvedValue({
      report_id: '101',
      status: 'processing',
    })
    getReportStatusMock.mockResolvedValue({
      report_id: '101',
      status: 'ready',
      expires_at: '2026-03-07T12:00:00Z',
    })

    const wrapper = mount(ReportExport, {
      global: {
        plugins: [pinia],
      },
    })

    const classInput = wrapper.find('input[type="text"]')
    expect((classInput.element as HTMLInputElement).value).toBe('Class A')

    await wrapper.find('button').trigger('click')
    await flushPromises()

    expect(exportClassReportMock).toHaveBeenCalledWith({
      class_name: 'Class A',
      format: 'pdf',
    })
    expect(wrapper.text()).toContain('101')
  })

  it('应该在报告就绪后触发下载', async () => {
    exportClassReportMock.mockResolvedValue({
      report_id: '102',
      status: 'ready',
      expires_at: '2026-03-07T12:00:00Z',
    })
    downloadReportMock.mockResolvedValue({
      blob: new Blob(['ok']),
      filename: 'class-a-report.pdf',
    })

    const wrapper = mount(ReportExport, {
      global: {
        plugins: [pinia],
      },
    })

    await wrapper.find('button').trigger('click')
    await flushPromises()

    const actionButtons = wrapper.findAll('button')
    await actionButtons[actionButtons.length - 1].trigger('click')
    await flushPromises()

    expect(downloadReportMock).toHaveBeenCalledWith('102')
  })
})
