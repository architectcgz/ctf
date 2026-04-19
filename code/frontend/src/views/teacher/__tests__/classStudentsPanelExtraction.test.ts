import { describe, expect, it } from 'vitest'

import classStudentsPageSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'

describe('class students panel extraction', () => {
  it('ClassStudentsPage 应将趋势复盘等互补视角下沉为独立入口，而不是继续在总览页直接渲染长内容面板', () => {
    expect(classStudentsPageSource).toContain('v-for="entry in workspaceEntries"')
    expect(classStudentsPageSource).toContain("@click=\"emit('openWorkspaceSection', entry.key)\"")
    expect(classStudentsPageSource).not.toContain('resolveWorkspacePanelComponent')
    expect(classStudentsPageSource).not.toContain('resolveWorkspacePanelProps')
    expect(classStudentsPageSource).not.toContain('<TeacherClassTrendPanel')
    expect(classStudentsPageSource).not.toContain('<TeacherClassReviewPanel')
    expect(classStudentsPageSource).not.toContain('<TeacherClassInsightsPanel')
    expect(classStudentsPageSource).not.toContain('<TeacherInterventionPanel')
  })
})
