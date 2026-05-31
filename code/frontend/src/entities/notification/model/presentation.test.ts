import { describe, expect, it } from 'vitest'

import {
  getNotificationAccentColor,
  getNotificationReadStateLabel,
  getNotificationTypeAccent,
  getNotificationTypeAccentColor,
  getNotificationTypeLabel,
} from './presentation'

describe('notification presentation', () => {
  it('maps notification types to stable labels and accents', () => {
    expect(getNotificationTypeLabel('system')).toBe('系统')
    expect(getNotificationTypeLabel('contest')).toBe('竞赛')
    expect(getNotificationTypeLabel('challenge')).toBe('训练')
    expect(getNotificationTypeLabel('team')).toBe('团队')

    expect(getNotificationTypeAccent('system')).toBe('primary')
    expect(getNotificationTypeAccent('contest')).toBe('warning')
    expect(getNotificationTypeAccent('challenge')).toBe('success')
    expect(getNotificationTypeAccent('team')).toBe('violet')
  })

  it('falls back to system presentation for unknown notification types', () => {
    expect(getNotificationTypeLabel('other')).toBe('系统')
    expect(getNotificationTypeAccent('other')).toBe('primary')
    expect(getNotificationTypeAccentColor('other')).toBe('var(--color-primary)')
  })

  it('exposes stable accent colors and read-state labels', () => {
    expect(getNotificationAccentColor('warning')).toBe('var(--color-warning)')
    expect(getNotificationAccentColor('success')).toBe('var(--color-success)')
    expect(getNotificationReadStateLabel(true)).toBe('未读')
    expect(getNotificationReadStateLabel(false)).toBe('已读')
  })
})
