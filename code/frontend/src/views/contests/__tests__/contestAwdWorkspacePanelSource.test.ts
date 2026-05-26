import { describe, expect, it } from 'vitest'

import awdDefenseAlertsPanelSource from '@/components/contests/awd/AWDDefenseAlertsPanel.vue?raw'
import awdDefenseColumnSource from '@/components/contests/awd/AWDDefenseColumn.vue?raw'
import awdAttackTargetGridSource from '@/components/contests/awd/AWDAttackTargetGrid.vue?raw'
import awdAttackResultFooterSource from '@/components/contests/awd/AWDAttackResultFooter.vue?raw'
import awdAttackToolbarSource from '@/components/contests/awd/AWDAttackToolbar.vue?raw'
import awdAttackVectorPanelSource from '@/components/contests/awd/AWDAttackVectorPanel.vue?raw'
import awdDefenseOperationsPanelSource from '@/components/contests/awd/AWDDefenseOperationsPanel.vue?raw'
import awdDefenseConnectionPanelSource from '@/components/contests/awd/AWDDefenseConnectionPanel.vue?raw'
import awdWorkspaceHudStripSource from '@/components/contests/awd/AWDWorkspaceHudStrip.vue?raw'
import awdWorkspaceIntelColumnSource from '@/components/contests/awd/AWDWorkspaceIntelColumn.vue?raw'
import awdWorkspaceSource from '@/components/contests/ContestAWDWorkspacePanel.vue?raw'
import studentRoutesSource from '@/router/routes/studentRoutes.ts?raw'

describe('ContestAWDWorkspacePanel source', () => {
  it('AWD 工作台应保留当前战情面板结构与运行态 service 标识', () => {
    expect(awdWorkspaceSource).toContain('formatServiceRef')
    expect(awdWorkspaceSource).toContain('AWDDefenseColumn')
    expect(awdWorkspaceSource).toContain('AWDAttackVectorPanel')
    expect(awdWorkspaceSource).toContain('AWDWorkspaceHudStrip')
    expect(awdWorkspaceSource).toContain('AWDWorkspaceIntelColumn')
    expect(awdWorkspaceSource).toContain('getSSHCommand')
    expect(awdWorkspaceSource).toContain('copySSHCommand')
    expect(awdDefenseColumnSource).toContain('我的防守')
    expect(awdDefenseColumnSource).toContain('AWDDefenseAlertsPanel')
    expect(awdDefenseColumnSource).toContain('AWDDefenseServiceList')
    expect(awdDefenseColumnSource).toContain('AWDDefenseOperationsPanel')
    expect(awdDefenseColumnSource).toContain("emit('refresh')")
    expect(awdDefenseAlertsPanelSource).toContain('defense-alert')
    expect(awdDefenseAlertsPanelSource).toContain('alert.challengeTitle')
    expect(awdDefenseAlertsPanelSource).toContain('alert.statusLabel')
    expect(awdAttackToolbarSource).toContain('目标题目')
    expect(awdAttackToolbarSource).toContain('队伍筛选')
    expect(awdAttackToolbarSource).toContain('id="awd-target-challenge"')
    expect(awdAttackToolbarSource).toContain('id="awd-target-search"')
    expect(awdAttackVectorPanelSource).toContain('攻击向量')
    expect(awdAttackVectorPanelSource).toContain('当前竞赛暂无可部署服务。')
    expect(awdAttackVectorPanelSource).toContain('AWDAttackToolbar')
    expect(awdAttackVectorPanelSource).toContain('AWDAttackTargetGrid')
    expect(awdAttackVectorPanelSource).toContain('AWDAttackResultFooter')
    expect(awdAttackTargetGridSource).toContain('输入获取到的 Flag...')
    expect(awdAttackTargetGridSource).toContain('awd-open-target-')
    expect(awdAttackTargetGridSource).toContain("emit('openTarget'")
    expect(awdAttackTargetGridSource).toContain("emit('submit'")
    expect(awdAttackResultFooterSource).toContain('result-alert')
    expect(awdAttackResultFooterSource).toContain('Terminal')
    expect(awdWorkspaceHudStripSource).toContain('当前回合')
    expect(awdWorkspaceHudStripSource).toContain('我的战队')
    expect(awdWorkspaceHudStripSource).toContain('战队服务')
    expect(awdWorkspaceHudStripSource).toContain('最高分')
    expect(awdWorkspaceIntelColumnSource).toContain('战场情报')
    expect(awdWorkspaceIntelColumnSource).toContain('最近战报')
    expect(awdWorkspaceIntelColumnSource).toContain('data-testid="awd-feedback-challenge-title"')
    expect(awdDefenseConnectionPanelSource).toContain('SSH 连接')
    expect(awdDefenseConnectionPanelSource).toContain('复制 SSH 命令')
    expect(awdDefenseConnectionPanelSource).toContain('复制密码')
    expect(awdDefenseConnectionPanelSource).toContain('密码')
    expect(awdDefenseConnectionPanelSource).toContain('票据将在')
    expect(awdDefenseConnectionPanelSource).toContain('expires_at')
    expect(awdWorkspaceSource).toContain('复制失败，请手动选择文本')
    expect(awdWorkspaceSource).toContain('copySSHPassword')
    expect(awdWorkspaceSource).not.toContain('openDefenseWorkbench')
    expect(studentRoutesSource).not.toContain("name: 'ContestAWDDefenseWorkbench'")
    expect(studentRoutesSource).not.toContain("path: 'contests/:id/awd/defense/:serviceId'")
  })

  it('攻击向量中区应从父页下沉到独立组件，同时保留筛选与提交 owner', () => {
    expect(awdWorkspaceSource).toContain('<AWDAttackVectorPanel')
    expect(awdWorkspaceSource).not.toContain('<section class="ops-panel">')
    expect(awdWorkspaceSource).not.toContain('请选择目标题目后开始攻击。')
    expect(awdWorkspaceSource).not.toContain('当前题目下没有匹配的目标队伍。')
    expect(awdWorkspaceSource).toContain('handleSubmit')
    expect(awdWorkspaceSource).toContain('flagInputs')
    expect(awdAttackVectorPanelSource).toContain('<section class="ops-panel">')
    expect(awdAttackVectorPanelSource).toContain("emit('update:activeChallengeKey'")
    expect(awdAttackVectorPanelSource).toContain("emit('updateFlag'")
    expect(awdAttackVectorPanelSource).toContain("emit('submit'")
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
