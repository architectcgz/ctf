import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import ChallengePackageImportReview from '../challenge/ChallengePackageImportReview.vue'
import challengePackageImportReviewSource from '../challenge/ChallengePackageImportReview.vue?raw'
import challengeDescriptionPanelSource from '../challenge/ChallengeDescriptionPanel.vue?raw'
import type { AdminChallengeImportPreview } from '@/api/contracts'

const preview: AdminChallengeImportPreview = {
  id: 'import-1',
  file_name: 'demo.zip',
  slug: 'demo-challenge',
  title: 'Demo Challenge',
  description: '# 一级标题\n\n- item-1\n\n<script>alert(1)</script>',
  category: 'web',
  difficulty: 'easy',
  points: 100,
  attachments: [],
  hints: [{ level: 1, title: 'Hint 1', content: '先检查入口参数。' }],
  flag: { type: 'static', prefix: 'flag' },
  runtime: { type: 'container', image_ref: 'ctf/demo:latest' },
  extensions: { topology: { source: '', enabled: false } },
  warnings: [],
  created_at: '2026-04-09T08:00:00.000Z',
}

describe('ChallengePackageImportReview', () => {
  it('应将题面 markdown 渲染为 HTML 并过滤危险标签', () => {
    const wrapper = mount(ChallengePackageImportReview, {
      props: {
        preview,
        committing: false,
      },
    })

    expect(wrapper.text()).toContain('题头')
    expect(wrapper.find('.import-review__statement-label').exists()).toBe(true)
    expect(wrapper.find('.import-review__statement .import-review__statement-label').exists()).toBe(
      false
    )

    const description = wrapper.get('[data-testid="import-review-description"]')
    const html = description.html()
    expect(html).toContain('<h1')
    expect(html).toContain('<li>item-1</li>')
    expect(html).not.toContain('<script')
    expect(wrapper.text()).not.toContain('# 一级标题')

    const sections = wrapper.findAll('article.import-review__section')
    expect(sections[0]?.text()).toContain('Runtime')
    expect(sections[0]?.text()).toContain('Topology')
    expect(sections[0]?.find('.import-review__overview > .import-review__runtime > dl').exists()).toBe(
      true
    )
    expect(sections[0]?.find('.import-review__definition--runtime').exists()).toBe(false)
    expect(sections[1]?.text()).toContain('提示')
    expect(sections[1]?.find('dl').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('提示系统')
  })

  it('题面区域应保持固定高度滚动容器，避免长题面破坏布局', () => {
    expect(challengeDescriptionPanelSource).toMatch(
      /\.import-review__statement\s*\{[^}]*max-height:\s*clamp\(15rem,\s*36vh,\s*24rem\);[^}]*overflow:\s*auto;/s
    )
    expect(challengePackageImportReviewSource).toMatch(
      /\.import-review__grid\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\);/s
    )
    expect(challengePackageImportReviewSource).toContain('<ChallengeDescriptionPanel')
  })
})
