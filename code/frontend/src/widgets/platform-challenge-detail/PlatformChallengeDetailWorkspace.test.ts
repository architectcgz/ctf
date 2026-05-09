import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import PlatformChallengeDetailWorkspace from './PlatformChallengeDetailWorkspace.vue'
import type { AdminChallengeListItem } from '@/api/contracts'

function createProps() {
  return {
    workspaceLabel: '题目详情',
    hasChallengeId: true,
    loading: false,
    panelTabs: [
      {
        key: 'detail' as const,
        label: '题目详情',
        tabId: 'admin-challenge-tab-detail',
        panelId: 'admin-challenge-panel-detail',
      },
    ],
    activePanel: 'detail' as const,
    setTabButtonRef: () => undefined,
    challenge: {
      id: 'challenge-1',
      title: 'SQLi',
      category: 'web',
      difficulty: 'easy',
      points: 100,
      status: 'draft',
      tags: [],
      solved_count: 0,
      total_attempts: 0,
      is_solved: false,
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    } as AdminChallengeListItem,
    downloadingAttachment: false,
    flagDraft: {
      flagConfigSummary: '静态 Flag',
      flagDraftSummary: '静态 Flag',
      flagPrefix: '',
      flagRegex: '',
      flagType: 'static' as const,
      flagValue: '',
      isSharedInstanceChallenge: false,
      saving: false,
    },
    challengeId: 'challenge-1',
  }
}

describe('PlatformChallengeDetailWorkspace', () => {
  it('应向 Topbar 透传标题与 challenge 标识并转发顶部动作事件', async () => {
    const wrapper = mount(PlatformChallengeDetailWorkspace, {
      props: createProps(),
      global: {
        stubs: {
          AdminChallengeTopbarPanel: {
            template:
              '<div data-testid="topbar" @click="$emit(\'open-topology\')">{{ workspaceLabel }}-{{ hasChallengeId }}</div>',
            props: ['workspaceLabel', 'hasChallengeId'],
          },
          AdminChallengeWorkspaceTabs: {
            template: '<div data-testid="tabs" />',
          },
        },
      },
    })

    expect(wrapper.get('[data-testid="topbar"]').text()).toContain('题目详情-true')

    await wrapper.get('[data-testid="topbar"]').trigger('click')
    expect(wrapper.emitted('openTopology')).toBeTruthy()
  })

  it('应转发 WorkspaceTabs 的交互事件与草稿 patch', async () => {
    const wrapper = mount(PlatformChallengeDetailWorkspace, {
      props: createProps(),
      global: {
        stubs: {
          AdminChallengeTopbarPanel: { template: '<div />' },
          AdminChallengeWorkspaceTabs: {
            template:
              '<button data-testid="tabs" @click="$emit(\'update:flag-draft\', { flagType: \'regex\' })">tabs</button>',
          },
        },
      },
    })

    await wrapper.get('[data-testid="tabs"]').trigger('click')
    expect(wrapper.emitted('updateFlagDraft')).toEqual([[{ flagType: 'regex' }]])
  })
})
