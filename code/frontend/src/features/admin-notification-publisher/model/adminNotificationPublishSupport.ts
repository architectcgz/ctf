import type { AdminNotificationPublishPayload } from '@/api/contracts'
import type { UserRole } from '@/utils/constants'

export type NotificationAudienceTarget = 'all' | 'role' | 'class' | 'user'

export interface NotificationPublishFormDraft {
  type: AdminNotificationPublishPayload['type']
  title: string
  content: string
  link: string
}

export interface NotificationPublishErrors {
  title?: string
  content?: string
  audience?: string
}

export function createDefaultNotificationPublishForm(): NotificationPublishFormDraft {
  return {
    type: 'system',
    title: '',
    content: '',
    link: '',
  }
}

function uniqueValues(values: string[]): string[] {
  return Array.from(
    new Set(values.map((value) => value.trim()).filter((value) => value.length > 0))
  )
}

interface BuildPayloadOptions {
  form: NotificationPublishFormDraft
  audienceTarget: NotificationAudienceTarget
  selectedRoles: UserRole[]
  selectedClasses: string[]
  selectedUserIds: string[]
}

export function buildAdminNotificationPublishPayload({
  form,
  audienceTarget,
  selectedRoles,
  selectedClasses,
  selectedUserIds,
}: BuildPayloadOptions): AdminNotificationPublishPayload {
  const trimmedTitle = form.title.trim()
  const trimmedContent = form.content.trim()
  const trimmedLink = form.link.trim()

  if (audienceTarget === 'role') {
    return {
      type: form.type,
      title: trimmedTitle,
      content: trimmedContent,
      link: trimmedLink || undefined,
      audience_rules: {
        mode: 'union',
        rules: [{ type: 'role', values: uniqueValues(selectedRoles) as UserRole[] }],
      },
    }
  }

  if (audienceTarget === 'class') {
    return {
      type: form.type,
      title: trimmedTitle,
      content: trimmedContent,
      link: trimmedLink || undefined,
      audience_rules: {
        mode: 'union',
        rules: [{ type: 'class', values: uniqueValues(selectedClasses) }],
      },
    }
  }

  if (audienceTarget === 'user') {
    return {
      type: form.type,
      title: trimmedTitle,
      content: trimmedContent,
      link: trimmedLink || undefined,
      audience_rules: {
        mode: 'union',
        rules: [{ type: 'user', values: uniqueValues(selectedUserIds) }],
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

interface ValidateOptions {
  form: NotificationPublishFormDraft
  audienceTarget: NotificationAudienceTarget
  selectedRoles: UserRole[]
  selectedClasses: string[]
  selectedUserIds: string[]
}

export function validateNotificationPublishForm({
  form,
  audienceTarget,
  selectedRoles,
  selectedClasses,
  selectedUserIds,
}: ValidateOptions): NotificationPublishErrors {
  const errors: NotificationPublishErrors = {}

  if (form.title.trim().length === 0) {
    errors.title = '请输入通知标题。'
  }

  if (form.content.trim().length === 0) {
    errors.content = '请输入通知内容。'
  }

  if (audienceTarget === 'role' && uniqueValues(selectedRoles).length === 0) {
    errors.audience = '请至少选择一个角色。'
  }

  if (audienceTarget === 'class' && uniqueValues(selectedClasses).length === 0) {
    errors.audience = '请至少选择一个班级。'
  }

  if (audienceTarget === 'user' && uniqueValues(selectedUserIds).length === 0) {
    errors.audience = '请至少选择一个用户。'
  }

  return errors
}
