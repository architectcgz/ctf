import { describe, expect, it } from 'vitest'

import featureIndexSource from '../index.ts?raw'
import modelIndexSource from './index.ts?raw'

const files = import.meta.glob('../ui/*.vue', { query: '?raw', eager: true })
const allUiSources = Object.values(files).map((m: any) => m.default) as string[]

describe('contest-awd-config boundary', () => {
  describe('public API', () => {
    it('feature 顶层 index.ts 应 barrel export model + ui', () => {
      expect(featureIndexSource).toContain("export * from './model'")
      expect(featureIndexSource).toContain("export * from './ui'")
    })

    it('model/index.ts 应暴露 AWD 配置相关 composable', () => {
      expect(modelIndexSource).toContain('useContestAwdConfig')
    })
  })

  describe('layer isolation', () => {
    it('不应 import pages / widgets / vue-router / stores', () => {
      const allSources = [...allUiSources, modelIndexSource]
      for (const source of allSources) {
        expect(source).not.toMatch(/from\s+['"]@\/pages\//)
        expect(source).not.toMatch(/from\s+['"]@\/widgets\//)
        expect(source).not.toMatch(/from\s+['"]vue-router['"]/)
        expect(source).not.toMatch(/from\s+['"]@\/stores\//)
      }
    })
  })

  describe('owner boundaries', () => {
    it('UI 组件不应直接调 API', () => {
      for (const source of allUiSources) {
        expect(source).not.toMatch(/from\s+['"]@\/api\/(?!contracts)/)
      }
    })
  })
})
