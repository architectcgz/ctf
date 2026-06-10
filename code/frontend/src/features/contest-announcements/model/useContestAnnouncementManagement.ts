import { computed, reactive, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

import {
  createAdminContestAnnouncement,
  deleteAdminContestAnnouncement,
  getAdminContestAnnouncements,
  getAdminContestAnnouncementSync,
} from '@/api/admin/contest-announcements'
import type { ContestAnnouncement, ContestDetailData } from '@/api/contracts'
import { ApiError } from '@/api/request'
import {
  applyContestAnnouncementSyncEvents,
  nextContestAnnouncementSyncCursor,
} from '@/entities/contest-announcement'
import { useToast } from '@/shared/model/common/useToast'

interface AnnouncementFormDraft {
  title: string
  content: string
}

interface AnnouncementFormErrors {
  title?: string
  content?: string
}

function createDefaultForm(): AnnouncementFormDraft {
  return {
    title: '',
    content: '',
  }
}

function humanizeRequestError(error: unknown, fallback: string): string {
  if (error instanceof ApiError && error.message.trim()) {
    return error.message
  }
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

export function useContestAnnouncementManagement(
  contest: MaybeRefOrGetter<ContestDetailData | null | undefined>
) {
  const toast = useToast()

  const announcements = ref<ContestAnnouncement[]>([])
  const loading = ref(false)
  const loadError = ref('')
  const publishing = ref(false)
  const deletingAnnouncementId = ref<string | null>(null)
  const syncCursor = ref('')

  const form = reactive<AnnouncementFormDraft>(createDefaultForm())
  const errors = reactive<AnnouncementFormErrors>({})

  const currentContest = computed(() => toValue(contest) ?? null)
  const contestId = computed(() => currentContest.value?.id ?? '')
  const canManageAnnouncements = computed(() => currentContest.value?.status !== 'ended')

  function clearErrors(): void {
    errors.title = undefined
    errors.content = undefined
  }

  function resetForm(): void {
    const defaults = createDefaultForm()
    form.title = defaults.title
    form.content = defaults.content
    clearErrors()
  }

  function validate(): boolean {
    clearErrors()

    if (!form.title.trim()) {
      errors.title = '请输入公告标题。'
    }

    if (!form.content.trim()) {
      errors.content = '请输入公告内容。'
    }

    return !errors.title && !errors.content
  }

  async function loadAnnouncements(): Promise<ContestAnnouncement[]> {
    if (!contestId.value) {
      announcements.value = []
      loadError.value = ''
      syncCursor.value = ''
      return []
    }

    loading.value = true
    loadError.value = ''
    try {
      const sync = await getAdminContestAnnouncementSync(contestId.value)
      const result = await getAdminContestAnnouncements(contestId.value)
      announcements.value = result
      syncCursor.value = nextContestAnnouncementSyncCursor(sync)
      return result
    } catch (error) {
      const message = humanizeRequestError(error, '公告加载失败，请稍后重试。')
      announcements.value = []
      loadError.value = message
      syncCursor.value = ''
      return []
    } finally {
      loading.value = false
    }
  }

  async function syncAnnouncementsIncrementally(): Promise<ContestAnnouncement[]> {
    if (!contestId.value) {
      syncCursor.value = ''
      return announcements.value
    }

    try {
      if (!syncCursor.value) {
        if (announcements.value.length > 0) {
          return loadAnnouncements()
        }
        const sync = await getAdminContestAnnouncementSync(contestId.value)
        syncCursor.value = nextContestAnnouncementSyncCursor(sync)
        return announcements.value
      }

      let afterId = syncCursor.value
      while (afterId) {
        const sync = await getAdminContestAnnouncementSync(contestId.value, afterId)
        announcements.value = applyContestAnnouncementSyncEvents(announcements.value, sync.events)
        syncCursor.value = nextContestAnnouncementSyncCursor(sync)
        if (!sync.has_more) {
          break
        }
        afterId = syncCursor.value
      }
      loadError.value = ''
      return announcements.value
    } catch (error) {
      loadError.value = humanizeRequestError(error, '公告同步失败，请稍后重试。')
      return announcements.value
    }
  }

  async function publishAnnouncement(): Promise<ContestAnnouncement | null> {
    if (!contestId.value || !canManageAnnouncements.value) {
      return null
    }

    if (!validate()) {
      return null
    }

    publishing.value = true
    try {
      const created = await createAdminContestAnnouncement(contestId.value, {
        title: form.title.trim(),
        content: form.content.trim(),
      })
      resetForm()
      await loadAnnouncements()
      toast.success('公告已发布')
      return created
    } catch (error) {
      toast.error(humanizeRequestError(error, '公告发布失败，请稍后重试。'))
      return null
    } finally {
      publishing.value = false
    }
  }

  async function deleteAnnouncement(announcementId: string): Promise<boolean> {
    if (!contestId.value || !canManageAnnouncements.value) {
      return false
    }

    deletingAnnouncementId.value = announcementId
    try {
      await deleteAdminContestAnnouncement(contestId.value, announcementId)
      await loadAnnouncements()
      toast.success('公告已删除')
      return true
    } catch (error) {
      toast.error(humanizeRequestError(error, '公告删除失败，请稍后重试。'))
      return false
    } finally {
      if (deletingAnnouncementId.value === announcementId) {
        deletingAnnouncementId.value = null
      }
    }
  }

  watch(
    () => contestId.value,
    () => {
      announcements.value = []
      loadError.value = ''
      syncCursor.value = ''
      resetForm()
      deletingAnnouncementId.value = null
    }
  )

  return {
    announcements,
    loading,
    loadError,
    publishing,
    deletingAnnouncementId,
    form,
    errors,
    canManageAnnouncements,
    loadAnnouncements,
    syncAnnouncementsIncrementally,
    publishAnnouncement,
    deleteAnnouncement,
    resetForm,
  }
}
