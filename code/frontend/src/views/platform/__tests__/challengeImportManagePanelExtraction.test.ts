import { describe, expect, it } from 'vitest'

import challengeImportManageSource from '../ChallengeImportManage.vue?raw'
import challengeImportHeroPanelSource from '@/components/platform/challenge/ChallengeImportHeroPanel.vue?raw'

describe('ChallengeImportManage panel extraction', () => {
  it('应将导入页头部和操作按钮抽到独立 platform challenge 组件', () => {
    expect(challengeImportManageSource).toContain(
      "import ChallengeImportHeroPanel from '@/components/platform/challenge/ChallengeImportHeroPanel.vue'"
    )
    expect(challengeImportManageSource).toContain('<ChallengeImportHeroPanel')
    expect(challengeImportHeroPanelSource).toContain('Challenge Import')
    expect(challengeImportHeroPanelSource).toContain('返回题目目录')
    expect(challengeImportHeroPanelSource).toContain('题目包规范')
    expect(challengeImportHeroPanelSource).toContain('下载示例题目包')
  })
})
