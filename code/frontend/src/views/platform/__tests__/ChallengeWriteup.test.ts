import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import ChallengeWriteupEditorPage from '@/components/platform/writeup/ChallengeWriteupEditorPage.vue'
import ChallengeWriteupViewPage from '@/components/platform/writeup/ChallengeWriteupViewPage.vue'
import challengeWriteupEditorSource from '@/components/platform/writeup/ChallengeWriteupEditorPage.vue?raw'
import challengeWriteupViewSource from '@/components/platform/writeup/ChallengeWriteupViewPage.vue?raw'
import challengeWriteupRouteSource from '../ChallengeWriteup.vue?raw'
import challengeWriteupViewRouteSource from '../ChallengeWriteupView.vue?raw'
import { ApiError } from '@/api/request'

const adminApiMocks = vi.hoisted(() => ({
  getChallengeDetail: vi.fn().mockResolvedValue({
    id: '11',
    title: '双节点演练',
    category: 'web',
    difficulty: 'easy',
    status: 'draft',
    points: 100,
    created_at: '2026-03-10T00:00:00.000Z',
  }),
  getChallengeWriteup: vi.fn().mockResolvedValue({
    id: '21',
    challenge_id: '11',
    title: '官方题解',
    content: '## Step 1',
    visibility: 'public',
    is_recommended: true,
    recommended_at: '2026-03-10T03:00:00.000Z',
    created_at: '2026-03-10T00:00:00.000Z',
    updated_at: '2026-03-10T02:00:00.000Z',
  }),
  saveChallengeWriteup: vi.fn(),
  deleteChallengeWriteup: vi.fn(),
  recommendChallengeWriteup: vi.fn().mockResolvedValue({
    id: '21',
    challenge_id: '11',
    title: '官方题解',
    content: '## Step 1',
    visibility: 'public',
    is_recommended: true,
    recommended_at: '2026-03-10T03:00:00.000Z',
    created_at: '2026-03-10T00:00:00.000Z',
    updated_at: '2026-03-10T03:00:00.000Z',
  }),
  unrecommendChallengeWriteup: vi.fn().mockResolvedValue({
    id: '21',
    challenge_id: '11',
    title: '官方题解',
    content: '## Step 1',
    visibility: 'public',
    is_recommended: false,
    created_at: '2026-03-10T00:00:00.000Z',
    updated_at: '2026-03-10T04:00:00.000Z',
  }),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
}))

const confirmMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/admin/authoring', () => adminApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))
vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: confirmMock,
}))

