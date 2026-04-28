import { describe, expect, it } from 'vitest'

import awdWorkspaceSource from '@/components/contests/ContestAWDWorkspacePanel.vue?raw'

describe('ContestAWDWorkspacePanel source', () => {
  it('AWD 工作台应保留当前战情面板结构与运行态 service 标识', () => {
    expect(awdWorkspaceSource).toContain('防守监控')
    expect(awdWorkspaceSource).toContain('攻击向量')
    expect(awdWorkspaceSource).toContain('战场情报')
    expect(awdWorkspaceSource).toContain('最近战报')
    expect(awdWorkspaceSource).toContain('目标题目')
    expect(awdWorkspaceSource).toContain('队伍筛选')
    expect(awdWorkspaceSource).toContain('输入获取到的 Flag...')
    expect(awdWorkspaceSource).toContain('当前竞赛暂无可部署服务。')
    expect(awdWorkspaceSource).toContain('formatServiceRef')
    expect(awdWorkspaceSource).toContain('id="awd-target-challenge"')
    expect(awdWorkspaceSource).toContain('id="awd-target-search"')
    expect(awdWorkspaceSource).toContain('data-testid="awd-feedback-challenge-title"')
    expect(awdWorkspaceSource).toContain('ssh_profile')
    expect(awdWorkspaceSource).toContain('buildOpenSSHConfig')
    expect(awdWorkspaceSource).toContain('copySSHConfig')
    expect(awdWorkspaceSource).toContain('VS Code')
  })

  it('不再暴露浏览器文件防守工作台入口', () => {
    expect(awdWorkspaceSource).not.toContain("name: 'ContestAWDDefenseWorkbench'")
    expect(awdWorkspaceSource).not.toContain('openDefenseWorkbench(challenge.awd_service_id)')
    expect(awdWorkspaceSource).not.toContain('asset-btn--defense')
    expect(awdWorkspaceSource).not.toContain('class="defense-workbench"')
  })
})
