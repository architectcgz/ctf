import { describe, expect, it } from 'vitest'

import awdWorkspaceSource from '@/features/contest-awd-workspace/ui/ContestAWDWorkspacePanel.vue?raw'
import awdAttackResultFooterSource from '@/features/contest-awd-workspace/ui/AWDAttackResultFooter.vue?raw'
import awdAttackTargetGridSource from '@/features/contest-awd-workspace/ui/AWDAttackTargetGrid.vue?raw'
import awdAttackToolbarSource from '@/features/contest-awd-workspace/ui/AWDAttackToolbar.vue?raw'
import awdAttackVectorPanelSource from '@/features/contest-awd-workspace/ui/AWDAttackVectorPanel.vue?raw'
import awdDefenseColumnSource from '@/features/contest-awd-workspace/ui/AWDDefenseColumn.vue?raw'
import awdDefenseServiceListSource from '@/features/contest-awd-workspace/ui/AWDDefenseServiceList.vue?raw'
import awdDefenseAlertsPanelSource from '@/features/contest-awd-workspace/ui/AWDDefenseAlertsPanel.vue?raw'
import awdAttackVectorStateSource from '@/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.ts?raw'
import awdDefensePresentationSource from '@/features/contest-awd-workspace/model/awdDefensePresentation.ts?raw'
import awdDefenseAccessPanelSource from '@/features/contest-awd-workspace/model/useAwdDefenseAccessPanel.ts?raw'
import awdWorkspacePresentationSource from '@/features/contest-awd-workspace/model/useAwdWorkspacePresentation.ts?raw'
import awdWorkspaceSummarySource from '@/features/contest-awd-workspace/model/useAwdWorkspaceSummary.ts?raw'
import awdDefenseOperationsPanelSource from '@/features/contest-awd-workspace/ui/AWDDefenseOperationsPanel.vue?raw'
import awdDefenseConnectionPanelSource from '@/features/contest-awd-workspace/ui/AWDDefenseConnectionPanel.vue?raw'
import awdWorkspaceHudStripSource from '@/features/contest-awd-workspace/ui/AWDWorkspaceHudStrip.vue?raw'
import awdWorkspaceIntelColumnSource from '@/features/contest-awd-workspace/ui/AWDWorkspaceIntelColumn.vue?raw'
import studentRoutesSource from '@/router/routes/studentRoutes.ts?raw'

const awdActionSurfaceSource = [
  awdWorkspaceSource,
  awdWorkspaceHudStripSource,
  awdDefenseServiceListSource,
  awdAttackToolbarSource,
  awdAttackTargetGridSource,
].join('\n')

describe('contest awd workspace ui strategy', () => {
  it('awd workspace should keep its extracted tactical surface and stable selectors', () => {
    expect(awdWorkspaceSource).toContain('AWDDefenseColumn')
    expect(awdWorkspaceSource).toContain('AWDAttackVectorPanel')
    expect(awdWorkspaceSource).toContain('AWDWorkspaceHudStrip')
    expect(awdWorkspaceSource).toContain('AWDWorkspaceIntelColumn')
    expect(awdWorkspaceSource).toContain('useAwdDefenseAccessPanel')
    expect(awdDefenseColumnSource).toContain('AWDDefenseAlertsPanel')
    expect(awdDefenseColumnSource).toContain('AWDDefenseServiceList')
    expect(awdDefenseColumnSource).toContain('AWDDefenseOperationsPanel')
    expect(awdDefenseAlertsPanelSource).toContain('defense-alert')
    expect(awdDefenseOperationsPanelSource).toContain("emit('refresh')")
  })

  it('awd workspace should keep stable war-room action primitives and targeting selectors', () => {
    expect(awdActionSurfaceSource).toContain('class="hud-refresh-btn"')
    expect(awdActionSurfaceSource).toContain('class="asset-btn asset-btn--primary"')
    expect(awdActionSurfaceSource).toContain('class="war-room-select"')
    expect(awdActionSurfaceSource).toContain('class="war-room-input"')
    expect(awdActionSurfaceSource).toContain('class="flag-input"')
    expect(awdActionSurfaceSource).toContain('class="submit-btn"')
    expect(awdActionSurfaceSource).toContain('id="awd-target-challenge"')
    expect(awdActionSurfaceSource).toContain('id="awd-target-search"')
    expect(awdActionSurfaceSource).not.toMatch(/^\.contest-btn\s*\{/m)
    expect(awdActionSurfaceSource).not.toMatch(/^\.contest-btn--primary\s*\{/m)
    expect(awdActionSurfaceSource).not.toMatch(/^\.contest-btn--ghost\s*\{/m)
  })

  it('awd workspace source contracts should continue to expose the current tactical model seams', () => {
    expect(awdAttackToolbarSource).toContain('目标题目')
    expect(awdAttackToolbarSource).toContain('队伍筛选')
    expect(awdAttackVectorPanelSource).toContain('攻击向量')
    expect(awdAttackVectorPanelSource).toContain('AWDAttackToolbar')
    expect(awdAttackVectorPanelSource).toContain('AWDAttackTargetGrid')
    expect(awdAttackVectorPanelSource).toContain('AWDAttackResultFooter')
    expect(awdAttackResultFooterSource).toContain('result-alert')
    expect(awdWorkspaceHudStripSource).toContain('当前回合')
    expect(awdWorkspaceIntelColumnSource).toContain('战场情报')
    expect(awdWorkspacePresentationSource).toContain('getChallengeTitleForEvent')
    expect(awdWorkspacePresentationSource).toContain('formatAttackResultToast')
    expect(awdWorkspacePresentationSource).toContain('eventDirectionLabel')
    expect(awdWorkspaceSummarySource).toContain('defenseAlerts')
    expect(awdWorkspaceSummarySource).toContain('serviceCount')
    expect(awdWorkspaceSummarySource).toContain('currentRoundLabel')
    expect(awdDefensePresentationSource).toContain('getDefenseServiceStatusLabel')
    expect(awdDefensePresentationSource).toContain('getDefenseInstanceStatusLabel')
    expect(awdDefensePresentationSource).toContain('toDefenseServiceCards')
    expect(awdAttackVectorStateSource).toContain('activeChallengeKey')
    expect(awdAttackVectorStateSource).toContain('activeChallengeRuntimeKey')
    expect(awdAttackVectorStateSource).toContain('attackToolbarChallengeOptions')
    expect(awdDefenseAccessPanelSource).toContain('openDefenseService')
    expect(awdDefenseAccessPanelSource).toContain('copySSHCommand')
    expect(awdDefenseAccessPanelSource).toContain('copySSHPassword')
    expect(awdDefenseConnectionPanelSource).toContain('asset-ssh')
    expect(awdDefenseConnectionPanelSource).toContain('复制 SSH 命令')
    expect(awdDefenseConnectionPanelSource).toContain('复制密码')
    expect(studentRoutesSource).toContain("path: 'contests/:id'")
  })
})
