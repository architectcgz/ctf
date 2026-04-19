import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import TeacherAWDReviewDetail from '../TeacherAWDReviewDetail.vue'
import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'
import { useAuthStore } from '@/stores/auth'

const routeMock = {
  params: {
    contestId: 'contest-1',
  },
  query: {} as Record<string, string>,
}

const teacherApiMocks = vi.hoisted(() => ({
  getTeacherAWDReview: vi.fn(),
  exportTeacherAWDReviewArchive: vi.fn(),
  exportTeacherAWDReviewReport: vi.fn(),
}))

const pushMock = vi.fn()
const replaceMock = vi.fn()

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeMock,
    useRouter: () => ({ push: pushMock, replace: replaceMock }),
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

describe('TeacherAWDReviewDetail', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    routeMock.params.contestId = 'contest-1'
    routeMock.query = {}
    pushMock.mockReset()
    replaceMock.mockReset()
    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())

    teacherApiMocks.getTeacherAWDReview.mockResolvedValue({
      generated_at: '2026-04-12T10:00:00Z',
      scope: {
        snapshot_type: 'live',
        requested_by: 5,
        requested_id: 'contest-1',
      },
      contest: {
        id: 'contest-1',
        title: '春季 AWD 联训',
        mode: 'awd',
        status: 'running',
        current_round: 2,
        round_count: 4,
        team_count: 6,
        export_ready: false,
      },
      overview: {
        round_count: 4,
        team_count: 6,
        service_count: 12,
        attack_count: 8,
        traffic_count: 20,
      },
      rounds: [
        {
          id: 'round-1',
          contest_id: 'contest-1',
          round_number: 1,
          status: 'finished',
          service_count: 6,
          attack_count: 3,
          traffic_count: 8,
        },
        {
          id: 'round-2',
          contest_id: 'contest-1',
          round_number: 2,
          status: 'running',
          service_count: 6,
          attack_count: 5,
          traffic_count: 12,
        },
      ],
      selected_round: {
        round: {
          id: 'round-2',
          contest_id: 'contest-1',
          round_number: 2,
          status: 'running',
          service_count: 6,
          attack_count: 5,
          traffic_count: 12,
        },
        teams: [
          {
            team_id: 'team-1',
            team_name: 'Blue Team',
            captain_id: 'stu-1',
            total_score: 320,
            member_count: 4,
          },
        ],
        services: [],
        attacks: [],
        traffic: [],
      },
    })
    teacherApiMocks.exportTeacherAWDReviewArchive.mockResolvedValue({
      report_id: '31',
      status: 'processing',
    })
    teacherApiMocks.exportTeacherAWDReviewReport.mockResolvedValue({
      report_id: '32',
      status: 'processing',
    })

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'teacher-1',
        username: 'teacher',
        role: 'teacher',
      },
      'token'
    )
  })

  it('默认显示整场总览，并在进行中赛事上禁用教师报告导出', async () => {
    const wrapper = mount(TeacherAWDReviewDetail)

    await flushPromises()

    expect(teacherApiMocks.getTeacherAWDReview).toHaveBeenCalledWith('contest-1', {
      round: undefined,
      team_id: undefined,
    })
    expect(wrapper.text()).toContain('AWD复盘')
    expect(wrapper.text()).toContain('春季 AWD 联训')
    expect(
      wrapper.find('[data-testid="awd-review-export-report"]').attributes('disabled')
    ).toBeDefined()
  })

  it('query 带 round 时应切到单轮摘要，并在赛后开放两个导出按钮', async () => {
    routeMock.query = { round: '2' }
    teacherApiMocks.getTeacherAWDReview.mockResolvedValueOnce({
      generated_at: '2026-04-12T10:00:00Z',
      scope: {
        snapshot_type: 'final',
        requested_by: 5,
        requested_id: 'contest-1',
      },
      contest: {
        id: 'contest-1',
        title: '期末 AWD 复盘',
        mode: 'awd',
        status: 'ended',
        current_round: 4,
        round_count: 4,
        team_count: 6,
        export_ready: true,
      },
      overview: {
        round_count: 4,
        team_count: 6,
        service_count: 12,
        attack_count: 8,
        traffic_count: 20,
      },
      rounds: [],
      selected_round: {
        round: {
          id: 'round-2',
          contest_id: 'contest-1',
          round_number: 2,
          status: 'finished',
          service_count: 6,
          attack_count: 5,
          traffic_count: 12,
        },
        teams: [],
        services: [],
        attacks: [],
        traffic: [],
      },
    })

    const wrapper = mount(TeacherAWDReviewDetail)

    await flushPromises()

    expect(teacherApiMocks.getTeacherAWDReview).toHaveBeenCalledWith('contest-1', {
      round: 2,
      team_id: undefined,
    })
    expect(wrapper.text()).toContain('第 2 轮')
    expect(
      wrapper.find('[data-testid="awd-review-export-archive"]').attributes('disabled')
    ).toBeUndefined()
    expect(
      wrapper.find('[data-testid="awd-review-export-report"]').attributes('disabled')
    ).toBeUndefined()
  })

  it('轮次切换区应并入目录段结构，不再保留独立筛选卡片壳', () => {
    expect(awdReviewDetailSource).toContain(
      'class="workspace-directory-section teacher-directory-section awd-review-round-section"'
    )
    expect(awdReviewDetailSource).toContain('class="list-heading"')
    expect(awdReviewDetailSource).not.toContain('teacher-controls-title')
    expect(awdReviewDetailSource).not.toContain('teacher-controls-copy')
    expect(awdReviewDetailSource).not.toContain(
      '默认展示整场总览；切换到单轮后，可继续按队伍查看该轮服务、攻击和流量证据。'
    )
  })

  it('detail 页头部应切到 workspace 语义，不再保留 teacher journal eyebrow', () => {
    expect(awdReviewDetailSource).toContain(
      '<header class="teacher-topbar workspace-tab-heading awd-review-detail-header">'
    )
    expect(awdReviewDetailSource).toContain(
      '<div class="teacher-heading workspace-tab-heading__main">'
    )
    expect(awdReviewDetailSource).toContain(
      '<div class="workspace-overline awd-review-detail-overline">AWD Review</div>'
    )
    expect(awdReviewDetailSource).toContain(
      '<h1 class="teacher-title workspace-page-title">{{ activeTitle }}</h1>'
    )
    expect(awdReviewDetailSource).toContain('<p class="teacher-copy workspace-page-copy">')
    expect(awdReviewDetailSource).toMatch(
      /\.awd-review-detail-overline\s*\{[\s\S]*font-size:\s*var\(--journal-overline-font-size,\s*var\(--font-size-0-70\)\);[\s\S]*letter-spacing:\s*var\(--journal-overline-letter-spacing,\s*0\.2em\);[\s\S]*text-transform:\s*uppercase;[\s\S]*color:\s*var\(--journal-accent,\s*var\(--color-primary\)\);/s
    )
    expect(awdReviewDetailSource).not.toContain(
      '<div class="teacher-surface-eyebrow journal-eyebrow">AWD Review Workspace</div>'
    )
  })

  it('detail 页各复盘分区应改用 workspace overline，而不是继续保留 journal eyebrow', () => {
    expect(awdReviewDetailSource).toContain(
      '<div class="workspace-overline awd-review-section-overline">Round Summary</div>'
    )
    expect(awdReviewDetailSource).toContain(
      '<div class="workspace-overline awd-review-section-overline">Services</div>'
    )
    expect(awdReviewDetailSource).toContain(
      '<div class="workspace-overline awd-review-section-overline">Attacks</div>'
    )
    expect(awdReviewDetailSource).toContain(
      '<div class="workspace-overline awd-review-section-overline">Traffic</div>'
    )
    expect(awdReviewDetailSource).toContain(
      '<div class="workspace-overline awd-review-section-overline">Contest Meta</div>'
    )
    expect(awdReviewDetailSource).toMatch(
      /\.awd-review-section-overline\s*\{[\s\S]*font-size:\s*var\(--journal-overline-font-size,\s*var\(--font-size-0-70\)\);[\s\S]*letter-spacing:\s*var\(--journal-overline-letter-spacing,\s*0\.2em\);[\s\S]*text-transform:\s*uppercase;[\s\S]*color:\s*var\(--journal-accent,\s*var\(--color-primary\)\);/s
    )
    expect(awdReviewDetailSource).not.toContain(
      '<div class="teacher-surface-eyebrow journal-eyebrow">Round Summary</div>'
    )
    expect(awdReviewDetailSource).not.toContain(
      '<div class="teacher-surface-eyebrow journal-eyebrow">Services</div>'
    )
    expect(awdReviewDetailSource).not.toContain(
      '<div class="teacher-surface-eyebrow journal-eyebrow">Attacks</div>'
    )
    expect(awdReviewDetailSource).not.toContain(
      '<div class="teacher-surface-eyebrow journal-eyebrow">Traffic</div>'
    )
    expect(awdReviewDetailSource).not.toContain(
      '<div class="teacher-surface-eyebrow journal-eyebrow">Contest Meta</div>'
    )
  })

  it('管理员在 AWD 复盘详情里返回目录和切换轮次时应使用后台教学运营路由', async () => {
    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
      },
      'token'
    )

    const wrapper = mount(TeacherAWDReviewDetail)

    await flushPromises()

    await wrapper.findAll('button').find((button) => button.text().includes('返回目录'))?.trigger('click')
    await wrapper.findAll('.awd-review-round-pill')[1]?.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({ name: 'AdminAWDReviewIndex' })
    expect(replaceMock).toHaveBeenCalledWith({
      name: 'AdminAWDReviewDetail',
      params: {
        contestId: 'contest-1',
      },
      query: {
        round: '1',
      },
    })
  })

  it('管理员打开 AWD 复盘详情时应切换到管理员根壳，而不是继续使用教师根壳', async () => {
    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
      },
      'token'
    )

    const wrapper = mount(TeacherAWDReviewDetail)

    await flushPromises()

    expect(wrapper.classes()).toContain('workspace-shell')
    expect(wrapper.classes()).toContain('journal-shell-admin')
    expect(wrapper.classes()).toContain('journal-hero')
    expect(wrapper.classes()).not.toContain('teacher-management-shell')
    expect(wrapper.classes()).not.toContain('teacher-surface-hero')
    expect(wrapper.find('.ui-btn').exists()).toBe(true)
    expect(wrapper.find('.teacher-btn').exists()).toBe(false)
  })
})
