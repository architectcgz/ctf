<template>
  <SecuritySettingsWorkspaceShell
    :password-saving="passwordSaving"
    :password-error="passwordError"
    :password-form="passwordForm"
    :password-field-errors="passwordFieldErrors"
    :security-stats="securityStats"
    :password-tips="passwordTips"
    :submit-password-change="handleSubmitPasswordChange"
  />
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouteNavigationTransport } from '@/shared/model/navigation/useRouteNavigationTransport'
import {
  SecuritySettingsWorkspaceShell,
  useSecuritySettingsPage,
} from '@/features/profile'

const authStore = useAuthStore()
const { push } = useRouteNavigationTransport()

const {
  passwordSaving,
  passwordError,
  passwordForm,
  passwordFieldErrors,
  securityStats,
  passwordTips,
  submitPasswordChange,
} = useSecuritySettingsPage()

async function handleSubmitPasswordChange(): Promise<void> {
  const ok = await submitPasswordChange()
  if (ok) {
    authStore.logout()
    await push('/login')
  }
}
</script>
