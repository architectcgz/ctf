import { describe, expect, it } from 'vitest'

import {
  getContestAccentColor,
  getContestAccentVarStyle,
  getContestActionLabel,
  getContestModeLabel,
  getContestStatusBadgeClass,
  getContestStatusLabel,
} from './presentation'

describe('contest presentation', () => {
  it('maps contest statuses to stable labels, action labels, and badge classes', () => {
    expect(getContestStatusLabel('running')).toBe('进行中')
    expect(getContestStatusLabel('registering')).toBe('报名中')
    expect(getContestStatusLabel('ended')).toBe('已结束')

    expect(getContestActionLabel('running')).toBe('进入竞赛')
    expect(getContestActionLabel('registering')).toBe('立即报名')
    expect(getContestActionLabel('ended')).toBe('查看详情')

    expect(getContestStatusBadgeClass('running')).toBe('contest-status-pill--running')
    expect(getContestStatusBadgeClass('registering')).toBe('contest-status-pill--registering')
    expect(getContestStatusBadgeClass('draft')).toBe('contest-status-pill--draft')
    expect(getContestStatusBadgeClass('frozen')).toBe('contest-status-pill--frozen')
    expect(getContestStatusBadgeClass('ended')).toBe('contest-status-pill--ended')
    expect(getContestStatusBadgeClass('cancelled')).toBe('contest-status-pill--cancelled')
  })

  it('maps contest mode and accent colors through entity presentation owner', () => {
    expect(getContestModeLabel('jeopardy')).toBe('Jeopardy')
    expect(getContestModeLabel('awd')).toBe('AWD')

    expect(getContestAccentColor('running')).toBe('var(--color-primary)')
    expect(getContestAccentColor('registering')).toBe('var(--color-warning)')
    expect(getContestAccentColor('ended')).toContain('var(--color-text-muted)')
  })

  it('builds css var styles for contest accent consumers', () => {
    expect(getContestAccentVarStyle('running')).toEqual({
      '--contest-accent': 'var(--color-primary)',
    })
    expect(getContestAccentVarStyle('frozen', '--contest-row-accent')).toEqual({
      '--contest-row-accent': expect.any(String),
    })
  })
})
