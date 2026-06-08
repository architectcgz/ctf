import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import { useAuthStore } from '@/stores/auth'
import {
  contestApiMocks,
  contestDetailPageSource,
  contestDetailRoutePageSource,
  contestDetailSource,
  contestDetailWorkspaceSource,
  contestPresentationSource,
  contestTeamDialogsSource,
  contestTeamPanelSource,
  contestTeamWorkspaceSectionSource,
  destructiveConfirmMock,
  resetContestDetailTestHarness,
  routeQueryTransportSource,
  router,
  teamPresentationSource,
  webSocketMocks,
} from './ContestDetail.test-harness'
import ContestDetail from '@/pages/contests/ContestDetailRoutePage.vue'

describe('ContestDetail', () => {
  beforeEach(async () => {
    await resetContestDetailTestHarness()
  })

  it('页面应通过 feature route model 获取路由与派生状态，不再直接管理 tab 和 contest 可见性逻辑', () => {
    expect(contestDetailSource).toContain('useContestDetailRoutePage')
    expect(contestDetailSource).toContain(
      "import { ContestDetailWorkspace } from '@/widgets/contest-detail-workspace'"
    )
    expect(contestDetailSource).not.toContain("from '@/composables/useUrlSyncedTabs'")
    expect(contestDetailSource).not.toContain("from '@/stores/auth'")
    expect(contestDetailSource).not.toContain("from '@/utils/contest'")
    expect(contestDetailSource).not.toContain('const workspaceTabOrder')
    expect(contestDetailSource).not.toContain('const contestAccentStyle = computed')
    expect(contestDetailSource).not.toContain('const contestAccessible = computed')
    expect(contestDetailSource).not.toContain(':contest-accent-style')
    expect(contestDetailSource).not.toContain('ContestOverviewPanel')
    expect(contestDetailSource).not.toContain('ContestTeamDialogs')
    expect(contestDetailSource).not.toContain('router,')
    expect(contestDetailRoutePageSource).toContain(
      "import { useRouteQueryTransport } from '@/shared/model/navigation/useRouteQueryTransport'"
    )
    expect(contestDetailRoutePageSource).toContain(
      "import { useRouteQueryTabs } from '@/shared/model/navigation/useRouteQueryTabs'"
    )
    expect(contestDetailRoutePageSource).not.toContain("from 'vue-router'")
    expect(contestDetailRoutePageSource).not.toContain(
      "from '@/shared/model/navigation/useUrlSyncedTabs'"
    )
    expect(contestDetailRoutePageSource).not.toContain('getContestAccentColor')
    expect(contestDetailRoutePageSource).toContain(
      'const { params, query, replaceQuery } = useRouteQueryTransport()'
    )
    expect(routeQueryTransportSource).toContain(
      'const params = computed<Record<string, unknown>>(() => route.params)'
    )
    expect(contestDetailWorkspaceSource).toContain(
      "import { getContestAccentVarStyle } from '@/entities/contest'"
    )
    expect(contestPresentationSource).toContain('getContestAccentVarStyle')
  })

  it('队伍展示 owner 应由 team entity 承接，而不是继续散落在 route model 和 feature ui', () => {
    expect(contestDetailPageSource).toContain(
      "import { isCurrentUserTeamCaptain } from '@/entities/team'"
    )
    expect(contestDetailPageSource).not.toContain('team.value.captain_user_id === currentUserId')
    expect(contestDetailRoutePageSource).toContain(
      "import { getTeamMemberCount } from '@/entities/team'"
    )
    expect(contestDetailRoutePageSource).not.toContain('page.team.value?.members.length ?? 0')
    expect(contestTeamPanelSource).toContain(
      "import { buildTeamMemberPresentation, getTeamEmptyStateLabel } from '@/entities/team'"
    )
    expect(contestTeamPanelSource).not.toContain('member.user_id === team.captain_user_id')
    expect(contestTeamPanelSource).not.toContain('当前账号尚未加入队伍。')
    expect(contestTeamWorkspaceSectionSource).toContain(
      "import { getTeamInviteCodeLabel } from '@/entities/team'"
    )
    expect(contestTeamWorkspaceSectionSource).not.toContain('邀请码: {{ team.invite_code }}')
    expect(teamPresentationSource).toContain('isCurrentUserTeamCaptain')
    expect(teamPresentationSource).toContain('getTeamInviteCodeLabel')
    expect(teamPresentationSource).toContain('buildTeamMemberPresentation')
  })
})
