import { describe, it, expect } from 'vitest'
import { useSanitize } from '../useSanitize'

describe('useSanitize', () => {
  it('应该清理恶意脚本', () => {
    const { sanitizeHtml } = useSanitize()
    const dirty = '<p>Hello</p><script>alert("xss")</script>'
    const clean = sanitizeHtml(dirty)

    expect(clean).toContain('<p>Hello</p>')
    expect(clean).not.toContain('<script>')
    expect(clean).not.toContain('alert')
  })

  it('应该保留允许的标签', () => {
    const { sanitizeHtml } = useSanitize()
    const html = '<p>Text</p><strong>Bold</strong><a href="/link">Link</a>'
    const clean = sanitizeHtml(html)

    expect(clean).toContain('<p>')
    expect(clean).toContain('<strong>')
    expect(clean).toContain('<a')
  })

  it('应该移除不允许的属性', () => {
    const { sanitizeHtml } = useSanitize()
    const html = '<p onclick="alert()">Text</p>'
    const clean = sanitizeHtml(html)

    expect(clean).not.toContain('onclick')
  })
})
