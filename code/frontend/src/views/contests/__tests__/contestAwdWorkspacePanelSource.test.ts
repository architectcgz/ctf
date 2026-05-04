import { describe, expect, it } from 'vitest'

import awdDefenseConnectionPanelSource from '@/components/contests/awd/AWDDefenseConnectionPanel.vue?raw'
import awdDefenseOperationsPanelSource from '@/components/contests/awd/AWDDefenseOperationsPanel.vue?raw'
import awdWorkspaceSource from '@/components/contests/ContestAWDWorkspacePanel.vue?raw'

describe('ContestAWDWorkspacePanel source', () => {
  it('AWD 工作台应保留当前战情面板结构与运行态 service 标识', () => {
    expect(awdWorkspaceSource).toContain('我的防守')
    expect(awdWorkspaceSource).toContain('AWDDefenseOperationsPanel')
    expect(awdDefenseOperationsPanelSource).toContain('Web 防守')
    expect(awdDefenseOperationsPanelSource).toContain('防守范围')
    expect(awdDefenseOperationsPanelSource).toContain('打开服务')
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
    expect(awdWorkspaceSource).toContain('getSSHCommand')
    expect(awdWorkspaceSource).toContain('copySSHCommand')
    expect(awdWorkspaceSource).toContain('copySSHConfig')
    expect(awdDefenseConnectionPanelSource).toContain('VS Code Remote-SSH')
    expect(awdDefenseConnectionPanelSource).toContain('复制 VS Code 命令')
    expect(awdDefenseConnectionPanelSource).toContain('OpenSSH 配置')
    expect(awdDefenseConnectionPanelSource).toContain('票据将在')
    expect(awdDefenseConnectionPanelSource).toContain('expires_at')
    expect(awdWorkspaceSource).toContain('复制失败，请手动选择文本')
  })

  it('学生战场不暴露完整浏览器文件工作台入口', () => {
    expect(awdWorkspaceSource).not.toContain('AWDDefenseFileWorkbench')
    expect(awdWorkspaceSource).not.toContain('requestContestAWDDefenseDirectory')
    expect(awdWorkspaceSource).not.toContain('requestContestAWDDefenseFile')
    expect(awdWorkspaceSource).not.toContain('loadDefenseDirectory')
    expect(awdWorkspaceSource).not.toContain('openDefenseFile')
    expect(awdWorkspaceSource).not.toContain('requestContestAWDDefenseCommand')
    expect(awdWorkspaceSource).not.toContain('saveContestAWDDefenseFile')
    expect(awdWorkspaceSource).not.toContain('app.py')
  })
})
