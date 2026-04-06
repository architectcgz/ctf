import { reactive, ref } from 'vue'

import { getUsers, publishAdminNotification } from '@/api/admin'
import { getClasses } from '@/api/teacher'
import type {
  AdminNotificationPublishPayload,
  AdminNotificationPublishResult,
  AdminUserListItem,
  TeacherClassItem,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'
import type { UserRole } from '@/utils/constants'

export type NotificationAudienceTarget = 'all' | 'role' | 'class' | 'user'

export interface NotificationPublishFormDraft {
  type: AdminNotificationPublishPayload['type']
  title: string
  content: string
  link: string
}

interface NotificationPublishErrors {
  title?: string
  content?: string
  audience?: string
}

function createDefaultForm(): NotificationPublishFormDraft {
  return {
    type: 'system',
    title: '',
    content: '',
    link: '',
  }
}

function uniqueValues(values: string[]): string[] {
  return Array.from(new Set(values.map((value) => value.trim()).filter((value) => value.length > 0)))
}

export function useAdminNotificationPublisher() {
  const toast = useToast()

  const form = reactive<NotificationPublishFormDraft>(createDefaultForm())
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
  }

  function buildPayload(): AdminNotificationPublishPayload {
    const trimmedTitle = form.title.trim()
    const trimmedContent = form.content.trim()
    const trimmedLink = form.link.trim()

    if (audienceTarget.value === 'role') {
      return {
        type: form.type,
        title: trimmedTitle,
        content: trimmedContent,
        link: trimmedLink || undefined,
        audience_rules: {
          mode: 'union',
          rules: [{ type: 'role', values: uniqueValues(selectedRoles.value) as UserRole[] }],
        },
      }
    }

    if (audienceTarget.value === 'class') {
      return {
        type: form.type,
        title: trimmedTitle,
        content: trimmedContent,
        link: trimmedLink || undefined,
        audience_rules: {
          mode: 'union',
          rules: [{ type: 'class', values: uniqueValues(selectedClasses.value) }],
        },
      }
    }

    if (audienceTarget.value === 'user') {
      return {
        type: form.type,
        title: trimmedTitle,
        content: trimmedContent,
        link: trimmedLink || undefined,
        audience_rules: {
          mode: 'union',
          rules: [{ type: 'user', values: uniqueValues(selectedUserIds.value) }],
        },
      }
    }

    return {
      type: form.type,
      title: trimmedTitle,
      content: trimmedContent,
      link: trimmedLink || undefined,
      audience_rules: {
        mode: 'union',
        rules: [{ type: 'all' }],
      },
    }
  }

  function validate(): boolean {
    clearErrors()

    if (form.title.trim().length === 0) {
      errors.title = '请输入通知标题。'
    }

    if (form.content.trim().length === 0) {
      errors.content = '请输入通知内容。'
    }

    if (audienceTarget.value === 'role' && uniqueValues(selectedRoles.value).length === 0) {
      errors.audience = '请至少选择一个角色。'
    }

    if (audienceTarget.value === 'class' && uniqueValues(selectedClasses.value).length === 0) {
      errors.audience = '请至少选择一个班级。'
    }

    if (audienceTarget.value === 'user' && uniqueValues(selectedUserIds.value).length === 0) {
      errors.audience = '请至少选择一个用户。'
    }

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
      latestUserSearchRequestID += 1
      userOptions.value = []
      loadingUsers.value = false
      return
    }

    const requestID = ++latestUserSearchRequestID
    loadingUsers.value = true
    try {
      const response = await getUsers({
        page: 1,
        page_size: 20,
        keyword: normalizedKeyword,
      })
      if (requestID !== latestUserSearchRequestID) return
      userOptions.value = response.list
    } finally {
      if (requestID !== latestUserSearchRequestID) return
      loadingUsers.value = false
    }
  }

  function reset(): void {
    const initial = createDefaultForm()
    form.type = initial.type
    form.title = initial.title
    form.content = initial.content
    form.link = initial.link
    audienceTarget.value = 'all'
    classOptions.value = []
    userOptions.value = []
    userKeyword.value = ''
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
      toast.error(error instanceof Error && error.message ? error.message : '通知发布失败，请稍后重试。')
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
