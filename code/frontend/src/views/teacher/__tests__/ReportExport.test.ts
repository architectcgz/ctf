import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import ReportExport from '../ReportExport.vue'
import { useAuthStore } from '@/stores/auth'

const {
  downloadReportMock,
  exportClassReportMock,
  getClassReviewMock,
  getClassStudentsMock,
  getClassSummaryMock,
  getClassTrendMock,
  getReportStatusMock,
} = vi.hoisted(() => ({
  exportClassReportMock: vi.fn(),
  downloadReportMock: vi.fn(),
  getClassStudentsMock: vi.fn(),
  getClassReviewMock: vi.fn(),
  getClassSummaryMock: vi.fn(),
  getClassTrendMock: vi.fn(),
  getReportStatusMock: vi.fn(),
}))

vi.mock('@/api/teacher', () => ({
  exportClassReport: exportClassReportMock,
  getClassStudents: getClassStudentsMock,
  getClassReview: getClassReviewMock,
  getClassSummary: getClassSummaryMock,
  getClassTrend: getClassTrendMock,
}))

vi.mock('@/api/assessment', () => ({
  downloadReport: downloadReportMock,
  getReportStatus: getReportStatusMock,
}))

describe('ReportExport', () => {
  let pinia: ReturnType<typeof createPinia>

  const dialogStub = {
    props: ['modelValue'],
    template: '<div v-if="modelValue"><slot name="header" /><slot /></div>',
  }

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    exportClassReportMock.mockReset()
    downloadReportMock.mockReset()
    getClassStudentsMock.mockReset()
    getClassReviewMock.mockReset()
    getClassSummaryMock.mockReset()
    getClassTrendMock.mockReset()
    getReportStatusMock.mockReset()
    localStorage.clear()

    getClassStudentsMock.mockResolvedValue([
      {
        id: 'stu-1',
        username: 'alice',
        solved_count: 4,
        total_score: 320,
        weak_dimension: 'crypto',
      },
      {
        id: 'stu-2',
        username: 'bob',
        solved_count: 2,
        total_score: 180,
        weak_dimension: 'web',
      },
    ])
    getClassReviewMock.mockResolvedValue({
      class_name: 'Class A',
      items: [
        {
          key: 'activity',
          title: '班级活跃度需要补强',
          detail: '建议优先跟进低活跃学生。',
          accent: 'warning',
        },
      ],
    })
    getClassSummaryMock.mockResolvedValue({
      class_name: 'Class A',
      student_count: 2,
      average_solved: 3,
      active_student_count: 2,
      active_rate: 100,
      recent_event_count: 8,
    })
    getClassTrendMock.mockResolvedValue({
      class_name: 'Class A',
      points: [
        { date: '2026-03-05', active_student_count: 1, event_count: 2, solve_count: 1 },
        { date: '2026-03-06', active_student_count: 2, event_count: 4, solve_count: 3 },
      ],
    })

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
        stubs: {
          ElDialog: dialogStub,
          LineChart: true,
        },
      },
    })

    await flushPromises()

    const classInput = wrapper.find('input[type="text"]')
    expect((classInput.element as HTMLInputElement).value).toBe('Class A')

    const previewButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('打开报告预览'))
    expect(previewButton).toBeTruthy()

    await previewButton!.trigger('click')
    await flushPromises()

    expect(getClassStudentsMock).toHaveBeenCalledWith('Class A')
    expect(getClassReviewMock).toHaveBeenCalledWith('Class A')
    expect(wrapper.text()).toContain('Live Preview')
    expect(wrapper.text()).toContain('班级活跃度需要补强')

    const exportButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('创建导出任务'))
    expect(exportButton).toBeTruthy()

    await exportButton!.trigger('click')
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
        stubs: {
          ElDialog: dialogStub,
          LineChart: true,
        },
      },
    })

    await flushPromises()

    const exportButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('创建导出任务'))
    expect(exportButton).toBeTruthy()

    await exportButton!.trigger('click')
    await flushPromises()

    const actionButtons = wrapper.findAll('button')
    await actionButtons[actionButtons.length - 1].trigger('click')
    await flushPromises()

    expect(downloadReportMock).toHaveBeenCalledWith('102')
  })
})
