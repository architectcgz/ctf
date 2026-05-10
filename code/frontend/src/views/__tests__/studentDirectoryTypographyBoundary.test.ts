import { describe, expect, it } from 'vitest'

import contestDetailSource from '@/views/contests/ContestDetail.vue?raw'
import contestListSource from '@/views/contests/ContestList.vue?raw'
import instanceListSource from '@/views/instances/InstanceList.vue?raw'
import scoreboardDetailSource from '@/views/scoreboard/ScoreboardDetail.vue?raw'
import scoreboardSource from '@/views/scoreboard/ScoreboardView.vue?raw'
import appStyleSource from '@/style.css?raw'

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
})
