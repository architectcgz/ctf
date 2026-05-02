import { beforeEach, describe, expect, it, vi } from 'vitest'

const authMocks = vi.hoisted(() => ({
  register: vi.fn(),
}))

vi.mock('./useAuth', () => ({
  useAuth: () => authMocks,
}))

import { useRegisterPage } from './useRegisterPage'

describe('useRegisterPage', () => {
  beforeEach(() => {
    authMocks.register.mockReset()
  })

  it('应提交注册并清理班级空白字符', async () => {
    authMocks.register.mockResolvedValue(undefined)

    const page = useRegisterPage()
    page.form.username = 'alice'
    page.form.password = 'secure-pass'
    page.form.class_name = '  CTF-1  '

    await page.onSubmit()

    expect(authMocks.register).toHaveBeenCalledWith({
      username: 'alice',
      password: 'secure-pass',
      class_name: 'CTF-1',
    })
    expect(page.loading.value).toBe(false)
    expect(page.submitError.value).toBe('')
  })

  it('缺少必填字段时不应提交', async () => {
    const page = useRegisterPage()
    page.form.username = 'alice'

    await page.onSubmit()

    expect(authMocks.register).not.toHaveBeenCalled()
  })

  it('应回填注册失败错误', async () => {
    authMocks.register.mockRejectedValue(new Error('用户名已存在'))

    const page = useRegisterPage()
    page.form.username = 'alice'
    page.form.password = 'secure-pass'

    await page.onSubmit()

    expect(page.submitError.value).toBe('用户名已存在')
    expect(page.loading.value).toBe(false)
  })
})
