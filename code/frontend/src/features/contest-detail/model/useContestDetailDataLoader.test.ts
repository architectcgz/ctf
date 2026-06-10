import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'

import type {
  ContestAnnouncement,
  ContestChallengeItem,
  ContestDetailData,
  TeamData,
} from '@/api/contracts'
import { useContestDetailDataLoader } from './useContestDetailDataLoader'

const contestApiMocks = vi.hoisted(() => ({
  getAnnouncementSync: vi.fn(),
  getAnnouncements: vi.fn(),
  getContestChallenges: vi.fn(),
  getContestDetail: vi.fn(),
  getMyTeam: vi.fn(),
}))

vi.mock('@/api/contest', () => contestApiMocks)

function buildContest(): ContestDetailData {
  return {
    id: 'contest-1',
    title: '2026 春季赛',
    description: '详情页',
    mode: 'jeopardy',
    status: 'running',
    starts_at: '2026-04-22T08:00:00.000Z',
    ends_at: '2026-04-22T18:00:00.000Z',
    scoreboard_frozen: false,
  }
}

describe('useContestDetailDataLoader', () => {
  beforeEach(() => {
    contestApiMocks.getAnnouncementSync.mockReset()
    contestApiMocks.getAnnouncements.mockReset()
    contestApiMocks.getContestChallenges.mockReset()
    contestApiMocks.getContestDetail.mockReset()
    contestApiMocks.getMyTeam.mockReset()

    contestApiMocks.getContestDetail.mockResolvedValue(buildContest())
    contestApiMocks.getMyTeam.mockResolvedValue(null)
    contestApiMocks.getContestChallenges.mockResolvedValue([])
    contestApiMocks.getAnnouncementSync.mockResolvedValue({
      events: [],
      next_cursor: '8',
      has_more: false,
    })
    contestApiMocks.getAnnouncements.mockResolvedValue([
      {
        id: 'announcement-1',
        title: '报名提醒',
        content: '请在今晚前完成组队。',
        created_at: '2026-04-22T09:00:00.000Z',
      },
    ])
  })

  it('加载详情页公告时应先锚定同步 cursor 再读取列表', async () => {
    const contest = ref<ContestDetailData | null>(null)
    const team = ref<TeamData | null>(null)
    const challenges = ref<ContestChallengeItem[]>([])
    const announcements = ref<ContestAnnouncement[]>([])
    const announcementsError = ref('')
    const loading = ref(false)

    const loader = useContestDetailDataLoader({
      contestId: 'contest-1',
      contest,
      team,
      challenges,
      announcements,
      announcementsError,
      loading,
      resetPageState: vi.fn(),
      startCountdown: vi.fn(),
      stopCountdown: vi.fn(),
      syncSelectedChallengeFromQuery: vi.fn(),
      clearSubmissionState: vi.fn(),
      onLoadFailed: vi.fn(),
    })

    await loader.loadPage()

    expect(contestApiMocks.getAnnouncementSync).toHaveBeenCalledWith('contest-1')
    expect(contestApiMocks.getAnnouncements).toHaveBeenCalledWith('contest-1')
    expect(contestApiMocks.getAnnouncementSync.mock.invocationCallOrder[0]).toBeLessThan(
      contestApiMocks.getAnnouncements.mock.invocationCallOrder[0]
    )
    expect(announcements.value).toEqual([expect.objectContaining({ id: 'announcement-1' })])
    expect(announcementsError.value).toBe('')
  })
})
