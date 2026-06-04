import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import type { SkillProfileData } from '@/api/contracts'

import StudentInsightOverviewSection from './StudentInsightOverviewSection.vue'

const profile: SkillProfileData = {
  dimensions: [
    { key: 'web', name: 'Web', value: 82 },
    { key: 'crypto', name: '密码', value: 64 },
  ],
}

function mountOverview(loading = false) {
  return mount(StudentInsightOverviewSection, {
    props: {
      profile,
      loading,
    },
    global: {
      stubs: {
        SkillRadar: {
          template: '<div data-test="skill-radar">Radar</div>',
        },
      },
    },
  })
}

describe('StudentInsightOverviewSection', () => {
  it('loading 和 loaded 应共用同一套 SectionCard 外壳', () => {
    const loadingWrapper = mountOverview(true)
    const loadedWrapper = mountOverview(false)

    expect(loadingWrapper.findAll('.section-card')).toHaveLength(2)
    expect(loadedWrapper.findAll('.section-card')).toHaveLength(2)
    expect(loadingWrapper.findAll('.insight-overview-card')).toHaveLength(2)
    expect(loadedWrapper.findAll('.insight-overview-card')).toHaveLength(2)

    expect(loadingWrapper.text()).toContain('六维能力分布')
    expect(loadingWrapper.text()).toContain('维度得分占比')
    expect(loadingWrapper.find('.insight-overview-loading-radar').exists()).toBe(true)
    expect(loadingWrapper.findAll('.insight-overview-loading-bar-row')).toHaveLength(6)
    expect(loadingWrapper.find('[data-test="skill-radar"]').exists()).toBe(false)
    expect(loadingWrapper.html()).not.toContain('student-insight-glass-surface')

    expect(loadedWrapper.find('.insight-overview-loading-radar').exists()).toBe(false)
    expect(loadedWrapper.find('[data-test="skill-radar"]').exists()).toBe(true)
    expect(loadedWrapper.text()).toContain('Web')
    expect(loadedWrapper.text()).toContain('82%')
  })

  it('无画像维度时应在同一卡片结构内显示空状态', () => {
    const wrapper = mount(StudentInsightOverviewSection, {
      props: {
        profile: { dimensions: [] },
      },
      global: {
        stubs: {
          SkillRadar: {
            template: '<div data-test="skill-radar">Radar</div>',
          },
        },
      },
    })

    expect(wrapper.findAll('.section-card')).toHaveLength(2)
    expect(wrapper.text()).toContain('暂无画像维度数据')
  })
})
