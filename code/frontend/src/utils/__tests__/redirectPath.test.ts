import { describe, expect, it } from 'vitest'

import { sanitizeRedirectPath } from '../redirectPath'

describe('sanitizeRedirectPath', () => {
  it('应阻止 open redirect 路径', () => {
    expect(sanitizeRedirectPath('//evil.com')).toBe('/')
    expect(sanitizeRedirectPath('https://evil.com/phish')).toBe('/')
  })

  it('应把 legacy 教师端路径归一到 academy 命名空间', () => {
    expect(sanitizeRedirectPath('/teacher/dashboard')).toBe('/academy/overview')
    expect(sanitizeRedirectPath('/teacher/instances?tab=running#lab')).toBe(
      '/academy/instances?tab=running#lab'
    )
  })

  it('应保留 dynamic legacy 教师端路径的参数与后缀', () => {
    expect(
      sanitizeRedirectPath('/teacher/classes/class-a/students/student-1/review-archive?tab=all#top')
    ).toBe('/academy/classes/class-a/students/student-1/review-archive?tab=all#top')
  })

  it('非 legacy 前端路径应保持原样', () => {
    expect(sanitizeRedirectPath('/contests/1')).toBe('/contests/1')
  })
})
