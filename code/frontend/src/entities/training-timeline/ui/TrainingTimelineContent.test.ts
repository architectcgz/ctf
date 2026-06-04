import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import type { TimelineEvent } from '../model'
import TrainingTimelineContent from './TrainingTimelineContent.vue'

const timelineEvent: TimelineEvent = {
  id: 'event-1',
  type: 'solve',
  title: 'web-1',
  detail: '第 2 次提交命中 Flag',
  created_at: '2026-04-19T10:00:00Z',
}

describe('TrainingTimelineContent', () => {
  it('loading 和 loaded 应共用同一套训练记录结构', () => {
    const loadingWrapper = mount(TrainingTimelineContent, {
      props: {
        timeline: [],
        loading: true,
      },
    })
    const loadedWrapper = mount(TrainingTimelineContent, {
      props: {
        timeline: [timelineEvent],
      },
    })

    expect(loadingWrapper.find('.workspace-panel-header.timeline-header').exists()).toBe(true)
    expect(loadedWrapper.find('.workspace-panel-header.timeline-header').exists()).toBe(true)
    expect(loadedWrapper.text()).toContain('Timeline')
    expect(loadedWrapper.text()).toContain('Timeline Log')
    expect(loadingWrapper.find('.timeline-directory-shell.workspace-directory-list').exists()).toBe(
      true
    )
    expect(loadedWrapper.find('.timeline-directory-shell.workspace-directory-list').exists()).toBe(
      true
    )

    expect(loadingWrapper.findAll('.timeline-metric-card')).toHaveLength(4)
    expect(loadingWrapper.get('.timeline-metric-grid').classes()).toEqual(
      expect.arrayContaining(['progress-strip', 'metric-panel-grid'])
    )
    loadingWrapper.findAll('.timeline-metric-card').forEach((card) => {
      expect(card.classes()).toEqual(
        expect.arrayContaining([
          'progress-card',
          'metric-panel-card',
          'metric-panel-default-surface',
          'metric-panel-workspace-surface',
        ])
      )
    })
    expect(loadingWrapper.findAll('.timeline-metric-skeleton-label')).toHaveLength(4)
    expect(loadingWrapper.findAll('.timeline-event-item--loading')).toHaveLength(2)
    expect(loadingWrapper.text()).not.toContain('当前还没有训练动态。')
    expect(loadingWrapper.html()).not.toContain('workspace-glass-region')
    expect(loadingWrapper.html()).not.toContain('workspace-glass-metric-surface')

    expect(loadedWrapper.findAll('.timeline-event-item--loading')).toHaveLength(0)
    expect(loadedWrapper.text()).toContain('web-1')
    expect(loadedWrapper.text()).toContain('第 2 次提交命中 Flag')
    expect(loadedWrapper.text()).toContain('成功解题')
    expect(loadedWrapper.html()).not.toContain('workspace-glass-region')
    expect(loadedWrapper.html()).not.toContain('workspace-glass-metric-surface')
  })

  it('空列表只显示训练记录空状态，不回退到加载骨架', () => {
    const wrapper = mount(TrainingTimelineContent, {
      props: {
        timeline: [],
      },
    })

    expect(wrapper.find('.timeline-event-item--loading').exists()).toBe(false)
    expect(wrapper.text()).toContain('当前还没有训练动态。')
  })
})
