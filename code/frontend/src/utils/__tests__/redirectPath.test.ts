import { describe, expect, it } from 'vitest'

import { sanitizeRedirectPath } from '../redirectPath'

describe('sanitizeRedirectPath', () => {
  it('应阻止 open redirect 路径', () => {
    expect(sanitizeRedirectPath('//evil.com')).toBe('/')
    expect(sanitizeRedirectPath('https://evil.com/phish')).toBe('/')
  })

  it('应拒绝 legacy 教师端前端页面路径', () => {
    expect(sanitizeRedirectPath('/teacher/dashboard')).toBe('/')
    expect(sanitizeRedirectPath('/teacher/instances?tab=running#lab')).toBe('/')
  })

  it('应拒绝 dynamic legacy 教师端前端页面路径', () => {
    expect(
      sanitizeRedirectPath('/teacher/classes/class-a/students/student-1/review-archive?tab=all#top')
    ).toBe('/')
  })

  it('非 legacy 前端路径应保持原样', () => {
    expect(sanitizeRedirectPath('/contests/1')).toBe('/contests/1')
  })
})
