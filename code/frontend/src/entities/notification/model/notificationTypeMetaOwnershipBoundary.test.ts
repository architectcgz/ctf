import { describe, expect, it } from 'vitest'

import entityTypesSource from './types.ts?raw'
import entityModelIndexSource from './index.ts?raw'

describe('NotificationTypeMeta ownership boundaries', () => {
  it('entities/notification/model/types.ts 应为 NotificationTypeMeta 的唯一真相源', () => {
    expect(entityTypesSource).toContain('export interface NotificationTypeMeta')
    expect(entityTypesSource).toContain('icon: Component')
    expect(entityTypesSource).toContain('label: string')
    expect(entityTypesSource).toContain('accentColor: string')
    expect(entityTypesSource).toContain('iconWrapStyle: Record<string, string>')
    expect(entityTypesSource).toContain('badgeStyle: Record<string, string>')
  })

  it('entities/notification/model/index.ts 应通过 barrel export 暴露 NotificationTypeMeta', () => {
    expect(entityModelIndexSource).toContain('NotificationTypeMeta')
  })
})
