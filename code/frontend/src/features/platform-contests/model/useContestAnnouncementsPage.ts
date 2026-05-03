import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getContest } from '@/api/admin/contests'
import type { ContestDetailData } from '@/api/contracts'
import { ApiError } from '@/api/request'
import { useContestAnnouncementManagement } from '@/features/contest-announcements'

export function useContestAnnouncementsPage() {
  const route = useRoute()
  const router = useRouter()

  const contestId = computed(() => String(route.params.id ?? ''))
  const contest = ref<ContestDetailData | null>(null)
  const loading = ref(true)
  const loadError = ref('')

  const management = useContestAnnouncementManagement(computed(() => contest.value))

  function humanizeRequestError(error: unknown, fallback: string): string {
    if (error instanceof ApiError && error.message.trim()) {
      return error.message
    }
    if (error instanceof Error && error.message.trim()) {
      return error.message
    }
    return fallback
  }

  function formatTime(value: string): string {
    return new Date(value).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  function goBackToStudio(): void {
    void router.push({ name: 'ContestEdit', params: { id: contestId.value } })
  }

  async function loadPage(): Promise<void> {
    if (!contestId.value) {
      loadError.value = '缺少竞赛编号。'
      loading.value = false
      return
    }

    loading.value = true
    loadError.value = ''
    try {
      contest.value = await getContest(contestId.value)
      await management.loadAnnouncements()
    } catch (error) {
      loadError.value = humanizeRequestError(error, '竞赛公告加载失败')
    } finally {
      loading.value = false
    }
  }

  async function handleSubmit(): Promise<void> {
    await management.publishAnnouncement()
  }

  async function handleDelete(announcementId: string): Promise<void> {
    await management.deleteAnnouncement(announcementId)
  }

  onMounted(() => {
    void loadPage()
  })

  return {
    contest,
    loading,
    loadError,
    management,
    formatTime,
    goBackToStudio,
    loadPage,
    handleSubmit,
    handleDelete,
  }
}
