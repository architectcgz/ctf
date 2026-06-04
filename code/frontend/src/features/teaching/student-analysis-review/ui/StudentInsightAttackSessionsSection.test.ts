import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import StudentInsightAttackSessionsSection from './StudentInsightAttackSessionsSection.vue'
import studentInsightAttackSessionsSectionSource from './StudentInsightAttackSessionsSection.vue?raw'

describe('StudentInsightAttackSessionsSection', () => {
  it('应将复盘摘要放在列表容器上方', () => {
    const wrapper = mount(StudentInsightAttackSessionsSection, {
      props: {
        attackSessions: {
          summary: {
            total_sessions: 1,
            success_count: 1,
            failed_count: 0,
            in_progress_count: 0,
            unknown_count: 0,
            event_count: 2,
            capture_available_count: 0,
          },
          sessions: [
            {
              id: 'sess-1',
              mode: 'practice',
              student_id: 'stu-1',
              challenge_id: '11',
              title: 'web-1',
              started_at: '2026-05-03T08:00:00Z',
              ended_at: '2026-05-03T08:10:00Z',
              result: 'success',
              event_count: 2,
              capture_count: 0,
              events: [],
            },
          ],
        },
        evidence: {
          summary: {
            total_events: 6,
            proxy_request_count: 2,
            submit_count: 3,
            success_count: 1,
            challenge_id: '11',
          },
          events: [],
        },
        reviewChallengeOptions: [],
        reviewWorkspaceLoading: false,
        reviewWorkspaceQuery: {},
      },
    })

    expect(wrapper.find('.student-insight-kpi-grid').exists()).toBe(true)
    expect(wrapper.get('.student-insight-kpi-grid').classes()).toContain('teacher-summary-grid')
    expect(wrapper.get('.student-insight-kpi-grid').classes()).toContain('metric-panel-default-surface')
    expect(wrapper.get('.student-insight-kpi-grid').classes()).not.toContain('metric-panel-teacher-surface')
    expect(wrapper.find('.evidence-state-surface').exists()).toBe(true)
    expect(wrapper.html().indexOf('student-insight-kpi-grid')).toBeLessThan(
      wrapper.html().indexOf('evidence-state-surface')
    )
    expect(studentInsightAttackSessionsSectionSource).toMatch(
      /class="student-insight-kpi-grid[\s\S]*teacher-summary-grid[\s\S]*metric-panel-default-surface[\s\S]*<StudentInsightStateSurface/s
    )
  })
})
