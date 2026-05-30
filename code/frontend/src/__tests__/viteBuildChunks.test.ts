import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

const viteConfigSource = readFileSync(`${process.cwd()}/vite.config.ts`, 'utf-8')

describe('vite build chunk split', () => {
  it('应该把 echarts 相关依赖拆成更细的独立 chunk', () => {
    expect(viteConfigSource).toContain("id.includes('/vue-echarts/')")
    expect(viteConfigSource).toContain("return 'vue-echarts'")
    expect(viteConfigSource).toContain("id.includes('/echarts/charts/')")
    expect(viteConfigSource).toContain("return 'echarts-charts'")
    expect(viteConfigSource).toContain("id.includes('/echarts/components/')")
    expect(viteConfigSource).toContain("return 'echarts-components'")
    expect(viteConfigSource).toContain("id.includes('/echarts/renderers/')")
    expect(viteConfigSource).toContain("return 'echarts-renderers'")
    expect(viteConfigSource).toContain("id.includes('/zrender/')")
    expect(viteConfigSource).toContain("return 'zrender'")
  })
})
