import { describe, expect, it } from 'vitest'

import {
  getUserDisplayName,
  getUserName,
  getUserUsername,
  getUserUsernameHandle,
} from './presentation'

describe('user presentation', () => {
  it('prefers stable display name, then username, then explicit fallback', () => {
    expect(getUserDisplayName({ name: 'Alice', username: 'alice' })).toBe('Alice')
    expect(getUserDisplayName({ name: '   ', username: 'alice' })).toBe('alice')
    expect(getUserDisplayName({ username: 'alice' })).toBe('alice')
    expect(getUserDisplayName(null, '选手')).toBe('选手')
  })

  it('returns raw username with fallback for directory consumers', () => {
    expect(getUserUsername({ username: 'alice' })).toBe('alice')
    expect(getUserUsername({ username: '   ' }, '未登录')).toBe('未登录')
  })

  it('returns only the real name for name-owned presentation slots', () => {
    expect(getUserName({ name: 'Alice', username: 'alice' })).toBe('Alice')
    expect(getUserName({ name: '   ', username: 'alice' }, '未设置姓名')).toBe('未设置姓名')
    expect(getUserName(null, '未设置姓名')).toBe('未设置姓名')
  })

  it('builds username handles through entity presentation owner', () => {
    expect(getUserUsernameHandle({ username: 'alice' })).toBe('@alice')
    expect(getUserUsernameHandle({ username: '' })).toBe('--')
    expect(getUserUsernameHandle(null, '未设置用户名')).toBe('未设置用户名')
  })
})
