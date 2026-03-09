import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { ElButton, ElTable, ElTableColumn } from 'element-plus'

import ClassManagement from '../ClassManagement.vue'

const pushMock = vi.fn()

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

describe('ClassManagement', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()
    teacherApiMocks.getClasses.mockReset()
    teacherApiMocks.getClasses.mockResolvedValue([
      { name: 'Class A', student_count: 2 },
      { name: 'Class B', student_count: 3 },
    ])
  })

  it('应该展示班级列表并支持进入班级学生页', async () => {
    const wrapper = mount(ClassManagement, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('班级管理')
    expect(wrapper.text()).toContain('Class A')
    expect(wrapper.text()).toContain('Class B')

    await wrapper.findAll('button').find((node) => node.text().includes('进入'))?.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'TeacherClassStudents',
      params: { className: 'Class A' },
    })
  })
})
