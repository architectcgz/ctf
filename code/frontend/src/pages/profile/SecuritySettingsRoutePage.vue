<template>
  <SecuritySettingsWorkspaceShell
    :password-saving="passwordSaving"
    :password-error="passwordError"
    :password-form="passwordForm"
    :password-field-errors="passwordFieldErrors"
    :security-stats="securityStats"
    :password-tips="passwordTips"
    :submit-password-change="submitPasswordChange"
  />
</template>

<script setup lang="ts">
import {
  SecuritySettingsWorkspaceShell,
  useSecuritySettingsPage,
} from '@/features/profile'

const {
  passwordSaving,
  passwordError,
  passwordForm,
  passwordFieldErrors,
  securityStats,
  passwordTips,
  submitPasswordChange,
} = useSecuritySettingsPage()
</script>
