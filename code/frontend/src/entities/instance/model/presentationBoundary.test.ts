import { describe, expect, it } from 'vitest'

import presentationSource from './presentation.ts?raw'
import modelIndexSource from './index.ts?raw'
import entityIndexSource from '../index.ts?raw'

describe('instance entity presentation boundaries', () => {
  describe('status mapping ownership', () => {
    it('状态文案映射应收口在 entities/instance，不在 features 中重复定义', () => {
      expect(presentationSource).toContain('const instanceStatusLabels')
      expect(presentationSource).toContain("pending: '等待中'")
      expect(presentationSource).toContain("running: '运行中'")
      expect(presentationSource).toContain("failed: '启动失败'")
      expect(presentationSource).toContain("crashed: '运行异常'")
    })

    it('状态 tone 映射应收口在 entities/instance', () => {
      expect(presentationSource).toContain('const instanceStatusTones')
      expect(presentationSource).toContain('running: \'success\'')
      expect(presentationSource).toContain('failed: \'danger\'')
    })

    it('所有状态展示规则应通过 barrelexport 暴露', () => {
      expect(modelIndexSource).toContain("export {")
      expect(modelIndexSource).toContain('getInstanceStatusLabel')
      expect(modelIndexSource).toContain('getInstanceStatusTone')
      expect(modelIndexSource).toContain('getInstanceStatusDotClass')
      expect(modelIndexSource).toContain('getInstanceStatusPillClass')
      expect(entityIndexSource).toContain("export * from './model'")
    })
  })

  describe('remaining time ownership', () => {
    it('剩余时间计算应收口在 entities/instance，不在页面或 widget 中内联', () => {
      expect(presentationSource).toContain('export function getInstanceRemainingSeconds')
      expect(presentationSource).toContain('export function getInstanceRemainingTone')
      expect(presentationSource).toContain('export function formatInstanceRemainingTime')
    })
  })

  describe('waiting state ownership', () => {
    it('排队/等待文案应收口在 entities/instance', () => {
      expect(presentationSource).toContain('export function getInstanceWaitingHint')
      expect(presentationSource).toContain('export function getInstanceWaitingQueueLabel')
      expect(presentationSource).toContain('export function getInstanceWaitingEtaLabel')
      expect(presentationSource).toContain('export function getInstanceWaitingProgressLabel')
    })
  })

  describe('student display ownership', () => {
    it('instance 相关的学生展示应收口在 entities/instance，复用 entities/user', () => {
      expect(presentationSource).toContain("import { getUserDisplayName, getUserUsernameHandle } from '@/entities/user'")
      expect(presentationSource).toContain('export function getInstanceStudentDisplayName')
      expect(presentationSource).toContain('export function getInstanceStudentIdentityLabel')
      expect(presentationSource).toContain('export function getInstanceStudentSecondaryLabel')
    })
  })

  describe('layer isolation', () => {
    it('entities/instance 不应依赖 features/widgets/pages', () => {
      expect(presentationSource).not.toMatch(/from\s+['"]@\/features\//)
      expect(presentationSource).not.toMatch(/from\s+['"]@\/widgets\//)
      expect(presentationSource).not.toMatch(/from\s+['"]@\/pages\//)
    })

    it('entities/instance 不应直接 import vue-router', () => {
      expect(presentationSource).not.toMatch(/from\s+['"]vue-router['"]/)
    })
  })
})
