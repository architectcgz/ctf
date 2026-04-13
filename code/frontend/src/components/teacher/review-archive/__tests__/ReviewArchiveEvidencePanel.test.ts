import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import ReviewArchiveEvidencePanel from '../ReviewArchiveEvidencePanel.vue'
import type {
  ReviewArchiveEvidenceItemData,
  ReviewArchiveManualReviewItemData,
  ReviewArchiveWriteupItemData,
  TimelineEvent,
} from '@/api/contracts'

function buildTimeline(): TimelineEvent[] {
  return [
    {
      id: 'practice-instance-start',
      type: 'instance',
      title: 'web-1',
      detail: '启动练习实例',
      created_at: '2026-04-13T09:00:00Z',
      challenge_id: 'web-1',
      meta: { raw_type: 'instance_start' },
    },
    {
      id: 'practice-solve',
      type: 'solve',
      title: 'web-1',
      detail: '提交命中 Flag',
      created_at: '2026-04-13T09:08:00Z',
      challenge_id: 'web-1',
      is_correct: true,
      points: 100,
      meta: { raw_type: 'flag_submit' },
    },
    {
      id: 'awd-success-alpha',
      type: 'awd_attack_submit',
      title: 'awd-web',
      detail: 'AWD 攻击命中 Alpha，得分 120',
      created_at: '2026-04-13T10:01:00Z',
      challenge_id: 'awd-web',
      is_correct: true,
      points: 120,
      meta: { raw_type: 'awd_attack_submit' },
    },
    {
      id: 'awd-failed-beta',
      type: 'awd_attack_submit',
      title: 'awd-web',
      detail: 'AWD 攻击未命中 Beta',
      created_at: '2026-04-13T10:05:00Z',
      challenge_id: 'awd-web',
      is_correct: false,
      meta: { raw_type: 'awd_attack_submit' },
    },
  ]
}

function buildEvidence(): ReviewArchiveEvidenceItemData[] {
  return [
    {
      type: 'instance_access',
      challenge_id: 'web-1',
      title: 'web-1',
      detail: '访问攻击目标',
      timestamp: '2026-04-13T09:01:00Z',
      meta: { event_stage: 'access' },
    },
    {
      type: 'instance_proxy_request',
      challenge_id: 'web-1',
      title: 'web-1',
      detail: '经平台代理发起 POST /login',
      timestamp: '2026-04-13T09:02:00Z',
      meta: { event_stage: 'exploit' },
    },
    {
      type: 'challenge_submission',
      challenge_id: 'web-1',
      title: 'web-1',
      detail: '提交命中 Flag',
      timestamp: '2026-04-13T09:08:00Z',
      meta: { event_stage: 'submit', is_correct: true, points: 100 },
    },
    {
      type: 'awd_attack_submission',
      challenge_id: 'awd-web',
      title: 'awd-web',
      detail: 'AWD 攻击命中 Alpha，得分 120',
      timestamp: '2026-04-13T10:01:00Z',
      meta: {
        event_stage: 'exploit',
        is_success: true,
        score_gained: 120,
        victim_team_name: 'Alpha',
      },
    },
    {
      type: 'awd_attack_submission',
      challenge_id: 'awd-web',
      title: 'awd-web',
      detail: 'AWD 攻击未命中 Beta',
      timestamp: '2026-04-13T10:05:00Z',
      meta: {
        event_stage: 'exploit',
        is_success: false,
        score_gained: 0,
        victim_team_name: 'Beta',
      },
    },
  ]
}

function buildWriteups(): ReviewArchiveWriteupItemData[] {
  return [
    {
      id: 'writeup-1',
      challenge_id: 'web-1',
      challenge_title: 'web-1',
      title: '从回显到 flag',
      submission_status: 'published',
      visibility_status: 'visible',
      is_recommended: true,
      published_at: '2026-04-13T09:20:00Z',
      updated_at: '2026-04-13T09:20:00Z',
    },
  ]
}

function buildManualReviews(): ReviewArchiveManualReviewItemData[] {
  return [
    {
      id: 'manual-1',
      challenge_id: 'web-1',
      challenge_title: 'web-1',
      answer: '完整答案正文',
      review_status: 'approved',
      submitted_at: '2026-04-13T09:18:00Z',
      score: 100,
      review_comment: '通过',
      reviewer_name: 'teacher-a',
    },
  ]
}

describe('ReviewArchiveEvidencePanel', () => {
  it('应该将练习事件与AWD事件分到不同复盘分区', () => {
    const wrapper = mount(ReviewArchiveEvidencePanel, {
      props: {
        timeline: buildTimeline(),
        evidence: buildEvidence(),
        writeups: buildWriteups(),
        manualReviews: buildManualReviews(),
      },
    })

    expect(wrapper.text()).toContain('练习复盘')
    expect(wrapper.text()).toContain('AWD 复盘')
    expect(wrapper.text()).toContain('web-1')
    expect(wrapper.text()).toContain('awd-web')
    expect(wrapper.text()).toContain('Alpha')
    expect(wrapper.text()).toContain('Beta')
    expect(wrapper.text()).not.toContain('web-1 / Alpha')
  })

  it('应该按challenge聚合练习案例并按victim team拆分AWD案例', async () => {
    const wrapper = mount(ReviewArchiveEvidencePanel, {
      props: {
        timeline: buildTimeline(),
        evidence: buildEvidence(),
        writeups: buildWriteups(),
        manualReviews: buildManualReviews(),
      },
    })

    const practiceCards = wrapper.findAll('[data-testid="practice-case-card"]')
    const awdCards = wrapper.findAll('[data-testid="awd-case-card"]')

    expect(practiceCards).toHaveLength(1)
    expect(practiceCards[0].text()).toContain('web-1')

    expect(awdCards).toHaveLength(2)
    const awdCardTexts = awdCards.map((card) => card.text())
    expect(awdCardTexts.some((text) => text.includes('Alpha'))).toBe(true)
    expect(awdCardTexts.some((text) => text.includes('Beta'))).toBe(true)

    await wrapper.find('[data-testid="practice-case-toggle"]').trigger('click')

    expect(practiceCards[0].text()).toContain('从回显到 flag')
    expect(practiceCards[0].text()).toContain('teacher-a')
  })

  it('应该默认折叠案例卡并在展开后显示事件明细', async () => {
    const wrapper = mount(ReviewArchiveEvidencePanel, {
      props: {
        timeline: buildTimeline(),
        evidence: buildEvidence(),
        writeups: buildWriteups(),
        manualReviews: buildManualReviews(),
      },
    })

    const toggle = wrapper.find('[data-testid="practice-case-toggle"]')
    expect(toggle.attributes('aria-expanded')).toBe('false')
    expect(wrapper.text()).not.toContain('经平台代理发起 POST /login')

    await toggle.trigger('click')

    expect(toggle.attributes('aria-expanded')).toBe('true')
    expect(wrapper.text()).toContain('经平台代理发起 POST /login')
    expect(wrapper.text()).toContain('提交命中 Flag')
  })
})
