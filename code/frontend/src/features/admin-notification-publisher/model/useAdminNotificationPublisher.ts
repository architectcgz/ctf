import { reactive, ref } from 'vue'

import { publishAdminNotification } from '@/api/admin/platform'
import { getUsers } from '@/api/admin/users'
import { getClasses } from '@/api/teacher'
import { useAbortController } from '@/composables/useAbortController'
import type {
  AdminNotificationPublishPayload,
  AdminNotificationPublishResult,
  AdminUserListItem,
  TeacherClassItem,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'
import type { UserRole } from '@/utils/constants'
import {
  buildAdminNotificationPublishPayload,
  createDefaultNotificationPublishForm,
  validateNotificationPublishForm,
  type NotificationAudienceTarget,
  type NotificationPublishErrors,
  type NotificationPublishFormDraft,
} from './adminNotificationPublishSupport'

export type { NotificationAudienceTarget, NotificationPublishFormDraft }

export function useAdminNotificationPublisher() {
  const toast = useToast()
  const { createController, abort } = useAbortController()

  const form = reactive<NotificationPublishFormDraft>(createDefaultNotificationPublishForm())
  const audienceTarget = ref<NotificationAudienceTarget>('all')
  const selectedRoles = ref<UserRole[]>([])
  const selectedClasses = ref<string[]>([])
  const selectedUserIds = ref<string[]>([])

  const loadingClasses = ref(false)
  const loadingUsers = ref(false)
  const submitting = ref(false)

  const classOptions = ref<TeacherClassItem[]>([])
  const userOptions = ref<AdminUserListItem[]>([])
  const userKeyword = ref('')
  let latestUserSearchRequestID = 0

  const errors = reactive<NotificationPublishErrors>({})

  function isCanceledError(error: unknown): boolean {
    return (
      !!error &&
      typeof error === 'object' &&
      ('code' in error ? (error as { code?: unknown }).code === 'ERR_CANCELED' : false)
    )
  }

  function clearUserSearchState(options?: { preserveKeyword?: boolean }) {
    latestUserSearchRequestID += 1
    abort()
    if (!options?.preserveKeyword) {
      userKeyword.value = ''
    }
    userOptions.value = []
    loadingUsers.value = false
  }

  function clearErrors() {
    errors.title = undefined
    errors.content = undefined
    errors.audience = undefined
  }

  function resetSelections() {
    selectedRoles.value = []
    selectedClasses.value = []
    selectedUserIds.value = []
  }

  function setAudienceTarget(next: NotificationAudienceTarget) {
    audienceTarget.value = next
    resetSelections()
    errors.audience = undefined
    if (next !== 'user') {
      clearUserSearchState()
    }
  }

  function buildPayload(): AdminNotificationPublishPayload {
    return buildAdminNotificationPublishPayload({
      form,
      audienceTarget: audienceTarget.value,
      selectedRoles: selectedRoles.value,
      selectedClasses: selectedClasses.value,
      selectedUserIds: selectedUserIds.value,
    })
  }

  function validate(): boolean {
    clearErrors()
    const validation = validateNotificationPublishForm({
      form,
      audienceTarget: audienceTarget.value,
      selectedRoles: selectedRoles.value,
      selectedClasses: selectedClasses.value,
      selectedUserIds: selectedUserIds.value,
    })
    errors.title = validation.title
    errors.content = validation.content
    errors.audience = validation.audience

    return !errors.title && !errors.content && !errors.audience
  }

  async function loadClasses(): Promise<void> {
    loadingClasses.value = true
    try {
      classOptions.value = await getClasses()
    } finally {
      loadingClasses.value = false
    }
  }

  async function searchUsers(keyword: string): Promise<void> {
    const normalizedKeyword = keyword.trim()
    userKeyword.value = keyword
    if (!normalizedKeyword) {
      clearUserSearchState({ preserveKeyword: true })
      return
    }

    const requestID = ++latestUserSearchRequestID
    const controller = createController()
    loadingUsers.value = true
    try {
      const response = await getUsers({
        page: 1,
        page_size: 20,
        keyword: normalizedKeyword,
      }, {
        signal: controller.signal,
      })
      if (requestID !== latestUserSearchRequestID) return
      userOptions.value = response.list
    } catch (error) {
      if (requestID !== latestUserSearchRequestID) return
      if (isCanceledError(error)) return
      throw error
    } finally {
      if (requestID !== latestUserSearchRequestID) return
      loadingUsers.value = false
    }
  }

  function reset(): void {
    const initial = createDefaultNotificationPublishForm()
    form.type = initial.type
    form.title = initial.title
    form.content = initial.content
    form.link = initial.link
    audienceTarget.value = 'all'
    classOptions.value = []
    clearUserSearchState()
    clearErrors()
    resetSelections()
  }

  async function submit(): Promise<AdminNotificationPublishResult | null> {
    if (!validate()) {
      return null
    }

    submitting.value = true
    try {
      const result = await publishAdminNotification(buildPayload())
      toast.success(`通知发布成功，已投递 ${result.recipient_count} 人。`)
      return result
    } catch (error) {
      toast.error(
        error instanceof Error && error.message ? error.message : '通知发布失败，请稍后重试。'
      )
      return null
    } finally {
      submitting.value = false
    }
  }

  return {
    form,
    audienceTarget,
    selectedRoles,
    selectedClasses,
    selectedUserIds,
    loadingClasses,
    loadingUsers,
    submitting,
    classOptions,
    userOptions,
    userKeyword,
    errors,
    setAudienceTarget,
    buildPayload,
    loadClasses,
    searchUsers,
    submit,
    reset,
  }
}
