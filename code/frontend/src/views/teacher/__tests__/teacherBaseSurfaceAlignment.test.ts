import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import studentManagementSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'
import dashboardSource from '@/components/teacher/dashboard/TeacherDashboardPage.vue?raw'
import instanceManagementSource from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue?raw'
import awdReviewIndexSource from '@/views/teacher/TeacherAWDReviewIndex.vue?raw'
import awdReviewDetailSource from '@/views/teacher/TeacherAWDReviewDetail.vue?raw'

const teacherSurfaceSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-surface.css`,
  'utf-8'
)

describe('teacher base surface alignment', () => {
  it('teacher base pages should soften control, card, and divider borders', () => {
    expect(teacherSurfaceSource).toMatch(
      /\.teacher-btn\s*\{[\s\S]*border:\s*1px solid var\(--teacher-control-border\);/s
    )
    expect(teacherSurfaceSource).toContain('.teacher-management-shell {')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-hero')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-summary')

    expect(classManagementSource).toContain('teacher-management-shell')
    expect(classManagementSource).not.toContain('.teacher-btn {')
    expect(classManagementSource).not.toMatch(/^\.teacher-hero\s*\{/m)
    expect(classManagementSource).not.toMatch(/^\.teacher-summary\s*\{/m)
    expect(classManagementSource).toMatch(
      /\.teacher-badge-card\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )
    expect(classManagementSource).toMatch(
      /\.teacher-tip-block\s*\{[\s\S]*border-top:\s*1px dashed var\(--teacher-divider\);/s
    )

    expect(studentManagementSource).toContain('teacher-management-shell')
    expect(studentManagementSource).not.toContain('.teacher-btn {')
    expect(studentManagementSource).not.toMatch(/^\.teacher-hero\s*\{/m)
    expect(studentManagementSource).not.toMatch(/^\.teacher-summary\s*\{/m)
    expect(studentManagementSource).toMatch(
      /\.teacher-badge-card\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )

    expect(dashboardSource).toContain('--teacher-card-border:')
    expect(dashboardSource).toContain('--teacher-control-border:')
    expect(dashboardSource).toContain('--teacher-divider:')
    expect(dashboardSource).toMatch(
      /\.teacher-btn\s*\{[\s\S]*border:\s*1px solid var\(--teacher-control-border\);/s
    )
    expect(dashboardSource).toMatch(
      /\.teacher-badge-card\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )
    expect(dashboardSource).toMatch(
      /\.teacher-tip-block\s*\{[\s\S]*border-top:\s*1px dashed var\(--teacher-divider\);/s
    )

    expect(instanceManagementSource).toContain('teacher-management-shell')
    expect(instanceManagementSource).not.toContain('.teacher-btn {')
    expect(instanceManagementSource).toContain(
      '--teacher-management-hero-border: var(--teacher-card-border);'
    )
    expect(instanceManagementSource).not.toMatch(/^\.teacher-hero\s*\{/m)
    expect(instanceManagementSource).not.toMatch(/^\.teacher-summary\s*\{/m)
    expect(instanceManagementSource).toMatch(
      /\.teacher-badge-card\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )
    expect(instanceManagementSource).toMatch(
      /\.teacher-tip-block\s*\{[\s\S]*border-top:\s*1px dashed var\(--teacher-divider\);/s
    )
  })

  it('teacher summary cards should explicitly adopt metric-panel classes and rely on shared variables', () => {
    expect(classManagementSource).toContain('class="teacher-summary metric-panel-default-surface"')
    expect(classManagementSource).toContain(
      'class="teacher-summary-grid progress-strip metric-panel-grid"'
    )
    expect(classManagementSource).toContain('class="progress-card metric-panel-card"')
    expect(classManagementSource).not.toContain(
      'class="teacher-summary-item progress-card metric-panel-card"'
    )
    expect(classManagementSource).toContain('class="progress-card-label metric-panel-label"')
    expect(classManagementSource).not.toContain(
      'class="teacher-summary-label progress-card-label metric-panel-label"'
    )
    expect(classManagementSource).toContain('class="progress-card-value metric-panel-value"')
    expect(classManagementSource).not.toContain(
      'class="teacher-summary-value progress-card-value metric-panel-value"'
    )
    expect(classManagementSource).toContain('class="progress-card-hint metric-panel-helper"')
    expect(classManagementSource).not.toContain(
      'class="teacher-summary-helper progress-card-hint metric-panel-helper"'
    )

    expect(studentManagementSource).toContain(
      'class="teacher-summary metric-panel-default-surface"'
    )
    expect(studentManagementSource).toContain(
      'class="teacher-summary-grid progress-strip metric-panel-grid"'
    )
    expect(studentManagementSource).toContain('class="progress-card metric-panel-card"')
    expect(studentManagementSource).not.toContain(
      'class="teacher-summary-item progress-card metric-panel-card"'
    )
    expect(studentManagementSource).toContain('class="progress-card-label metric-panel-label"')
    expect(studentManagementSource).not.toContain(
      'class="teacher-summary-label progress-card-label metric-panel-label"'
    )
    expect(studentManagementSource).toContain('class="progress-card-value metric-panel-value"')
    expect(studentManagementSource).not.toContain(
      'class="teacher-summary-value progress-card-value metric-panel-value"'
    )
    expect(studentManagementSource).toContain('class="progress-card-hint metric-panel-helper"')
    expect(studentManagementSource).not.toContain(
      'class="teacher-summary-helper progress-card-hint metric-panel-helper"'
    )

    expect(instanceManagementSource).toContain(
      'class="teacher-summary metric-panel-default-surface"'
    )
    expect(instanceManagementSource).toContain(
      'class="teacher-summary-grid progress-strip metric-panel-grid"'
    )
    expect(instanceManagementSource).toContain('class="progress-card metric-panel-card"')
    expect(instanceManagementSource).not.toContain(
      'class="teacher-summary-item progress-card metric-panel-card"'
    )
    expect(instanceManagementSource).toContain('class="progress-card-label metric-panel-label"')
    expect(instanceManagementSource).not.toContain(
      'class="teacher-summary-label progress-card-label metric-panel-label"'
    )
    expect(instanceManagementSource).toContain('class="progress-card-value metric-panel-value"')
    expect(instanceManagementSource).not.toContain(
      'class="teacher-summary-value progress-card-value metric-panel-value"'
    )
    expect(instanceManagementSource).toContain('class="progress-card-hint metric-panel-helper"')
    expect(instanceManagementSource).not.toContain(
      'class="teacher-summary-helper progress-card-hint metric-panel-helper"'
    )

    expect(awdReviewIndexSource).toContain('metric-panel-default-surface')
    expect(awdReviewIndexSource).toContain('metric-panel-grid')
    expect(awdReviewIndexSource).toContain('metric-panel-card')
    expect(awdReviewIndexSource).not.toContain('teacher-summary-item')
    expect(awdReviewDetailSource).toContain('metric-panel-default-surface')
    expect(awdReviewDetailSource).toContain('metric-panel-grid')
    expect(awdReviewDetailSource).toContain('metric-panel-card')
    expect(awdReviewDetailSource).not.toContain('teacher-summary-item')

    expect(teacherSurfaceSource).not.toContain('--metric-panel-border: var(--teacher-card-border);')
    expect(teacherSurfaceSource).not.toContain(
      '--metric-panel-radius: var(--workspace-radius-lg, 18px);'
    )
    expect(teacherSurfaceSource).not.toContain('--metric-panel-value-size: var(--font-size-26);')
    expect(teacherSurfaceSource).not.toContain('--metric-panel-helper-line-height: 1.7;')
    expect(teacherSurfaceSource).not.toMatch(
      /\.teacher-summary-item,\s*\.teacher-management-shell \.report-summary-item\s*\{\s*min-width:\s*0;\s*border:\s*1px solid/s
    )
    expect(teacherSurfaceSource).not.toMatch(
      /\.teacher-summary-label,\s*\.teacher-management-shell \.report-summary-label\s*\{\s*font-size:\s*var\(--font-size-11\)/s
    )
    expect(teacherSurfaceSource).not.toMatch(
      /\.teacher-summary-value,\s*\.teacher-management-shell \.report-summary-value\s*\{\s*margin-top:\s*var\(--space-2-5,\s*0\.625rem\);\s*font-size:\s*var\(--font-size-26\)/s
    )
  })

  it('awd review pages should soften page header, cards, and divider borders through shared teacher shells', () => {
    expect(awdReviewIndexSource).toContain('teacher-management-shell')
    expect(awdReviewIndexSource).not.toContain('.teacher-btn {')
    expect(awdReviewIndexSource).not.toMatch(/^\.teacher-hero\s*\{/m)
    expect(awdReviewIndexSource).not.toMatch(/^\.teacher-summary\s*\{/m)

    expect(awdReviewDetailSource).toContain('teacher-management-shell')
    expect(awdReviewDetailSource).not.toContain('.teacher-btn {')
    expect(awdReviewDetailSource).not.toMatch(/^\.teacher-hero\s*\{/m)
    expect(awdReviewDetailSource).not.toMatch(/^\.teacher-summary\s*\{/m)
  })
})
