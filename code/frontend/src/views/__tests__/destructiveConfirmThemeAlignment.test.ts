import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

const sharedStylesSource = readFileSync(`${process.cwd()}/src/style.css`, 'utf-8')

describe('destructive confirm theme alignment', () => {
  it('危险确认框应通过主题 token 适配深浅色，而不是继续写死旧颜色回退', () => {
    expect(sharedStylesSource).toContain('.app-destructive-confirm-box.el-message-box {')
    expect(sharedStylesSource).toContain(
      '--destructive-confirm-accent: var(--color-danger);'
    )
    expect(sharedStylesSource).toMatch(
      /--destructive-confirm-border:\s*color-mix\([\s\S]*var\(--destructive-confirm-accent\) 24%,[\s\S]*var\(--journal-border,\s*var\(--color-border-default\)\)[\s\S]*\);/
    )
    expect(sharedStylesSource).toMatch(
      /--destructive-confirm-soft-bg:\s*color-mix\([\s\S]*var\(--destructive-confirm-accent\) 10%,[\s\S]*var\(--journal-surface,\s*var\(--color-bg-surface\)\)[\s\S]*\);/
    )
    expect(sharedStylesSource).toMatch(
      /--destructive-confirm-soft-text:\s*color-mix\([\s\S]*var\(--destructive-confirm-accent\) 88%,[\s\S]*var\(--journal-ink,\s*var\(--color-text-primary\)\)[\s\S]*\);/
    )
    expect(sharedStylesSource).not.toContain('#5f6f83')
    expect(sharedStylesSource).not.toContain('#a4b2c2')
    expect(sharedStylesSource).not.toContain('#b91c1c')
    expect(sharedStylesSource).not.toContain('#dc2626')
    expect(sharedStylesSource).not.toMatch(
      /\[data-theme="light"\] \.app-destructive-confirm-box \.el-message-box__btns/
    )
  })
})
