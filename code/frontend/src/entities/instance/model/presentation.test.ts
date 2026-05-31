import { describe, expect, it } from 'vitest'

import {
  formatInstanceAccessDisplay,
  formatInstanceRemainingTime,
  getInstanceRemainingSeconds,
  getInstanceRemainingTone,
  getInstanceStatusDotClass,
  getInstanceStatusLabel,
  getInstanceStatusPillClass,
  getInstanceStatusTone,
  getInstanceStudentDisplayName,
  getInstanceStudentIdentityLabel,
  getInstanceStudentSecondaryLabel,
  getInstanceWaitingEtaLabel,
  getInstanceWaitingHint,
  getInstanceWaitingProgressLabel,
  getInstanceWaitingQueueLabel,
} from './presentation'

describe('instance presentation', () => {
  it('maps instance status labels and tones through entity owner', () => {
    expect(getInstanceStatusLabel('running')).toBe('运行中')
    expect(getInstanceStatusLabel('crashed')).toBe('运行异常')
    expect(getInstanceStatusLabel('unknown')).toBe('unknown')
    expect(getInstanceStatusLabel('')).toBe('--')
    expect(getInstanceStatusTone('running')).toBe('success')
    expect(getInstanceStatusTone('creating')).toBe('warning')
    expect(getInstanceStatusTone('failed')).toBe('danger')
    expect(getInstanceStatusTone('unknown')).toBe('muted')
    expect(getInstanceStatusDotClass('running')).toBe('instance-status-dot--success')
    expect(getInstanceStatusPillClass('creating')).toBe('instance-status-pill--pending')
  })

  it('formats remaining time and expired fallback consistently', () => {
    expect(getInstanceRemainingSeconds('2099-01-01T00:01:05Z', Date.parse('2099-01-01T00:00:00Z'))).toBe(
      65
    )
    expect(getInstanceRemainingTone(0)).toBe('muted')
    expect(getInstanceRemainingTone(240)).toBe('danger')
    expect(getInstanceRemainingTone(480)).toBe('warning')
    expect(getInstanceRemainingTone(1200)).toBe('success')
    expect(formatInstanceRemainingTime(65)).toBe('00:01:05')
    expect(formatInstanceRemainingTime(0)).toBe('00:00:00')
    expect(formatInstanceRemainingTime(-1, { expiredLabel: '已到期' })).toBe('已到期')
  })

  it('builds waiting hints for pending and failure states', () => {
    const waitingInstance = {
      status: 'pending' as const,
      queue_position: 2,
      eta_seconds: 90,
      progress: 35,
    }

    expect(getInstanceWaitingHint(waitingInstance)).toBe(
      '实例正在排队创建，队列第 2 位，预计等待 1 分 30 秒，进度 35%'
    )
    expect(getInstanceWaitingQueueLabel(waitingInstance)).toBe('当前排队：第 2 位')
    expect(getInstanceWaitingEtaLabel(waitingInstance)).toBe('预计等待：1 分 30 秒')
    expect(getInstanceWaitingProgressLabel(waitingInstance)).toBe('创建进度：35%')
    expect(getInstanceWaitingHint({ status: 'failed' })).toBe('启动失败，当前目标不可访问')
    expect(getInstanceWaitingHint({ status: 'running' })).toBe('')
    expect(getInstanceWaitingQueueLabel({ status: 'running' })).toBe('')
  })

  it('formats access and student meta through entity owner', () => {
    expect(
      formatInstanceAccessDisplay({
        access: { command: 'nc 127.0.0.1 10001' },
        access_url: 'http://example.test',
      })
    ).toBe('nc 127.0.0.1 10001')
    expect(
      formatInstanceAccessDisplay({
        ssh_info: { host: '127.0.0.1', port: 2222 },
      })
    ).toBe('127.0.0.1:2222')
    expect(getInstanceStudentDisplayName({ student_name: 'Alice', student_username: 'alice' })).toBe(
      'Alice'
    )
    expect(getInstanceStudentDisplayName({ student_username: 'alice' })).toBe('alice')
    expect(
      getInstanceStudentIdentityLabel({ student_no: 'S-1001', student_username: 'alice' })
    ).toBe('S-1001')
    expect(getInstanceStudentIdentityLabel({ student_username: 'alice' })).toBe('@alice')
    expect(
      getInstanceStudentSecondaryLabel({ student_username: 'alice', class_name: 'Class A' })
    ).toBe('@alice · Class A')
  })
})
