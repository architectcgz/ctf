import { describe, expect, it, vi } from 'vitest'
import { reactive, ref } from 'vue'

import { useStudentAnalysisReviewQuerySync } from '../useStudentAnalysisReviewQuerySync'

describe('useStudentAnalysisReviewQuerySync', () => {
  it('应从路由 query 同步默认 review workspace 查询参数', () => {
    const route = reactive({
      query: {
        reviewMode: 'awd',
        reviewResult: 'failed',
        reviewChallengeId: '11',
      },
    })
    const sessionQuery = ref({
      with_events: true,
      limit: 20,
      offset: 0,
    })
    const setSessionQuery = vi.fn((next) => {
      sessionQuery.value = {
        ...sessionQuery.value,
        ...next,
      }
    })

    const sync = useStudentAnalysisReviewQuerySync({
      route,
      sessionQuery,
      selectedStudentId: ref('stu-1'),
      setSessionQuery,
      loadReviewWorkspace: vi.fn(),
      reloadAttackSessions: vi.fn(),
      studentIdFromRoute: () => 'stu-1',
      replaceReviewWorkspaceQuery: vi.fn(),
    })

    sync.syncReviewWorkspaceQueryFromRoute()

    expect(setSessionQuery).toHaveBeenCalledWith({
      with_events: true,
      limit: 20,
      offset: 0,
      mode: 'awd',
      result: 'failed',
      challenge_id: '11',
    })
  })

  it('切换 challenge 筛选时应回写路由 query 并刷新完整 review workspace', async () => {
    const replace = vi.fn()
    const loadReviewWorkspace = vi.fn()
    const reloadAttackSessions = vi.fn()
    const route = reactive({
      query: {
        keep: 'yes',
      },
    })
    const sessionQuery = ref({
      with_events: true,
      limit: 20,
      offset: 0,
    })

    const sync = useStudentAnalysisReviewQuerySync({
      route,
      sessionQuery,
      selectedStudentId: ref('stu-1'),
      setSessionQuery: (next) => {
        sessionQuery.value = {
          ...sessionQuery.value,
          ...next,
        }
      },
      loadReviewWorkspace,
      reloadAttackSessions,
      studentIdFromRoute: () => 'stu-1',
      replaceReviewWorkspaceQuery: async (nextQuery) => {
        await replace({
          query: {
            ...route.query,
            ...nextQuery,
          },
        })
      },
    })

    await sync.updateReviewWorkspaceFilters({ challenge_id: '11' })

    expect(replace).toHaveBeenCalledWith({
      query: {
        keep: 'yes',
        reviewMode: undefined,
        reviewResult: undefined,
        reviewChallengeId: '11',
      },
    })
    expect(loadReviewWorkspace).toHaveBeenCalledWith('stu-1')
    expect(reloadAttackSessions).not.toHaveBeenCalled()
  })
})
