import { describe, expect, it } from 'vitest'

import challengeDetailSource from '../ChallengeDetail.vue?raw'
import challengeInstanceCardSource from '@/components/challenge/ChallengeInstanceCard.vue?raw'

describe('challenge detail shared shell alignment', () => {
  it('通过变量接入共享 workspace shell，而不是继续本地重写整套壳层样式', () => {
    expect(challengeDetailSource).toContain('--workspace-shell-border: var(--journal-line-soft);')
    expect(challengeDetailSource).toContain('--workspace-shell-bg: var(--bg-shell);')
    expect(challengeDetailSource).toContain('--workspace-shadow-shell: var(--journal-shadow);')
    expect(challengeDetailSource).toContain('min-height: max(100%, calc(100vh - 5rem));')
    expect(challengeDetailSource).not.toMatch(
      /\.workspace-shell\s*\{[\s\S]*border:\s*1px solid var\(--journal-line-soft\);/s
    )
    expect(challengeDetailSource).not.toMatch(
      /\.workspace-shell\s*\{[\s\S]*box-shadow:\s*var\(--journal-shadow\);/s
    )
    expect(challengeDetailSource).not.toMatch(
      /:global\(\[data-theme='dark'\]\) \.workspace-shell\s*\{[^}]*background:/s
    )
    expect(challengeDetailSource).toContain('--workspace-shell-radial-strength: 14%;')
    expect(challengeDetailSource).toContain('--workspace-shell-radial-size: 24rem;')
  })

  it('主标签轨道应通过 page-tabs 变量接入共享规则，而不是继续本地声明 top-tabs/top-tab', () => {
    expect(challengeDetailSource).toContain('--page-top-tabs-gap: var(--space-7);')
    expect(challengeDetailSource).toContain('--page-top-tabs-padding: 0 var(--space-7);')
    expect(challengeDetailSource).toContain('--page-top-tab-active-border: var(--brand);')
    expect(challengeDetailSource).not.toMatch(/^\.top-tabs\s*,/m)
    expect(challengeDetailSource).not.toMatch(/^\.top-tab\s*,/m)
  })

  it('题目详情的 tab 与操作按钮应复用共享主题按钮和页签栈', () => {
    expect(challengeDetailSource).toContain('class="workspace-tab top-tab"')
    expect(challengeDetailSource).toContain(":class=\"{ active: activeWorkspaceTab === tab.id }\"")
    expect(challengeDetailSource).toContain('class="solution-tabbar top-tabs challenge-subtabs"')
    expect(challengeDetailSource).toContain('class="solution-tab top-tab challenge-subtab"')
    expect(challengeDetailSource).toContain('class="challenge-btn"')
    expect(challengeDetailSource).toContain(
      'class="challenge-btn challenge-btn-primary disabled:cursor-not-allowed disabled:opacity-50"'
    )
    expect(challengeDetailSource).not.toContain('workspace-tab--active')
    expect(challengeDetailSource).not.toMatch(/^\.sub-tabs\s*,/m)
    expect(challengeDetailSource).not.toMatch(/^\.sub-tab\s*,/m)
    expect(challengeDetailSource).not.toMatch(/\.primary-action\s*,/s)
    expect(challengeDetailSource).not.toMatch(/\.ghost-action\s*,/s)
    expect(challengeDetailSource).not.toMatch(/\.subtle-action\s*\{/s)

    expect(challengeInstanceCardSource).toContain(
      'class="instance-btn instance-btn-primary disabled:cursor-not-allowed disabled:opacity-50"'
    )
    expect(challengeInstanceCardSource).toContain(
      'class="instance-btn disabled:cursor-not-allowed disabled:opacity-50"'
    )
    expect(challengeInstanceCardSource).not.toMatch(/\.primary-action\s*,/s)
    expect(challengeInstanceCardSource).not.toMatch(/\.subtle-action\s*\{/s)
  })

  it('题目头部主信息块与右侧提交按钮应从统一主题主色链取色', () => {
    expect(challengeDetailSource).toContain('class="question-hero-main"')
    expect(challengeDetailSource).toContain('--brand: var(--color-primary);')
    expect(challengeDetailSource).toContain(
      '--brand-soft: color-mix(in srgb, var(--color-primary) 10%, transparent);'
    )
    expect(challengeDetailSource).toContain(
      '--brand-ink: color-mix(in srgb, var(--color-primary) 78%, var(--text-main));'
    )
    expect(challengeDetailSource).toContain(
      'class="challenge-btn challenge-btn-primary disabled:cursor-not-allowed disabled:opacity-50"'
    )
  })

  it('题目详情右侧工具栏与实例卡应避免浅色混入，并在夜间模式切到 dark color-scheme', () => {
    expect(challengeDetailSource).toMatch(
      /\.tool-pane\s*\{[\s\S]*background:\s*linear-gradient\([\s\S]*var\(--bg-panel\)[\s\S]*var\(--color-bg-base\)[\s\S]*var\(--bg-shell\)[\s\S]*var\(--color-bg-base\)[\s\S]*\);/s
    )
    expect(challengeDetailSource).toMatch(
      /\.challenge-prose :deep\(pre\)\s*\{[\s\S]*background:\s*color-mix\(in srgb,\s*var\(--bg-panel\)\s*72%,\s*var\(--color-bg-base\)\);/s
    )
    expect(challengeDetailSource).toMatch(
      /\.challenge-input,[\s\S]*\.flag-input\s*\{[\s\S]*background:\s*var\(--bg-panel\);/s
    )
    expect(challengeDetailSource).toMatch(
      /:global\(\[data-theme='dark'\]\) \.journal-shell\s*\{[\s\S]*color-scheme:\s*dark;/s
    )
    expect(challengeDetailSource).not.toContain('color-mix(in srgb, var(--bg-panel) 95%, white)')
    expect(challengeDetailSource).not.toContain('color-mix(in srgb, var(--bg-shell) 92%, white)')
    expect(challengeDetailSource).not.toContain('color-mix(in srgb, var(--bg-panel) 72%, white)')
    expect(challengeDetailSource).not.toContain('background: white;')

    expect(challengeInstanceCardSource).toMatch(
      /:global\(\[data-theme='dark'\]\) \.instance-shell\s*\{[\s\S]*--brand:\s*color-mix\(in srgb,\s*var\(--color-primary\)\s*88%,\s*var\(--color-text-primary\)\);/s
    )
    expect(challengeInstanceCardSource).not.toContain("var(--color-primary) 88%, white")
  })

  it('题目详情夜间模式应覆盖 workspace page 底色，避免外层主内容区继续发亮', () => {
    expect(challengeDetailSource).toMatch(
      /:global\(\[data-theme='dark'\]\) \.journal-shell\s*\{[\s\S]*--bg-page:\s*color-mix\(in srgb,\s*var\(--color-bg-base\)\s*94%,\s*var\(--color-bg-surface\)\);/s
    )
  })
})
