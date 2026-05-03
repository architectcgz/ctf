import { reactive, ref } from 'vue'

import { changePassword } from '@/api/auth'
import { useToast } from '@/composables/useToast'

export function useSecuritySettingsPage() {
  const toast = useToast()

  const passwordSaving = ref(false)
  const passwordError = ref<string | null>(null)

  const passwordForm = reactive({
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  })

  const passwordFieldErrors = reactive({
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  })

  const securityStats = [
    {
      key: 'policy',
      label: '密码策略',
      value: '已启用',
      helper: '最少 8 位并建议使用字母与数字混合',
    },
    { key: 'rotation', label: '建议轮换', value: '90 天', helper: '定期更新，降低长期暴露风险' },
    { key: 'session', label: '安全通道', value: '在线', helper: '密码修改请求通过受保护会话提交' },
    { key: 'scope', label: '同步范围', value: '全账号', helper: '更新后其他设备需重新登录验证' },
  ]

  const passwordTips = [
    '不要使用生日、姓名等易猜测信息',
    '避免在多个平台复用同一密码',
    '修改后其他设备需重新登录',
  ]

  function resetPasswordErrors(): void {
    passwordError.value = null
    passwordFieldErrors.oldPassword = ''
    passwordFieldErrors.newPassword = ''
    passwordFieldErrors.confirmPassword = ''
  }

  function validatePasswordForm(): boolean {
    resetPasswordErrors()

    if (!passwordForm.oldPassword.trim()) {
      passwordFieldErrors.oldPassword = '请填写当前密码'
    }

    if (!passwordForm.newPassword.trim()) {
      passwordFieldErrors.newPassword = '请填写新密码'
    } else if (passwordForm.newPassword.trim().length < 8) {
      passwordFieldErrors.newPassword = '新密码至少需要 8 位'
    }

    if (!passwordForm.confirmPassword.trim()) {
      passwordFieldErrors.confirmPassword = '请再次输入新密码'
    } else if (passwordForm.confirmPassword !== passwordForm.newPassword) {
      passwordFieldErrors.confirmPassword = '两次输入的新密码不一致'
    }

    return (
      !passwordFieldErrors.oldPassword &&
      !passwordFieldErrors.newPassword &&
      !passwordFieldErrors.confirmPassword
    )
  }

  async function submitPasswordChange(): Promise<void> {
    if (passwordSaving.value || !validatePasswordForm()) {
      return
    }

    passwordSaving.value = true
    passwordError.value = null
    try {
      await changePassword({
        old_password: passwordForm.oldPassword,
        new_password: passwordForm.newPassword,
      })
      passwordForm.oldPassword = ''
      passwordForm.newPassword = ''
      passwordForm.confirmPassword = ''
      toast.success('密码修改成功')
    } catch (err) {
      console.error('修改密码失败:', err)
      passwordError.value = err instanceof Error ? err.message : '修改密码失败，请稍后重试'
    } finally {
      passwordSaving.value = false
    }
  }

  return {
    passwordSaving,
    passwordError,
    passwordForm,
    passwordFieldErrors,
    securityStats,
    passwordTips,
    submitPasswordChange,
  }
}
