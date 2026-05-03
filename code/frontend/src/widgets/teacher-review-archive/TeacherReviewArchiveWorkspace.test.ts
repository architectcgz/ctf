import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import type { ReviewArchiveData } from '@/api/contracts'
import TeacherReviewArchiveWorkspace from './TeacherReviewArchiveWorkspace.vue'

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

describe('TeacherReviewArchiveWorkspace', () => {
  it('应透传 hero 和错误态重载事件', async () => {
    const wrapper = mount(TeacherReviewArchiveWorkspace, {
      props: {
        archive: null,
        loading: false,
        error: '加载失败',
        exporting: false,
      },
      global: {
        stubs: {
          ReviewArchiveHero: {
            name: 'ReviewArchiveHero',
            template:
              '<div><button id="back" @click="$emit(\'back\')" /><button id="open" @click="$emit(\'openAnalysis\')" /><button id="export" @click="$emit(\'exportArchive\')" /></div>',
          },
        },
      },
    })

    await wrapper.get('#back').trigger('click')
    await wrapper.get('#open').trigger('click')
    await wrapper.get('#export').trigger('click')
    await wrapper.get('.ui-btn.ui-btn--primary').trigger('click')

    expect(wrapper.emitted('back')).toBeTruthy()
    expect(wrapper.emitted('openAnalysis')).toBeTruthy()
    expect(wrapper.emitted('exportArchive')).toBeTruthy()
    expect(wrapper.emitted('reload')).toBeTruthy()
  })

  it('应渲染摘要指标和排序后的能力画像', () => {
    const wrapper = mount(TeacherReviewArchiveWorkspace, {
      props: {
        archive: createArchive(),
        loading: false,
        error: null,
        exporting: false,
      },
    })

    const skillTitles = wrapper.findAll('.skill-bars__head strong').map((node) => node.text())
    expect(wrapper.text()).toContain('50%')
    expect(wrapper.text()).toContain('有效提交')
    expect(skillTitles[0]).toBe('Web')
    expect(skillTitles[1]).toBe('密码')
  })
})
