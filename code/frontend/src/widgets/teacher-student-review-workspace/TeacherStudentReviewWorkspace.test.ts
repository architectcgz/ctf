import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import TeacherStudentReviewWorkspace from './TeacherStudentReviewWorkspace.vue'

describe('TeacherStudentReviewWorkspace', () => {
  it('应渲染空状态', () => {
    const wrapper = mount(TeacherStudentReviewWorkspace, {
      props: {
        evidence: null,
        attackSessions: null,
        loading: false,
        query: {},
      },
    })

    expect(wrapper.text()).toContain('暂无攻击会话')
  })

  it('应渲染摘要和事件标签', () => {
    const wrapper = mount(TeacherStudentReviewWorkspace, {
      props: {
        evidence: {
          summary: {
            total_events: 6,
            proxy_request_count: 2,
            submit_count: 3,
            success_count: 1,
            challenge_id: '11',
          },
          events: [
            {
              type: 'writeup',
              challenge_id: '11',
              title: 'web-1',
              detail: '提交题解',
              timestamp: '2026-05-03T08:12:00Z',
            },
          ],
        },
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
              events: [
                {
                  id: 'evt-1',
                  type: 'instance_proxy_request',
                  stage: 'exploit',
                  source: 'audit_logs',
                  occurred_at: '2026-05-03T08:01:00Z',
                  actor: { user_id: 'stu-1' },
                  target: { challenge_id: '11' },
                  summary: 'POST /login 200',
                  meta: {
                    request_method: 'POST',
                    target_path: '/login',
                    status_code: 200,
                  },
                  capture_available: false,
                },
              ],
            },
          ],
        },
        loading: false,
        query: {
          with_events: true,
          limit: 20,
          offset: 0,
        },
      },
    })

    expect(wrapper.text()).toContain('会话数')
    expect(wrapper.text()).toContain('成功会话')
    expect(wrapper.text()).toContain('实操请求')
    expect(wrapper.text()).toContain('训练闭环')
    expect(wrapper.text()).toContain('训练')
    expect(wrapper.text()).toContain('POST')
    expect(wrapper.text()).toContain('/login')
    expect(wrapper.findAll('.metric-panel-label svg')).toHaveLength(4)
  })

  it('应在筛选变更时发出查询更新事件', async () => {
    const wrapper = mount(TeacherStudentReviewWorkspace, {
      props: {
        evidence: {
          summary: {
            total_events: 0,
            proxy_request_count: 0,
            submit_count: 0,
            success_count: 0,
            challenge_id: '11',
          },
          events: [
            {
              type: 'instance_access',
              challenge_id: '11',
              title: 'web-1',
              detail: '访问',
              timestamp: '2026-05-03T08:00:00Z',
            },
          ],
        },
        attackSessions: {
          summary: {
            total_sessions: 1,
            success_count: 1,
            failed_count: 0,
            in_progress_count: 0,
            unknown_count: 0,
            event_count: 1,
            capture_available_count: 0,
          },
          sessions: [
            {
              id: 'sess-1',
              mode: 'practice',
              student_id: 'stu-1',
              title: 'web-1',
              started_at: '2026-05-03T08:00:00Z',
              ended_at: '2026-05-03T08:10:00Z',
              result: 'success',
              event_count: 1,
              capture_count: 0,
              events: [],
            },
          ],
        },
        loading: false,
        query: {},
      },
    })

    const selects = wrapper.findAll('select')
    await selects[0].setValue('11')
    await selects[1].setValue('awd')
    await selects[2].setValue('failed')

    expect(wrapper.emitted('updateFilters')?.[0]).toEqual([{ challenge_id: '11' }])
    expect(wrapper.emitted('updateFilters')?.[1]).toEqual([{ mode: 'awd' }])
    expect(wrapper.emitted('updateFilters')?.[2]).toEqual([{ result: 'failed' }])
  })
})
