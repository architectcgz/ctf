import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import type { ReviewArchiveData } from '@/api/contracts'
import TeacherReviewArchiveSummarySection from './TeacherReviewArchiveSummarySection.vue'

function createArchive(): ReviewArchiveData {
  return {
    generated_at: '2026-04-01T09:30:00Z',
    student: {
      id: 'stu-1',
      username: 'alice',
      name: 'Alice',
      class_name: 'Class A',
    },
    summary: {
      total_challenges: 20,
      total_solved: 10,
      total_score: 650,
      rank: 2,
      total_attempts: 30,
      timeline_event_count: 8,
      evidence_event_count: 6,
      writeup_count: 2,
      manual_review_count: 1,
      correct_submission_count: 4,
      last_activity_at: '2026-04-01T09:20:00Z',
    },
    skill_profile: {
      dimensions: [
        { key: 'crypto', name: '密码', value: 48 },
        { key: 'web', name: 'Web', value: 82 },
      ],
      updated_at: '2026-04-01T09:30:00Z',
    },
    timeline: [],
    evidence: [],
    writeups: [],
    manual_reviews: [],
    teacher_observations: {
      items: [],
    },
  }
}

describe('TeacherReviewArchiveSummarySection', () => {
  it('应渲染摘要指标与最近活跃信息', () => {
    const wrapper = mount(TeacherReviewArchiveSummarySection, {
      props: {
        archive: createArchive(),
      },
    })

    expect(wrapper.text()).toContain('50%')
    expect(wrapper.text()).toContain('有效提交')
    expect(wrapper.text()).toContain('4')
    expect(wrapper.text()).toContain('2026')
    expect(wrapper.findAll('.summary-card .lucide')).toHaveLength(3)
  })

  it('应按得分从高到低排序能力维度', () => {
    const wrapper = mount(TeacherReviewArchiveSummarySection, {
      props: {
        archive: createArchive(),
      },
    })

    const skillTitles = wrapper.findAll('.skill-bars__head strong').map((node) => node.text())
    expect(skillTitles[0]).toBe('Web')
    expect(skillTitles[1]).toBe('密码')
  })
})
