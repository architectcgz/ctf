import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/api/request'

import ChallengeWriteupManagePanel from '@/components/admin/writeup/ChallengeWriteupManagePanel.vue'

const pushMock = vi.fn()

const adminApiMocks = vi.hoisted(() => ({
  getChallengeWriteup: vi.fn(),
  deleteChallengeWriteup: vi.fn(),
}))

const teacherApiMocks = vi.hoisted(() => ({
  getTeacherWriteupSubmissions: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
}))

const confirmMock = vi.hoisted(() => vi.fn())

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/admin', () => adminApiMocks)
vi.mock('@/api/teacher', () => teacherApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))
vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: confirmMock,
}))

describe('ChallengeWriteupManagePanel', () => {
  beforeEach(() => {
    pushMock.mockReset()
    adminApiMocks.getChallengeWriteup.mockReset()
    adminApiMocks.deleteChallengeWriteup.mockReset()
    teacherApiMocks.getTeacherWriteupSubmissions.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    confirmMock.mockReset()
  })

  it('存在题解时应显示带边框的查看入口，并通过更多菜单进入编辑', async () => {
    adminApiMocks.getChallengeWriteup.mockResolvedValue({
      id: '21',
      challenge_id: '11',
      title: '官方题解',
      content: '## Step 1',
      visibility: 'public',
      is_recommended: true,
      recommended_at: '2026-03-10T03:00:00.000Z',
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T02:00:00.000Z',
    })
    teacherApiMocks.getTeacherWriteupSubmissions.mockResolvedValue({
      list: [
        {
          id: '31',
          user_id: '401',
          student_username: 'alice_01',
          student_name: '张三',
          student_no: '20240001',
          class_name: '网安 1 班',
          challenge_id: '11',
          challenge_title: '双节点演练',
          title: '利用链梳理',
          content_preview: '从入口点开始拆解',
          submission_status: 'published',
          visibility_status: 'visible',
          is_recommended: true,
          published_at: '2026-03-10T01:00:00.000Z',
          updated_at: '2026-03-10T02:30:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 6,
    })

    const wrapper = mount(ChallengeWriteupManagePanel, {
      props: {
        challengeId: '11',
        challengeTitle: '双节点演练',
      },
      global: {
        stubs: {
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('.writeup-manage-header').exists()).toBe(true)
    expect(wrapper.find('.writeup-manage-header .workspace-tab-heading').exists()).toBe(true)
    expect(wrapper.find('.writeup-manage-header .writeup-manage-actions').exists()).toBe(true)
    expect(wrapper.find('.writeup-manage-header + .writeup-manage-stats-shell').exists()).toBe(true)
    expect(wrapper.text()).toContain('题解列表')
    expect(wrapper.text()).toContain('来源')
    expect(wrapper.find('.writeup-manage-section__copy').exists()).toBe(false)
    expect(wrapper.findAll('.writeup-manage-stat')).toHaveLength(2)
    expect(wrapper.text()).toContain('官方题解1篇')
    expect(wrapper.text()).toContain('学员题解1篇')
    expect(wrapper.text()).toContain('官方题解')
    expect(wrapper.text()).toContain('官方')
    expect(wrapper.text()).toContain('public')
    expect(wrapper.text()).toContain('推荐题解')
    expect(wrapper.text()).toContain('张三')
    expect(wrapper.text()).toContain('20240001')
    expect(wrapper.text()).toContain('alice_01')
    expect(teacherApiMocks.getTeacherWriteupSubmissions).toHaveBeenCalledWith({
      challenge_id: '11',
      page: 1,
      page_size: 6,
    })

    const viewButton = wrapper.findAll('button').find((button) => button.text().trim() === '查看')
    const moreButton = wrapper.get('[data-testid="writeup-more-actions"]')
    expect(viewButton).toBeTruthy()
    expect(viewButton!.classes()).toContain('admin-btn-outline')
    expect(moreButton.attributes('aria-expanded')).toBe('false')

    await viewButton!.trigger('click')
    expect(pushMock).toHaveBeenNthCalledWith(1, {
      path: '/platform/challenges/11/writeup/view',
    })

    await moreButton.trigger('mouseenter')

    const editButton = wrapper.findAll('[role="menuitem"]').find((button) => button.text().trim() === '编辑')
    expect(moreButton.attributes('aria-expanded')).toBe('true')
    expect(editButton).toBeTruthy()

    await editButton!.trigger('click')
    expect(pushMock).toHaveBeenNthCalledWith(2, {
      path: '/platform/challenges/11/writeup',
    })
  })

  it('存在题解时应支持直接删除并刷新为空状态', async () => {
    confirmMock.mockResolvedValue(true)
    adminApiMocks.getChallengeWriteup
      .mockResolvedValueOnce({
        id: '21',
        challenge_id: '11',
        title: '官方题解',
        content: '## Step 1',
        visibility: 'public',
        is_recommended: true,
        recommended_at: '2026-03-10T03:00:00.000Z',
        created_at: '2026-03-10T00:00:00.000Z',
        updated_at: '2026-03-10T02:00:00.000Z',
      })
      .mockResolvedValueOnce(null)
    teacherApiMocks.getTeacherWriteupSubmissions.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 6,
    })

    const wrapper = mount(ChallengeWriteupManagePanel, {
      props: {
        challengeId: '11',
        challengeTitle: '双节点演练',
      },
      global: {
        stubs: {
          AppEmpty: {
            template:
              '<div><div>{{ title }}</div><div>{{ description }}</div><slot /></div>',
            props: ['title', 'description'],
          },
          AppLoading: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    await wrapper.get('[data-testid="writeup-more-actions"]').trigger('mouseenter')

    const deleteButton = wrapper.findAll('[role="menuitem"]').find((button) => button.text().trim() === '删除')
    expect(deleteButton).toBeTruthy()

    await deleteButton!.trigger('click')
    await flushPromises()

    expect(confirmMock).toHaveBeenCalledWith({
      message: '确定删除当前题解吗？删除后学员将无法继续查看。',
    })
    expect(adminApiMocks.deleteChallengeWriteup).toHaveBeenCalledWith('11')
    expect(toastMocks.success).toHaveBeenCalledWith('题解已删除')
    expect(wrapper.text()).toContain('当前还没有题解')
  })

  it('目录删除失败时应优先展示接口返回消息', async () => {
    confirmMock.mockResolvedValue(true)
    adminApiMocks.getChallengeWriteup.mockResolvedValue({
      id: '21',
      challenge_id: '11',
      title: '官方题解',
      content: '## Step 1',
      visibility: 'public',
      is_recommended: true,
      recommended_at: '2026-03-10T03:00:00.000Z',
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T02:00:00.000Z',
    })
    teacherApiMocks.getTeacherWriteupSubmissions.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 6,
    })
    adminApiMocks.deleteChallengeWriteup.mockRejectedValue(
      new ApiError('题解正在审核流程中，暂时不能删除', { code: 10007, status: 409 })
    )

    const wrapper = mount(ChallengeWriteupManagePanel, {
      props: {
        challengeId: '11',
        challengeTitle: '双节点演练',
      },
      global: {
        stubs: {
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    await wrapper.get('[data-testid="writeup-more-actions"]').trigger('mouseenter')

    const deleteButton = wrapper.findAll('[role="menuitem"]').find((button) => button.text().trim() === '删除')
    expect(deleteButton).toBeTruthy()

    await deleteButton!.trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('题解正在审核流程中，暂时不能删除')
    expect(wrapper.text()).toContain('官方题解')
  })

  it('没有题解时应显示空状态并支持编写题解', async () => {
    adminApiMocks.getChallengeWriteup.mockResolvedValue(null)
    teacherApiMocks.getTeacherWriteupSubmissions.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 6,
    })

    const wrapper = mount(ChallengeWriteupManagePanel, {
      props: {
        challengeId: '11',
        challengeTitle: '双节点演练',
      },
      global: {
        stubs: {
          AppEmpty: {
            template:
              '<div><div>{{ title }}</div><div>{{ description }}</div><slot /></div>',
            props: ['title', 'description'],
          },
          AppLoading: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('当前还没有题解')
    expect(wrapper.text()).toContain('双节点演练')

    const createButton = wrapper.findAll('button').find((button) => button.text().includes('编写题解'))
    expect(createButton).toBeTruthy()

    await createButton!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      path: '/platform/challenges/11/writeup',
    })
  })

  it('题解投稿应显示作者与学号，并支持分页翻页', async () => {
    adminApiMocks.getChallengeWriteup.mockResolvedValue(null)
    teacherApiMocks.getTeacherWriteupSubmissions
      .mockResolvedValueOnce({
        list: [
          {
            id: '31',
            user_id: '401',
            student_username: 'alice_01',
            student_name: '张三',
            student_no: '20240001',
            class_name: '网安 1 班',
            challenge_id: '11',
            challenge_title: '双节点演练',
            title: '利用链梳理',
            content_preview: '从入口点开始拆解',
            submission_status: 'published',
            visibility_status: 'visible',
            is_recommended: true,
            published_at: '2026-03-10T01:00:00.000Z',
            updated_at: '2026-03-10T02:30:00.000Z',
          },
        ],
        total: 8,
        page: 1,
        page_size: 6,
      })
      .mockResolvedValueOnce({
        list: [
          {
            id: '32',
            user_id: '402',
            student_username: 'bob_02',
            student_name: '',
            student_no: '',
            class_name: '',
            challenge_id: '11',
            challenge_title: '双节点演练',
            title: '二次利用记录',
            content_preview: '从回显点继续推进',
            submission_status: 'draft',
            visibility_status: 'hidden',
            is_recommended: false,
            published_at: undefined,
            updated_at: '2026-03-11T04:30:00.000Z',
          },
        ],
        total: 8,
        page: 2,
        page_size: 6,
      })

    const wrapper = mount(ChallengeWriteupManagePanel, {
      props: {
        challengeId: '11',
        challengeTitle: '双节点演练',
      },
      global: {
        stubs: {
          AppEmpty: {
            template:
              '<div><div>{{ title }}</div><div>{{ description }}</div><slot /></div>',
            props: ['title', 'description'],
          },
          AppLoading: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('.writeup-manage-header').exists()).toBe(true)
    expect(wrapper.find('.writeup-manage-header .workspace-tab-heading').exists()).toBe(true)
    expect(wrapper.find('.writeup-manage-header .writeup-manage-actions').exists()).toBe(true)
    expect(wrapper.find('.writeup-manage-header + .writeup-manage-stats-shell').exists()).toBe(true)
    expect(wrapper.text()).toContain('题解列表')
    expect(wrapper.text()).toContain('来源')
    expect(wrapper.find('.writeup-manage-section__copy').exists()).toBe(false)
    expect(wrapper.findAll('.writeup-manage-stat')).toHaveLength(2)
    expect(wrapper.text()).toContain('官方题解0篇')
    expect(wrapper.text()).toContain('学员题解8篇')
    expect(wrapper.text()).toContain('学员')
    expect(wrapper.text()).toContain('张三')
    expect(wrapper.text()).toContain('20240001')
    expect(wrapper.text()).toContain('共 8 篇题解')
    expect(wrapper.text()).toContain('1 / 2')

    const nextPageButton = wrapper.findAll('button').find((button) => button.text().trim() === '下一页')
    expect(nextPageButton).toBeTruthy()

    await nextPageButton!.trigger('click')
    await flushPromises()

    expect(teacherApiMocks.getTeacherWriteupSubmissions).toHaveBeenNthCalledWith(2, {
      challenge_id: '11',
      page: 2,
      page_size: 6,
    })
    expect(wrapper.text()).toContain('bob_02')
    expect(wrapper.text()).toContain('未设置学号')
    expect(wrapper.text()).toContain('草稿')
    expect(wrapper.text()).toContain('已隐藏')
    expect(wrapper.text()).toContain('2 / 2')
  })
})
