import { describe, expect, it } from 'vitest'

import awdDefenseOperationsPanelSource from '@/components/contests/awd/AWDDefenseOperationsPanel.vue?raw'
import awdDefenseConnectionPanelSource from '@/components/contests/awd/AWDDefenseConnectionPanel.vue?raw'
import awdWorkspaceSource from '@/components/contests/ContestAWDWorkspacePanel.vue?raw'

describe('ContestAWDWorkspacePanel source', () => {
  it('AWD 工作台应保留当前战情面板结构与运行态 service 标识', () => {
    expect(awdWorkspaceSource).toContain('我的防守')
    expect(awdWorkspaceSource).toContain('攻击向量')
    expect(awdWorkspaceSource).toContain('战场情报')
    expect(awdWorkspaceSource).toContain('最近战报')
    expect(awdWorkspaceSource).toContain('ContestAWDDefenseWorkbench')
    expect(awdWorkspaceSource).toContain('目标题目')
    expect(awdWorkspaceSource).toContain('队伍筛选')
    expect(awdWorkspaceSource).toContain('输入获取到的 Flag...')
    expect(awdWorkspaceSource).toContain('当前竞赛暂无可部署服务。')
    expect(awdWorkspaceSource).toContain('formatServiceRef')
    expect(awdWorkspaceSource).toContain('id="awd-target-challenge"')
    expect(awdWorkspaceSource).toContain('id="awd-target-search"')
    expect(awdWorkspaceSource).toContain('data-testid="awd-feedback-challenge-title"')
    expect(awdWorkspaceSource).toContain('AWDDefenseOperationsPanel')
    expect(awdWorkspaceSource).toContain('getSSHCommand')
    expect(awdWorkspaceSource).toContain('copySSHCommand')
    expect(awdDefenseConnectionPanelSource).toContain('SSH 连接')
    expect(awdDefenseConnectionPanelSource).toContain('复制 SSH 命令')
    expect(awdDefenseConnectionPanelSource).toContain('密码')
    expect(awdDefenseConnectionPanelSource).toContain('票据将在')
    expect(awdDefenseConnectionPanelSource).toContain('expires_at')
    expect(awdWorkspaceSource).toContain('复制失败，请手动选择文本')
    expect(awdWorkspaceSource).toContain('openDefenseWorkbench')
  })

  it('学生战场页不暴露源码文件防守工作台入口', () => {
    expect(awdWorkspaceSource).not.toContain('AWDDefenseFileWorkbench')
    expect(awdWorkspaceSource).not.toContain('requestContestAWDDefenseDirectory')
    expect(awdWorkspaceSource).not.toContain('requestContestAWDDefenseFile')
    expect(awdWorkspaceSource).not.toContain('loadDefenseDirectory')
    expect(awdWorkspaceSource).not.toContain('openDefenseFile')
    expect(awdWorkspaceSource).not.toContain('requestContestAWDDefenseCommand')
    expect(awdWorkspaceSource).not.toContain('saveContestAWDDefenseFile')
  })

  it('战场侧栏不再展示防守范围面板', () => {
    expect(awdDefenseOperationsPanelSource).not.toContain('防守范围')
    expect(awdDefenseOperationsPanelSource).not.toContain('当前服务暂无防守范围数据')
  })
})