describe('ChallengeWriteupEditorPage', () => {
  it('路由壳页应仅做组合，不直接绑定路由实例', () => {
    expect(challengeWriteupRouteSource).toContain('useChallengeWriteupPage')
    expect(challengeWriteupRouteSource).not.toContain('useRoute')
    expect(challengeWriteupRouteSource).not.toContain('useRouter')
    expect(challengeWriteupViewRouteSource).toContain('useChallengeWriteupViewPage')
    expect(challengeWriteupViewRouteSource).not.toContain('useRoute')
    expect(challengeWriteupViewRouteSource).not.toContain('useRouter')
  })

  it('题解查看页和编辑页应改用共享 ui-btn 原语而不是页面私有 admin-btn 按钮族', () => {
    expect(challengeWriteupViewSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(challengeWriteupViewSource).toContain('class="ui-btn ui-btn--primary"')
    expect(challengeWriteupViewSource).not.toContain('admin-btn admin-btn-ghost')
    expect(challengeWriteupViewSource).not.toContain('admin-btn admin-btn-primary')

    expect(challengeWriteupEditorSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(challengeWriteupEditorSource).toContain('class="ui-btn ui-btn--primary"')
    expect(challengeWriteupEditorSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(challengeWriteupEditorSource).toContain('class="ui-btn ui-btn--danger"')
    expect(challengeWriteupEditorSource).not.toContain('admin-btn admin-btn-ghost')
    expect(challengeWriteupEditorSource).not.toContain('admin-btn admin-btn-primary')
    expect(challengeWriteupEditorSource).not.toContain('admin-btn admin-btn-accent')
    expect(challengeWriteupEditorSource).not.toContain('admin-btn admin-btn-danger')
  })

  it('嵌入态题解编辑页应使用统一头部样式而不是旧 workspace-tab-heading', () => {
    expect(challengeWriteupEditorSource).toContain('class="list-heading writeup-tab-heading"')
    expect(challengeWriteupEditorSource).toContain(
      '<div class="workspace-overline">Admin Writeup</div>'
    )
    expect(challengeWriteupEditorSource).not.toContain(
      '<div class="journal-note-label">Admin Writeup</div>'
    )
    expect(challengeWriteupEditorSource).toContain('<h1 class="workspace-page-title">题解管理</h1>')
    expect(challengeWriteupEditorSource).not.toContain(
      'class="workspace-tab-heading writeup-tab-heading"'
    )
  })

  it('题解页根壳应统一挂到管理员 workspace shell，不再手写外层圆角', async () => {
    expect(challengeWriteupEditorSource).toContain('workspace-shell')
    expect(challengeWriteupEditorSource).toContain('journal-shell-admin')
    expect(challengeWriteupEditorSource).toContain('journal-hero')
    expect(challengeWriteupEditorSource).not.toContain('rounded-[24px]')

    const editorWrapper = mount(ChallengeWriteupEditorPage, {
      props: {
        challengeId: '11',
      },
      global: {
        stubs: {
          AppCard: { template: '<section><slot name="header" /><slot /></section>' },
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
          PageHeader: { template: '<div><slot /></div>' },
        },
      },
    })

    const viewWrapper = mount(ChallengeWriteupViewPage, {
      props: {
        challengeId: '11',
      },
      global: {
        stubs: {
          AppCard: { template: '<section><slot name="header" /><slot /></section>' },
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
          PageHeader: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    expect(editorWrapper.classes()).toContain('workspace-shell')
    expect(editorWrapper.classes()).toContain('journal-shell-admin')
    expect(editorWrapper.classes()).toContain('journal-hero')
    expect(viewWrapper.classes()).toContain('workspace-shell')
    expect(viewWrapper.classes()).toContain('journal-shell-admin')
    expect(viewWrapper.classes()).toContain('journal-hero')
  })

  it('删除题解失败时应优先展示接口返回消息', async () => {
    confirmMock.mockResolvedValue(true)
    adminApiMocks.deleteChallengeWriteup.mockRejectedValue(
      new ApiError('题解正在审核流程中，暂时不能删除', { code: 10007, status: 409 })
    )

    const wrapper = mount(ChallengeWriteupEditorPage, {
      props: {
        challengeId: '11',
      },
      global: {
        stubs: {
          AppCard: { template: '<section><slot name="header" /><slot /></section>' },
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
          PageHeader: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    const deleteButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('删除题解'))
    expect(deleteButton).toBeTruthy()

    await deleteButton!.trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('题解正在审核流程中，暂时不能删除')
    expect(toastMocks.error).not.toHaveBeenCalledWith('删除题解失败')
  })

  it('应该渲染已保存题解的核心信息', async () => {
    const wrapper = mount(ChallengeWriteupEditorPage, {
      props: {
        challengeId: '11',
      },
      global: {
        stubs: {
          AppCard: { template: '<section><slot name="header" /><slot /></section>' },
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
          PageHeader: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('双节点演练')
    expect(wrapper.text()).toContain('官方题解')
    expect(wrapper.text()).toContain('public')
    expect(wrapper.text()).not.toContain('scheduled')
    expect(wrapper.text()).toContain('推荐题解')
    expect(wrapper.text()).toContain('取消推荐')
    expect(wrapper.text()).toContain('恢复已保存版本')
    expect(wrapper.text()).toContain('2026-03-10T02:00:00.000Z')
  })

  it('应该支持取消推荐官方题解', async () => {
    const wrapper = mount(ChallengeWriteupEditorPage, {
      props: {
        challengeId: '11',
      },
      global: {
        stubs: {
          AppCard: { template: '<section><slot name="header" /><slot /></section>' },
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
          PageHeader: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    const actionButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('取消推荐'))
    expect(actionButton).toBeTruthy()

    await actionButton!.trigger('click')
    await flushPromises()

    expect(adminApiMocks.unrecommendChallengeWriteup).toHaveBeenCalledWith('11')
  })

  it('查看页应独立展示题解内容而不是编辑表单', async () => {
    const wrapper = mount(ChallengeWriteupViewPage, {
      props: {
        challengeId: '11',
      },
      global: {
        stubs: {
          AppCard: { template: '<section><slot name="header" /><slot /></section>' },
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
          PageHeader: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('查看题解')
    expect(wrapper.text()).toContain('题解正文')
    expect(wrapper.text()).toContain('Step 1')
    expect(wrapper.find('main > .writeup-snapshot-grid').exists()).toBe(true)
    expect(wrapper.find('main > .writeup-view-body').exists()).toBe(true)
    expect(wrapper.find('.writeup-reading-card').exists()).toBe(false)
    expect(wrapper.find('.workspace-tab-heading').exists()).toBe(false)
    expect(wrapper.find('input.writeup-field-input').exists()).toBe(false)
    expect(wrapper.find('select.writeup-field-input').exists()).toBe(false)
    expect(wrapper.find('textarea.writeup-content-input').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('保存题解')
    expect(wrapper.text()).not.toContain('删除题解')
    expect(wrapper.text()).not.toContain('恢复已保存版本')
    expect(wrapper.text()).not.toContain('取消推荐')
    expect(wrapper.find('.writeup-reading-card__hero .writeup-badges').exists()).toBe(false)
    expect(wrapper.find('.writeup-reading-card__hero .writeup-badge--accent').exists()).toBe(false)
  })
})
