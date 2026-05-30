import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const previewPath = resolve(process.cwd(), 'public/mockups/challenge-filter-patterns.html')
const challengeFilterPatternsSource = readFileSync(previewPath, 'utf8')

describe('ChallengeFilterPatternsMock', () => {
  it('提供三种可比对的筛选方案预览', () => {
    expect(challengeFilterPatternsSource).toContain('全量展示')
    expect(challengeFilterPatternsSource).toContain('主筛选折叠')
    expect(challengeFilterPatternsSource).toContain('摘要折叠')
    expect(challengeFilterPatternsSource).toContain('搜索题目标题')
    expect(challengeFilterPatternsSource).toContain('清空筛选')
  })

  it('包含可展开的主筛选折叠与摘要折叠原型', () => {
    expect(challengeFilterPatternsSource).toContain('id="progressive-toggle"')
    expect(challengeFilterPatternsSource).toContain('id="summary-toggle"')
    expect(challengeFilterPatternsSource).toContain('展开更多筛选')
    expect(challengeFilterPatternsSource).toContain('收起更多筛选')
    expect(challengeFilterPatternsSource).toContain('展开筛选')
    expect(challengeFilterPatternsSource).toContain('收起筛选')
    expect(challengeFilterPatternsSource).not.toContain('Primary First')
    expect(challengeFilterPatternsSource).not.toContain(
      '默认只保留高频条件，低频条件挂到可展开区域，适合字段逐步增长的目录页。'
    )
    expect(challengeFilterPatternsSource).not.toContain(
      '低频条件默认收起，需要时再展开，不占首屏高度。'
    )
    expect(challengeFilterPatternsSource).not.toContain('class="progressive-rail"')
    expect(challengeFilterPatternsSource).toContain('class="filter-actions__row"')
  })

  it('预览页应沿用平台工作台视觉骨架，而不是独立路由页面组件', () => {
    expect(challengeFilterPatternsSource).toContain('workspace-topbar')
    expect(challengeFilterPatternsSource).toContain('top-tabs')
    expect(challengeFilterPatternsSource).toContain('workspace-directory-section')
    expect(challengeFilterPatternsSource).toContain('journal-shell')
    expect(challengeFilterPatternsSource).toContain('mockups/challenge-filter-patterns.html')
  })
})
