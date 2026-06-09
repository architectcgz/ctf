import { vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'

export { ApiError } from '@/api/request'
export { default as challengeDetailSource } from '@/pages/challenges/ChallengeDetailRoutePage.vue?raw'
export { default as challengeDetailShellSource } from '@/widgets/challenge-detail-workspace/ChallengeDetailPage.vue?raw'
export { default as challengeDetailPageSource } from '@/features/challenge-detail/model/useChallengeDetailPage.ts?raw'
export { default as challengeDetailWorkspaceSource } from '@/widgets/challenge-detail-workspace/ChallengeDetailWorkspace.vue?raw'
export { default as challengeDetailRoutesSource } from '@/features/challenge-detail/model/challengeDetailRoutes.ts?raw'
export { default as challengeWorkspaceShellSource } from '@/features/challenge-detail/ui/ChallengeWorkspaceShell.vue?raw'
export { default as challengeQuestionPanelSource } from '@/features/challenge-detail/ui/ChallengeQuestionPanel.vue?raw'
export { default as challengeSolutionsPanelSource } from '@/features/challenge-detail/ui/ChallengeSolutionsPanel.vue?raw'
export { default as challengeSubmissionRecordsPanelSource } from '@/features/challenge-detail/ui/ChallengeSubmissionRecordsPanel.vue?raw'
export { default as challengeWriteupPanelSource } from '@/features/challenge-detail/ui/ChallengeWriteupPanel.vue?raw'
export { default as challengeActionAsideSource } from '@/features/challenge-detail/ui/ChallengeActionAside.vue?raw'
export { default as challengeInstanceCardSource } from '@/features/challenge-detail/ui/ChallengeInstanceCard.vue?raw'
export { default as instancePresentationSource } from '@/entities/instance/model/presentation.ts?raw'

const challengeApiMocks = vi.hoisted(() => ({
  getChallengeDetail: vi.fn(),
  getChallengeWriteup: vi.fn(),
  getRecommendedChallengeSolutions: vi.fn(),
  getCommunityChallengeSolutions: vi.fn(),
  getMyChallengeWriteupSubmission: vi.fn(),
  getMyChallengeSubmissionRecords: vi.fn(),
  upsertChallengeWriteupSubmission: vi.fn(),
  submitFlag: vi.fn(),
  unlockHint: vi.fn(),
  createInstance: vi.fn(),
  downloadAttachment: vi.fn(),
}))

const instanceApiMocks = vi.hoisted(() => ({
  getMyInstances: vi.fn(),
  destroyInstance: vi.fn(),
  extendInstance: vi.fn(),
  requestInstanceAccess: vi.fn(),
}))

export function createDeferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((nextResolve, nextReject) => {
    resolve = nextResolve
    reject = nextReject
  })

  return { promise, resolve, reject }
}

vi.mock('@/api/challenge', () => challengeApiMocks)
vi.mock('@/api/instance', () => instanceApiMocks)

export let router: ReturnType<typeof createRouter>

export function resetChallengeDetailTestHarness() {
  window.history.replaceState({}, '', '/challenges/1')
  router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/challenges', name: 'Challenges', component: { template: '<div />' } },
      { path: '/challenges/:id', name: 'ChallengeDetail', component: { template: '<div />' } },
    ],
  })

  Object.values(challengeApiMocks).forEach((mock) => {
    mock.mockReset()
  })
  Object.values(instanceApiMocks).forEach((mock) => {
    mock.mockReset()
  })

  challengeApiMocks.getChallengeDetail.mockResolvedValue({
    id: '1',
    title: 'Test Challenge',
    description: '<p>Test description</p>',
    category: 'web',
    difficulty: 'easy',
    tags: ['test'],
    points: 100,
    need_target: true,
    is_solved: false,
    attachment_url: 'https://example.com/file.zip',
    hints: [
      {
        id: 'hint-1',
        level: 1,
        title: '入口',
        content: '先观察登录表单的参数。',
      },
    ],
  })
  challengeApiMocks.getChallengeWriteup.mockResolvedValue({
    id: 'writeup-1',
    challenge_id: '1',
    title: '官方题解',
    content: '<p>Exploit path</p>',
    visibility: 'public',
    requires_spoiler_warning: true,
    created_at: '2026-03-10T00:00:00.000Z',
    updated_at: '2026-03-10T01:00:00.000Z',
  })
  challengeApiMocks.getRecommendedChallengeSolutions.mockResolvedValue([
    {
      id: 'recommended-1',
      source_type: 'official',
      source_id: 'writeup-1',
      challenge_id: '1',
      title: '精选官方题解',
      content: '<p>Exploit path</p>',
      author_name: '官方题解',
      is_recommended: true,
      recommended_at: '2026-03-10T02:00:00.000Z',
      updated_at: '2026-03-10T02:00:00.000Z',
    },
  ])
  challengeApiMocks.getCommunityChallengeSolutions.mockResolvedValue({
    list: [
      {
        id: 'community-1',
        challenge_id: '1',
        user_id: 'stu-2',
        title: '我的 SQLi 复盘',
        content: '先找注入点，再构造 payload。',
        content_preview: '先找注入点，再构造 payload。',
        author_name: 'student_b',
        submission_status: 'published',
        visibility_status: 'visible',
        is_recommended: false,
        published_at: '2026-03-12T01:00:00.000Z',
        updated_at: '2026-03-12T01:00:00.000Z',
      },
    ],
    total: 1,
    page: 1,
    page_size: 20,
  })
  challengeApiMocks.getMyChallengeWriteupSubmission.mockResolvedValue(null)
  challengeApiMocks.getMyChallengeSubmissionRecords.mockResolvedValue([])
  challengeApiMocks.upsertChallengeWriteupSubmission.mockResolvedValue({
    id: 'submission-1',
    user_id: 'stu-1',
    challenge_id: '1',
    title: '我的题解',
    content: '先找回显，再定位注入。',
    submission_status: 'draft',
    visibility_status: 'visible',
    is_recommended: false,
    created_at: '2026-03-12T00:00:00.000Z',
    updated_at: '2026-03-12T00:30:00.000Z',
  })
  challengeApiMocks.submitFlag.mockReset()
  challengeApiMocks.unlockHint.mockReset()
  challengeApiMocks.createInstance.mockResolvedValue({
    id: 'inst-1',
    challenge_id: '1',
    status: 'running',
    access_url: 'http://target.test',
    flag_type: 'static',
    expires_at: '2099-01-01T00:00:00Z',
    remaining_extends: 2,
    created_at: '2026-03-12T00:00:00.000Z',
  })
  challengeApiMocks.downloadAttachment.mockReset()

  instanceApiMocks.getMyInstances.mockResolvedValue([])
  instanceApiMocks.destroyInstance.mockReset()
  instanceApiMocks.extendInstance.mockReset()
  instanceApiMocks.requestInstanceAccess.mockReset()
}

export function cleanupChallengeDetailTestHarness() {
  vi.clearAllTimers()
  vi.useRealTimers()
}

export { challengeApiMocks, instanceApiMocks }
