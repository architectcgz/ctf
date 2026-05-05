import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'

import { useContestAwdDefenseWorkbenchPage } from './useContestAwdDefenseWorkbenchPage'

const routeState = vi.hoisted(() => ({
  value: null as { params: { id: string; serviceId: string } } | null,
}))

const contestApiMocks = vi.hoisted(() => ({
  getContestAWDWorkspace: vi.fn(),
  getContestChallenges: vi.fn(),
  requestContestAWDDefenseDirectory: vi.fn(),
  requestContestAWDDefenseFile: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const { reactive } = await vi.importActual<typeof import('vue')>('vue')
  routeState.value = reactive({
    params: {
      id: '7',
      serviceId: '7009',
    },
  })
  return {
    useRoute: () => routeState.value,
  }
})

vi.mock('@/api/contest', () => contestApiMocks)

describe('useContestAwdDefenseWorkbenchPage', () => {
  beforeEach(() => {
    routeState.value!.params.id = '7'
    routeState.value!.params.serviceId = '7009'
    contestApiMocks.getContestAWDWorkspace.mockReset()
    contestApiMocks.getContestChallenges.mockReset()
    contestApiMocks.requestContestAWDDefenseDirectory.mockReset()
    contestApiMocks.requestContestAWDDefenseFile.mockReset()

    contestApiMocks.getContestAWDWorkspace.mockResolvedValue({
      contest_id: '7',
      current_round: null,
      my_team: { team_id: '13', team_name: 'Red' },
      services: [
        {
          service_id: '7009',
          awd_challenge_id: '9',
          instance_id: '9001',
          instance_status: 'running',
          service_status: 'up',
          attack_received: 0,
          updated_at: '2026-05-05T08:00:00Z',
        },
      ],
      targets: [],
      recent_events: [],
    })
    contestApiMocks.getContestChallenges.mockResolvedValue([
      {
        id: '21',
        challenge_id: '9',
        awd_challenge_id: '9',
        awd_service_id: '7009',
        title: 'Bank Portal',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.requestContestAWDDefenseDirectory.mockResolvedValue({
      path: '.',
      entries: [{ name: 'app.py', path: 'app.py', type: 'file', size: 13 }],
    })
  })

  it('进入页面后应自动加载根目录并生成返回战场链接', async () => {
    let result!: ReturnType<typeof useContestAwdDefenseWorkbenchPage>

    mount(
      defineComponent({
        setup() {
          result = useContestAwdDefenseWorkbenchPage()
          return () => null
        },
      })
    )

    await flushPromises()

    expect(result.backLink.value).toEqual({
      name: 'ContestDetail',
      params: { id: '7' },
      query: { panel: 'challenges' },
    })
    expect(contestApiMocks.requestContestAWDDefenseDirectory).toHaveBeenCalledWith('7', '7009', '.')
    expect(result.serviceTitle.value).toBe('Bank Portal')
    expect(result.currentDirectoryPath.value).toBe('.')
    expect(result.error.value).toBe('')
  })

  it('旧文件响应晚到时不应覆盖当前文件状态', async () => {
    let resolveFirst!: (value: { path: string; content: string; size: number }) => void
    let resolveSecond!: (value: { path: string; content: string; size: number }) => void
    contestApiMocks.requestContestAWDDefenseFile
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveFirst = resolve
          })
      )
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveSecond = resolve
          })
      )

    let result!: ReturnType<typeof useContestAwdDefenseWorkbenchPage>

    mount(
      defineComponent({
        setup() {
          result = useContestAwdDefenseWorkbenchPage()
          return () => null
        },
      })
    )

    await flushPromises()

    void result.openFile('app.py')
    void result.openFile('index.php')

    resolveSecond({ path: 'index.php', content: '<?php echo 1;', size: 14 })
    await flushPromises()
    resolveFirst({ path: 'app.py', content: "print('late')", size: 13 })
    await flushPromises()

    expect(result.file.value?.path).toBe('index.php')
    expect(result.file.value?.content).toBe('<?php echo 1;')
  })

  it('切换服务后旧服务的文件响应不应写回当前页面', async () => {
    let resolveOldFile!: (value: { path: string; content: string; size: number }) => void
    contestApiMocks.requestContestAWDDefenseFile.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveOldFile = resolve
        })
    )
    let result!: ReturnType<typeof useContestAwdDefenseWorkbenchPage>

    mount(
      defineComponent({
        setup() {
          result = useContestAwdDefenseWorkbenchPage()
          return () => null
        },
      })
    )

    await flushPromises()

    void result.openFile('app.py')

    contestApiMocks.getContestChallenges.mockResolvedValueOnce([
      {
        id: '22',
        challenge_id: '10',
        awd_challenge_id: '10',
        awd_service_id: '7010',
        title: 'Service B',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 0,
        is_solved: false,
      },
    ])
    contestApiMocks.requestContestAWDDefenseDirectory.mockRejectedValueOnce({ status: 403 })

    routeState.value!.params.serviceId = '7010'
    await flushPromises()
    await flushPromises()

    resolveOldFile({ path: 'app.py', content: "print('late old service')", size: 25 })
    await flushPromises()

    expect(contestApiMocks.requestContestAWDDefenseDirectory).toHaveBeenCalledWith('7', '7010', '.')
    expect(result.file.value).toBeNull()
  })
})
