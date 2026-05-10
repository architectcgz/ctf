import { describe, expect, it } from 'vitest'

import contestDetailSource from '@/views/contests/ContestDetail.vue?raw'
import contestListSource from '@/views/contests/ContestList.vue?raw'
import instanceListSource from '@/views/instances/InstanceList.vue?raw'
import awdReviewDirectorySource from '@/components/platform/awd-review/AwdReviewDirectoryPanel.vue?raw'
import awdChallengeLibrarySource from '@/components/platform/awd-service/AWDChallengeLibraryPage.vue?raw'
import cheatDetectionPanelsSource from '@/components/platform/cheat/CheatDetectionReviewPanels.vue?raw'
import classStudentsPageSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import contestChallengeEditorDialogSource from '@/components/platform/contest/ContestChallengeEditorDialog.vue?raw'
import contestOperationsHubSource from '@/components/platform/contest/ContestOperationsHubWorkspacePanel.vue?raw'
import platformContestTableSource from '@/components/platform/contest/PlatformContestTable.vue?raw'
import scoreboardDetailSource from '@/views/scoreboard/ScoreboardDetail.vue?raw'
import scoreboardSource from '@/views/scoreboard/ScoreboardView.vue?raw'
import appStyleSource from '@/style.css?raw'

function extractTemplateSlot(source: string, slotName: string): string {
  const match = source.match(new RegExp(`#${slotName}[^>]*>([\\s\\S]*?)<\\/template>`))
  return match?.[1] ?? ''
}

describe('student directory typography boundary', () => {
  it('学生侧普通目录标题不应使用局部等宽字体变体', () => {
    expect(appStyleSource).not.toContain('.workspace-directory-row-title--mono')
    expect(contestListSource).not.toContain('workspace-directory-row-title--mono')
    expect(scoreboardSource).not.toContain('workspace-directory-row-title--mono')
    expect(instanceListSource).not.toContain('workspace-directory-row-title--mono')
  })

  it('学生侧赛事与排行榜普通文本不应强制使用等宽字体', () => {
    expect(contestDetailSource).not.toMatch(
      /\.team-summary__invite\s*\{[\s\S]*?font-family:\s*var\(--font-family-mono\)/m
    )
    expect(scoreboardSource).not.toMatch(
      /\.sb-cell--(?:rank|mono)\s*[\s\S]*?font-family:\s*var\(--font-family-mono\)/m
    )
    expect(scoreboardDetailSource).not.toMatch(
      /\.sb-cell--(?:rank|mono)\s*[\s\S]*?font-family:\s*var\(--font-family-mono\)/m
    )
    expect(scoreboardSource).not.toContain('class="sb-cell--mono"')
    expect(scoreboardDetailSource).not.toContain('class="sb-cell--mono"')
  })

  it('学生侧列表主标题列应保持纯净，不混入标签、序号或描述', () => {
    expect(contestListSource).not.toContain('contest-row-status-strip')
    expect(instanceListSource).not.toContain('instance-row-tags')
    expect(scoreboardSource).not.toContain('scoreboard-card-chips')
    expect(scoreboardSource).not.toContain('scoreboard-card-description')
  })

  it('教师和管理员列表主标题列也应保持纯净', () => {
    expect(classStudentsPageSource).toContain('<span>学生名称</span>')
    expect(classStudentsPageSource).toContain('<span>薄弱项</span>')
    expect(extractTemplateSlot(awdReviewDirectorySource, 'cell-title')).not.toContain(
      'latest_evidence_at'
    )
    expect(extractTemplateSlot(awdChallengeLibrarySource, 'cell-name')).not.toContain(
      '(row as AdminAwdChallengeData).slug'
    )
    expect(extractTemplateSlot(platformContestTableSource, 'cell-title')).not.toContain(
      '(row as ContestDetailData).description'
    )
    expect(extractTemplateSlot(contestOperationsHubSource, 'cell-title')).not.toContain(
      '(row as ContestDetailData).description'
    )
    expect(extractTemplateSlot(contestChallengeEditorDialogSource, 'cell-name')).not.toContain(
      '(row as AdminAwdChallengeData).slug'
    )
    expect(cheatDetectionPanelsSource).toContain(
      '</div>\n        <div class="cheat-directory-row-copy"'
    )
  })
})
