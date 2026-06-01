import { describe, expect, it } from 'vitest'

import featureIndexSource from '../index.ts?raw'
import modelIndexSource from './index.ts?raw'
import studioPageSource from './useChallengeTopologyStudioPage.ts?raw'
import draftSource from './topologyDraft.ts?raw'
import layoutSource from './topologyLayout.ts?raw'

describe('challenge-topology-studio boundary', () => {
  describe('public API', () => {
    it('feature 顶层 index.ts 应 barrel export model + ui', () => {
      expect(featureIndexSource).toContain("export * from './model'")
      expect(featureIndexSource).toContain("export * from './ui'")
    })

    it('model/index.ts 应暴露 12 个 composable + 2 个工具模块', () => {
      expect(modelIndexSource).toContain('useChallengeTopologyStudioPage')
      expect(modelIndexSource).toContain('useTopologyCanvasActions')
      expect(modelIndexSource).toContain('useTopologyDataLoader')
      expect(modelIndexSource).toContain('useTopologyPersistenceActions')
      expect(modelIndexSource).toContain('useTopologyStructureMutations')
      expect(modelIndexSource).toContain('useTopologyTemplateApply')
      expect(modelIndexSource).toContain('useTopologyTemplateMutations')
      expect(modelIndexSource).toContain("export * from './topologyDraft'")
      expect(modelIndexSource).toContain("export * from './topologyLayout'")
    })
  })

  describe('layer isolation', () => {
    const allSources = [studioPageSource, draftSource, layoutSource]

    it('model 层不应 import pages 或 widgets', () => {
      for (const source of allSources) {
        expect(source).not.toMatch(/from\s+['"]@\/pages\//)
        expect(source).not.toMatch(/from\s+['"]@\/widgets\//)
      }
    })

    it('model 层不应直接 import vue-router', () => {
      for (const source of allSources) {
        expect(source).not.toMatch(/from\s+['"]vue-router['"]/)
        expect(source).not.toMatch(/\buseRouter\s*\(/)
      }
    })
  })

  describe('owner boundaries', () => {
    it('useChallengeTopologyStudioPage 应组合子模块，不应内联 topology 构建与布局逻辑', () => {
      // 从子模块引入
      expect(studioPageSource).toContain("from './topologyDraft'")
      expect(studioPageSource).toContain("from './topologyLayout'")
      expect(studioPageSource).toContain("from './useTopologyCanvasActions'")
      expect(studioPageSource).toContain("from './useTopologyPersistenceActions'")
      // 不应内联
      expect(studioPageSource).not.toContain('function buildTopologyCanvasGraph(')
      expect(studioPageSource).not.toContain('function createEmptyTopologyDraft(')
    })

    it('topologyDraft 应收口 topology 数据结构与校验，不应处理 canvas 布局', () => {
      expect(draftSource).toContain('export function createDraftFromTopology')
      expect(draftSource).toContain('export function buildTopologyDraftValidationIssues')
      expect(draftSource).toContain('export function serializeTopologyDraft')
      expect(draftSource).not.toContain('CanvasNodePosition')
      expect(draftSource).not.toContain('buildTopologyCanvasGraph')
    })

    it('topologyLayout 应收口 canvas 图结构与节点位置计算，不应处理数据加载或持久化', () => {
      expect(layoutSource).toContain('export function buildTopologyCanvasGraph')
      expect(layoutSource).not.toContain('from \'@/api/')
      expect(layoutSource).not.toContain('saveChallengeTopology')
    })
  })

  describe('API access', () => {
    it('data loader / persistence / draft 可调 admin/authoring API，符合 feature model 层规则', () => {
      // topologyDraft 包含 API payload 类型转换，允许 import API
      expect(draftSource).toContain("from '@/api/admin/authoring'")
    })

    it('topologyLayout 应为纯函数模块，不应依赖 API 或异步 I/O', () => {
      expect(layoutSource).not.toMatch(/from\s+['"]@\/api\//)
      expect(layoutSource).not.toContain('async function')
    })
  })
})
