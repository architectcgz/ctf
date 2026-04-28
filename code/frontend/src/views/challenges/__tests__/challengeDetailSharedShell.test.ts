import { describe, expect, it } from 'vitest'

import challengeDetailSource from '../ChallengeDetail.vue?raw'
import challengeActionAsideSource from '@/components/challenge/ChallengeActionAside.vue?raw'
import challengeInstanceCardSource from '@/components/challenge/ChallengeInstanceCard.vue?raw'
import challengeQuestionPanelSource from '@/components/challenge/ChallengeQuestionPanel.vue?raw'
import challengeSolutionsPanelSource from '@/components/challenge/ChallengeSolutionsPanel.vue?raw'
import challengeWriteupPanelSource from '@/components/challenge/ChallengeWriteupPanel.vue?raw'

const challengeDetailWorkspaceSource = [
  challengeDetailSource,
  challengeActionAsideSource,
  challengeQuestionPanelSource,
  challengeSolutionsPanelSource,
  challengeWriteupPanelSource,
].join('\n')

describe('challenge detail shared shell alignment', () => {
  it('题目详情页应通过 journal-shell-user 接入共享学生侧 shell', () => {
    expect(challengeDetailSource).toContain(
      'class="journal-shell journal-shell-user journal-hero workspace-shell min-h-full"'
    )
  })

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
    expect(challengeDetailSource).toContain('class="workspace-tabbar top-tabs"')
    expect(challengeDetailSource).not.toContain('--page-top-tabs-gap: var(--space-7);')
    expect(challengeDetailSource).not.toContain('--page-top-tabs-padding: 0 var(--space-7);')
    expect(challengeDetailSource).not.toContain('--page-top-tab-active-border: var(--brand);')
    expect(challengeDetailSource).not.toMatch(/^\.top-tabs\s*,/m)
    expect(challengeDetailSource).not.toMatch(/^\.top-tab\s*,/m)
  })

  it('题目详情的 tab 与操作按钮应复用共享主题按钮和页签栈', () => {
    expect(challengeDetailSource).toContain('class="workspace-tab top-tab"')
    expect(challengeDetailSource).toContain(':class="{ active: activeWorkspaceTab === tab.id }"')
    expect(challengeDetailWorkspaceSource).toContain('class="solution-tabbar top-tabs challenge-subtabs"')
    expect(challengeDetailWorkspaceSource).toContain('class="solution-tab top-tab challenge-subtab"')
    expect(challengeDetailWorkspaceSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(challengeDetailWorkspaceSource).toContain('class="ui-btn ui-btn--sm ui-btn--ghost hint-toggle"')
    expect(challengeDetailWorkspaceSource).toContain(
      'class="ui-btn ui-btn--primary disabled:cursor-not-allowed disabled:opacity-50"'
    )
    expect(challengeDetailWorkspaceSource).toContain('class="ui-control-wrap"')
    expect(challengeDetailWorkspaceSource).toContain('class="ui-control challenge-input"')
    expect(challengeDetailWorkspaceSource).toContain('class="ui-control-wrap writeup-textarea-wrap"')
    expect(challengeDetailWorkspaceSource).toContain('class="ui-control-wrap flag-input-wrap"')
    expect(challengeDetailWorkspaceSource).not.toContain('workspace-tab--active')
    expect(challengeDetailWorkspaceSource).not.toMatch(/^\.sub-tabs\s*,/m)
    expect(challengeDetailWorkspaceSource).not.toMatch(/^\.sub-tab\s*,/m)
    expect(challengeDetailWorkspaceSource).not.toMatch(/\.primary-action\s*,/s)
    expect(challengeDetailWorkspaceSource).not.toMatch(/\.ghost-action\s*,/s)
    expect(challengeDetailWorkspaceSource).not.toMatch(/\.subtle-action\s*\{/s)

    expect(challengeInstanceCardSource).toContain(
      'class="ui-btn ui-btn--primary disabled:cursor-not-allowed disabled:opacity-50"'
    )
    expect(challengeInstanceCardSource).toContain(
      'class="ui-btn ui-btn--secondary disabled:cursor-not-allowed disabled:opacity-50"'
    )
    expect(challengeInstanceCardSource).toContain(
      'class="ui-btn ui-btn--danger disabled:cursor-not-allowed disabled:opacity-50"'
    )
    expect(challengeInstanceCardSource).not.toContain('class="instance-btn')
    expect(challengeInstanceCardSource).not.toMatch(/\.primary-action\s*,/s)
    expect(challengeInstanceCardSource).not.toMatch(/\.subtle-action\s*\{/s)
    expect(challengeInstanceCardSource).toContain('instance-status-text--success')
    expect(challengeInstanceCardSource).not.toContain('text-[var(--color-success)]')
    expect(challengeInstanceCardSource).not.toContain('text-[var(--color-warning)]')
    expect(challengeInstanceCardSource).not.toContain('text-[var(--color-danger)]')
  })

  it('题目头部主信息块与右侧提交按钮应从统一主题主色链取色', () => {
    expect(challengeDetailWorkspaceSource).toContain('class="question-hero-main"')
    expect(challengeDetailSource).toContain('--brand: var(--journal-accent);')
    expect(challengeDetailSource).toContain(
      '--brand-soft: color-mix(in srgb, var(--journal-accent) 10%, transparent);'
    )
    expect(challengeDetailSource).toContain('--brand-ink: var(--journal-accent-strong);')
    expect(challengeDetailWorkspaceSource).toContain(
      'class="ui-btn ui-btn--primary disabled:cursor-not-allowed disabled:opacity-50"'
    )
  })

  it('题目详情右侧工具栏与实例卡应避免浅色混入，并在夜间模式切到 dark color-scheme', () => {
    expect(challengeDetailSource).toMatch(
      /\.tool-pane\s*\{[\s\S]*background:\s*linear-gradient\([\s\S]*var\(--bg-panel\)[\s\S]*var\(--color-bg-base\)[\s\S]*var\(--bg-shell\)[\s\S]*var\(--color-bg-base\)[\s\S]*\);/s
    )
    expect(challengeDetailWorkspaceSource).toMatch(
      /\.challenge-prose :deep\(pre\)\s*\{[\s\S]*background:\s*color-mix\(in srgb,\s*var\(--bg-panel\)\s*72%,\s*var\(--color-bg-base\)\);/s
    )
    expect(challengeDetailWorkspaceSource).toMatch(
      /\.field \.ui-control-wrap\s*\{[\s\S]*background:\s*var\(--bg-panel\);/s
    )
    expect(challengeDetailWorkspaceSource).toMatch(
      /\.flag-input-wrap\s*\{[\s\S]*background:\s*var\(--bg-panel\);/s
    )
    expect(challengeDetailWorkspaceSource).not.toContain(
      'color-mix(in srgb, var(--bg-panel) 95%, white)'
    )
    expect(challengeDetailWorkspaceSource).not.toContain(
      'color-mix(in srgb, var(--bg-shell) 92%, white)'
    )
    expect(challengeDetailWorkspaceSource).not.toContain(
      'color-mix(in srgb, var(--bg-panel) 72%, white)'
    )
    expect(challengeDetailWorkspaceSource).not.toContain('background: white;')

    expect(challengeInstanceCardSource).toMatch(
      /:global\(\[data-theme='dark'\]\) \.instance-shell\s*\{[\s\S]*--brand:\s*color-mix\(in srgb,\s*var\(--color-primary\)\s*88%,\s*var\(--color-text-primary\)\);/s
    )
    expect(challengeInstanceCardSource).not.toContain('var(--color-primary) 88%, white')
  })

  it('题目详情夜间模式应覆盖 workspace page 底色，避免外层主内容区继续发亮', () => {
    expect(challengeDetailSource).toContain(
      '--bg-page: color-mix(in srgb, var(--color-bg-base) 94%, var(--color-bg-surface));'
    )
    expect(challengeDetailSource).not.toMatch(
      /^:global\(\[data-theme='dark'\]\) \.journal-shell\s*\{/m
    )
  })
})
