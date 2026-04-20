import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import TeacherAWDReviewDetail from '../TeacherAWDReviewDetail.vue'
import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'
import { useAuthStore } from '@/stores/auth'

const replaceMock = vi.fn()
const routeMock = {
  params: {
    contestId: 'contest-1',
  },
}

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeMock,
    useRouter: () => ({ replace: replaceMock, push: vi.fn() }),
  }
})

describe('TeacherAWDReviewDetail', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    useAuthStore().user = { id: 'teacher-1', role: 'teacher' } as never
    routeMock.params.contestId = 'contest-1'
    replaceMock.mockReset()
  })

  it('教师旧详情入口应跳到新的 overview 子页', async () => {
    mount(TeacherAWDReviewDetail)

    await flushPromises()

    expect(replaceMock).toHaveBeenCalledWith({
      name: 'TeacherAwdOverview',
      params: { contestId: 'contest-1' },
    })
  })

  it('管理员旧详情入口应跳到新的 replay 子页', async () => {
    useAuthStore().user = { id: 'admin-1', role: 'admin' } as never

    mount(TeacherAWDReviewDetail)

    await flushPromises()

    expect(replaceMock).toHaveBeenCalledWith({
      name: 'AdminAwdReplay',
      params: { id: 'contest-1' },
    })
  })

  it('旧详情页源码应只保留进入新工作台的过渡壳', () => {
    expect(awdReviewDetailSource).toContain('正在进入 AWD 复盘工作台')
    expect(awdReviewDetailSource).not.toContain('导出教师报告')
    expect(awdReviewDetailSource).not.toContain('轮次目录')
  })
})
