import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'

import ChallengeDetail from '../ChallengeDetail.vue'
import challengeDetailSource from '../ChallengeDetail.vue?raw'

const challengeApiMocks = vi.hoisted(() => ({
  getChallengeDetail: vi.fn(),
  getChallengeWriteup: vi.fn(),
  getRecommendedChallengeSolutions: vi.fn(),
  getCommunityChallengeSolutions: vi.fn(),
  getMyChallengeWriteupSubmission: vi.fn(),
  upsertChallengeWriteupSubmission: vi.fn(),
  submitFlag: vi.fn(),
  unlockHint: vi.fn(),
  createInstance: vi.fn(),
  downloadAttachment: vi.fn(),
}))

const instanceApiMocks = vi.hoisted(() => ({
  getMyInstances: vi.fn(),
  destroyInstance: vi.fn(),
  extendInstance: vi.fn(),
  requestInstanceAccess: vi.fn(),
}))

vi.mock('@/api/challenge', () => challengeApiMocks)
vi.mock('@/api/instance', () => instanceApiMocks)

describe('ChallengeDetail', () => {
  let router: ReturnType<typeof createRouter>

  beforeEach(() => {
    router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/challenges/:id', component: { template: '<div />' } }],
    })

    challengeApiMocks.getChallengeDetail.mockResolvedValue({
      id: '1',
      title: 'Test Challenge',
      description: '<p>Test description</p>',
      category: 'web',
      difficulty: 'easy',
      tags: ['test'],
      points: 100,
      need_target: true,
      is_solved: false,
      attachment_url: 'https://example.com/file.zip',
      hints: [
        {
          id: 'hint-1',
          level: 1,
          title: '入口',
          content: '先观察登录表单的参数。',
        },
      ],
    })
    challengeApiMocks.getChallengeWriteup.mockResolvedValue({
      id: 'writeup-1',
      challenge_id: '1',
      title: '官方题解',
      content: '<p>Exploit path</p>',
      visibility: 'public',
      is_released: true,
      requires_spoiler_warning: true,
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T01:00:00.000Z',
    })
    challengeApiMocks.getRecommendedChallengeSolutions.mockResolvedValue([
      {
        id: 'recommended-1',
        source_type: 'official',
        source_id: 'writeup-1',
        challenge_id: '1',
        title: '精选官方题解',
        content: '<p>Exploit path</p>',
        author_name: '官方题解',
        is_recommended: true,
        recommended_at: '2026-03-10T02:00:00.000Z',
        updated_at: '2026-03-10T02:00:00.000Z',
      },
    ])
    challengeApiMocks.getCommunityChallengeSolutions.mockResolvedValue({
      list: [
        {
          id: 'community-1',
          challenge_id: '1',
          user_id: 'stu-2',
          title: '我的 SQLi 复盘',
          content: '先找注入点，再构造 payload。',
          content_preview: '先找注入点，再构造 payload。',
          author_name: 'student_b',
          submission_status: 'published',
          visibility_status: 'visible',
          is_recommended: false,
          published_at: '2026-03-12T01:00:00.000Z',
          updated_at: '2026-03-12T01:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    challengeApiMocks.getMyChallengeWriteupSubmission.mockResolvedValue(null)
    challengeApiMocks.upsertChallengeWriteupSubmission.mockResolvedValue({
      id: 'submission-1',
      user_id: 'stu-1',
      challenge_id: '1',
      title: '我的复盘',
      content: '先找回显，再定位注入。',
      submission_status: 'draft',
      visibility_status: 'visible',
      is_recommended: false,
      created_at: '2026-03-12T00:00:00.000Z',
      updated_at: '2026-03-12T00:30:00.000Z',
    })
    challengeApiMocks.submitFlag.mockReset()
    challengeApiMocks.unlockHint.mockReset()
    challengeApiMocks.createInstance.mockResolvedValue({
      id: 'inst-1',
      challenge_id: '1',
      status: 'running',
      access_url: 'http://target.test',
      flag_type: 'static',
      expires_at: '2099-01-01T00:00:00Z',
      remaining_extends: 2,
      created_at: '2026-03-12T00:00:00.000Z',
    })
    challengeApiMocks.downloadAttachment.mockReset()

    instanceApiMocks.getMyInstances.mockResolvedValue([])
    instanceApiMocks.destroyInstance.mockReset()
    instanceApiMocks.extendInstance.mockReset()
    instanceApiMocks.requestInstanceAccess.mockReset()
  })

  afterEach(() => {
    vi.clearAllTimers()
    vi.useRealTimers()
  })

  it('应该渲染挑战详情', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.element.tagName).toBe('SECTION')
    expect(wrapper.classes()).toContain('journal-shell')
    expect(wrapper.classes()).toContain('journal-hero')
    expect(wrapper.classes()).toContain('min-h-full')
    expect(wrapper.text()).toContain('题目')
    expect(wrapper.text()).toContain('题解')
    expect(wrapper.text()).toContain('提交记录')
    expect(wrapper.text()).toContain('我的复盘')
    expect(wrapper.text()).toContain('Test Challenge')
    expect(wrapper.text()).toContain('题目描述')
  })

  it('应仅保留外层主容器卡片并移除内部二级卡片', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.findAll('.challenge-panel')).toHaveLength(0)
  })

  it('工作区应建立满高伸展布局链', () => {
    expect(challengeDetailSource).toContain('min-height: max(100%, calc(100vh - 5rem));')
    expect(challengeDetailSource).toContain('.detail-content {\n  display: flex;\n  flex: 1 1 auto;')
    expect(challengeDetailSource).toMatch(/\.detail-grid,\s*\.workspace-grid\s*{\s*display:\s*grid;\s*flex:\s*1 1 auto;/)
  })

  it('未解题时应显示题解锁定态', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const solutionTab = wrapper.findAll('button').find((node) => node.text().trim() === '题解')
    expect(solutionTab).toBeTruthy()

    await solutionTab!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('解出题目后可查看推荐题解与社区题解')
    expect(wrapper.text()).not.toContain('精选官方题解')
    expect(challengeApiMocks.getRecommendedChallengeSolutions).not.toHaveBeenCalled()
    expect(challengeApiMocks.getCommunityChallengeSolutions).not.toHaveBeenCalled()
  })

  it('已解题时应通过顶部标签切换到推荐题解、社区题解和我的复盘', async () => {
    challengeApiMocks.getChallengeDetail.mockResolvedValueOnce({
      id: '1',
      title: 'Solved Challenge',
      description: '<p>Test description</p>',
      category: 'web',
      difficulty: 'easy',
      tags: ['test'],
      points: 100,
      need_target: true,
      is_solved: true,
      attachment_url: 'https://example.com/file.zip',
      hints: [],
    })

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(challengeApiMocks.getRecommendedChallengeSolutions).toHaveBeenCalledWith('1')
    expect(challengeApiMocks.getCommunityChallengeSolutions).toHaveBeenCalledWith('1')
    const solutionTab = wrapper.findAll('button').find((node) => node.text().trim() === '题解')
    expect(solutionTab).toBeTruthy()

    await solutionTab!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('推荐题解')
    expect(wrapper.text()).toContain('社区题解')
    expect(wrapper.text()).toContain('精选官方题解')

    const communityTab = wrapper.findAll('button').find((node) => node.text().trim() === '社区题解')
    expect(communityTab).toBeTruthy()

    await communityTab!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('我的 SQLi 复盘')

    const writeupTab = wrapper.findAll('button').find((node) => node.text().trim() === '我的复盘')
    expect(writeupTab).toBeTruthy()

    await writeupTab!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('解题过程复盘')
  })

  it('顶部主切换应暴露 tabs 语义，题解页签下仍保留次级 tabs 语义', async () => {
    challengeApiMocks.getChallengeDetail.mockResolvedValueOnce({
      id: '1',
      title: 'Solved Challenge',
      description: '<p>Test description</p>',
      category: 'web',
      difficulty: 'easy',
      tags: ['test'],
      points: 100,
      need_target: true,
      is_solved: true,
      attachment_url: 'https://example.com/file.zip',
      hints: [],
    })

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const tablists = wrapper.findAll('[role="tablist"]')
    expect(tablists).toHaveLength(1)

    const topTabs = wrapper.findAll('[role="tab"]')
    expect(topTabs).toHaveLength(4)
    expect(topTabs[0].attributes('aria-selected')).toBe('true')
    expect(topTabs[1].attributes('aria-selected')).toBe('false')

    await topTabs[1].trigger('click')
    await wrapper.vm.$nextTick()

    const nestedTablists = wrapper.findAll('[role="tablist"]')
    expect(nestedTablists.length).toBeGreaterThanOrEqual(2)
    expect(wrapper.find('[role="tabpanel"]').exists()).toBe(true)
  })

  it('应该支持保存个人题解草稿', async () => {
    challengeApiMocks.getChallengeDetail.mockResolvedValueOnce({
      id: '1',
      title: 'Solved Challenge',
      description: '<p>Test description</p>',
      category: 'web',
      difficulty: 'easy',
      tags: ['test'],
      points: 100,
      need_target: true,
      is_solved: true,
      attachment_url: 'https://example.com/file.zip',
      hints: [],
    })

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const writeupTab = wrapper.findAll('button').find((node) => node.text().trim() === '我的复盘')
    expect(writeupTab).toBeTruthy()
    await writeupTab!.trigger('click')
    await wrapper.vm.$nextTick()

    const titleInput = wrapper.find('input[placeholder*="完整链路"]')
    const contentInput = wrapper.find('textarea')
    const draftButton = wrapper.findAll('button').find((node) => node.text().trim() === '保存草稿')

    await titleInput.setValue('我的复盘')
    await contentInput.setValue('先找回显，再定位注入。')
    await draftButton!.trigger('click')
    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 0))

    expect(challengeApiMocks.upsertChallengeWriteupSubmission).toHaveBeenCalledWith('1', {
      title: '我的复盘',
      content: '先找回显，再定位注入。',
      submission_status: 'draft',
    })
    expect(wrapper.text()).toContain('草稿')
  })

  it('我的复盘应通过顶部标签切换进入，默认不显示编辑区', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).not.toContain('解题过程复盘')
    expect(wrapper.find('input[placeholder*="完整链路"]').exists()).toBe(false)

    const writeupTab = wrapper.findAll('button').find((node) => node.text().trim() === '我的复盘')
    expect(writeupTab).toBeTruthy()

    await writeupTab!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('解题过程复盘')
    expect(wrapper.find('input[placeholder*="完整链路"]').exists()).toBe(true)
  })

  it('只有题目标签显示右侧工具区，其他标签应切换为单栏内容', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).toContain('Flag 提交')

    const solutionTab = wrapper.findAll('button').find((node) => node.text().trim() === '题解')
    const recordsTab = wrapper.findAll('button').find((node) => node.text().trim() === '提交记录')
    const writeupTab = wrapper.findAll('button').find((node) => node.text().trim() === '我的复盘')

    expect(solutionTab).toBeTruthy()
    expect(recordsTab).toBeTruthy()
    expect(writeupTab).toBeTruthy()

    await solutionTab!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).not.toContain('Flag 提交')
    expect(wrapper.text()).toContain('题解区')

    await recordsTab!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).not.toContain('Flag 提交')
    expect(wrapper.text()).toContain('提交记录')

    await writeupTab!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).not.toContain('Flag 提交')
    expect(wrapper.text()).toContain('解题过程复盘')
  })

  it('只有切到题目标签时才显示题目基本信息，题解标签下不重复显示题头', async () => {
    challengeApiMocks.getChallengeDetail.mockResolvedValueOnce({
      id: '1',
      title: 'Solved Challenge',
      description: '<p>Test description</p>',
      category: 'web',
      difficulty: 'easy',
      tags: ['test'],
      points: 100,
      need_target: true,
      is_solved: true,
      attachment_url: 'https://example.com/file.zip',
      hints: [],
    })

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).toContain('Solved Challenge')
    expect(wrapper.text()).toContain('分值')

    const solutionTab = wrapper.findAll('button').find((node) => node.text().trim() === '题解')
    expect(solutionTab).toBeTruthy()

    await solutionTab!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).not.toContain('Solved Challenge')
    expect(wrapper.text()).not.toContain('分值')
    expect(wrapper.text()).toContain('推荐题解')
  })

  it('提示内容应支持前端展开查看且不再调用解锁接口', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).not.toContain('先观察登录表单的参数。')
    expect(wrapper.text()).not.toContain('解锁提示')

    const toggleButton = wrapper.find('button.hint-toggle')
    expect(toggleButton.exists()).toBe(true)
    expect(toggleButton.text()).toContain('展开提示')

    await toggleButton.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('先观察登录表单的参数。')
    expect(challengeApiMocks.unlockHint).not.toHaveBeenCalled()
  })

  it('manual review 提交后应显示待审核反馈', async () => {
    challengeApiMocks.submitFlag.mockResolvedValue({
      is_correct: false,
      status: 'pending_review',
      message: '答案已提交，等待教师审核',
      submitted_at: '2026-03-12T01:00:00.000Z',
    })

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const flagInput = wrapper.find('input[placeholder="flag{...}"]')
    const submitButton = wrapper.findAll('button').find((node) => node.text().trim() === '提交')

    await flagInput.setValue('exploit chain answer')
    await submitButton!.trigger('click')
    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 0))

    expect(challengeApiMocks.submitFlag).toHaveBeenCalledWith('1', 'exploit chain answer')
    expect(wrapper.text()).toContain('答案已提交，等待教师审核')
    expect(wrapper.text()).not.toContain('已完成 ✓')
  })

  it('Flag 输入应提供可访问标签', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const flagInput = wrapper.find('input[aria-label="Flag"]')
    expect(flagInput.exists()).toBe(true)
  })

  it('启动靶机后应停留在题目页并显示实例卡片', async () => {
    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const startButton = wrapper.findAll('button').find((node) => node.text().includes('启动靶机'))
    expect(startButton).toBeTruthy()

    await startButton!.trigger('click')
    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 0))

    expect(challengeApiMocks.createInstance).toHaveBeenCalledWith('1')
    expect(router.currentRoute.value.fullPath).toBe('/challenges/1')
    expect(wrapper.text()).toContain('打开目标')
    expect(wrapper.text()).toContain('靶机实例')
  })

  it('createInstance 返回 pending 时应显示等待文案并触发轮询', async () => {
    vi.useFakeTimers()
    challengeApiMocks.createInstance.mockResolvedValueOnce({
      id: 'inst-1',
      challenge_id: '1',
      status: 'pending',
      access_url: '',
      flag_type: 'static',
      expires_at: '2099-01-01T00:00:00Z',
      remaining_extends: 2,
      created_at: '2026-03-12T00:00:00.000Z',
      queue_position: 3,
      eta_seconds: 120,
      progress: 18,
    })
    instanceApiMocks.getMyInstances.mockReset()
    instanceApiMocks.getMyInstances.mockResolvedValueOnce([]).mockResolvedValueOnce([
      {
        id: 'inst-1',
        challenge_id: 1,
        status: 'running',
        access_url: 'http://127.0.0.1:30000',
        flag_type: 'static',
        expires_at: '2099-01-01T00:00:00Z',
        remaining_extends: 2,
        created_at: '2026-03-12T00:00:00.000Z',
        challenge_title: 'Test Challenge',
        category: 'web',
        difficulty: 'easy',
      },
    ])

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await vi.advanceTimersByTimeAsync(100)

    const startButton = wrapper.findAll('button').find((node) => node.text().includes('启动靶机'))
    expect(startButton).toBeTruthy()

    await startButton!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('实例正在排队创建')
    expect(wrapper.text()).toContain('等待实例就绪')
    expect(instanceApiMocks.getMyInstances).toHaveBeenCalledTimes(1)

    await vi.advanceTimersByTimeAsync(3000)
    await wrapper.vm.$nextTick()

    expect(instanceApiMocks.getMyInstances).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('打开目标')
  })

  it('已存在实例时应直接显示实例信息', async () => {
    instanceApiMocks.getMyInstances.mockResolvedValue([
      {
        id: 'inst-9',
        challenge_id: 1,
        status: 'running',
        access_url: 'http://127.0.0.1:30000',
        flag_type: 'static',
        expires_at: '2099-01-01T00:00:00Z',
        remaining_extends: 1,
        created_at: '2026-03-12T00:00:00.000Z',
        challenge_title: 'Test Challenge',
        category: 'web',
        difficulty: 'easy',
      },
    ])

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).toContain('打开目标')
    expect(wrapper.text()).toContain('http://127.0.0.1:30000')
    expect(wrapper.text()).toContain('1 次')
    expect(wrapper.text()).not.toContain('挑战信息')
    expect(wrapper.text()).not.toContain('启动靶机')
  })

  it('题目不需要靶机时应展示提示文案', async () => {
    challengeApiMocks.getChallengeDetail.mockResolvedValueOnce({
      id: '1',
      title: 'No Target Challenge',
      description: '<p>Analyze only</p>',
      category: 'misc',
      difficulty: 'easy',
      tags: ['misc'],
      points: 50,
      need_target: false,
      is_solved: false,
      hints: [],
    })

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).toContain('该题目不需要靶机')
    expect(wrapper.text()).not.toContain('启动靶机')
  })

  it('应将 markdown 描述渲染为 HTML', async () => {
    challengeApiMocks.getChallengeDetail.mockResolvedValueOnce({
      id: '1',
      title: 'Markdown Challenge',
      description: '# 一级标题\n\n## 二级标题\n\n- item-1',
      category: 'misc',
      difficulty: 'easy',
      tags: ['misc'],
      points: 50,
      need_target: false,
      is_solved: false,
      hints: [],
    })

    await router.push('/challenges/1')
    await router.isReady()

    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    const content = wrapper.find('.prose')
    expect(content.html()).toContain('<h1')
    expect(content.html()).toContain('<h2')
    expect(content.html()).toContain('<li')
  })
})
