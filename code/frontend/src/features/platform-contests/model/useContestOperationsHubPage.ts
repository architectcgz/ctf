import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getContests } from '@/api/admin/contests'
import type { ContestDetailData } from '@/api/contracts'

export function useContestOperationsHubPage() {
  const router = useRouter()

  const loading = ref(true)
  const loadError = ref('')
  const contests = ref<ContestDetailData[]>([])

  const awdContests = computed(() => contests.value.filter((item) => item.mode === 'awd'))
  const operableContests = computed(() =>
    awdContests.value.filter((item) =>
      ['running', 'frozen', 'ended', 'registering'].includes(item.status)
    )
  )
  const runningContestCount = computed(
    () => operableContests.value.filter((item) => item.status === 'running').length
  )
  const frozenContestCount = computed(
    () => operableContests.value.filter((item) => item.status === 'frozen').length
  )
  const preferredContest = computed(
    () =>
      operableContests.value.find((item) => item.status === 'running') ||
      operableContests.value.find((item) => item.status === 'frozen') ||
      operableContests.value[0] ||
      null
  )

  async function loadContests(): Promise<void> {
    loading.value = true
    loadError.value = ''

    try {
      const response = await getContests({
        page: 1,
        page_size: 100,
      })
      contests.value = response.list
    } catch (error) {
      contests.value = []
      loadError.value = error instanceof Error ? error.message : '赛事运维目录加载失败'
    } finally {
      loading.value = false
    }
  }

  async function handleEnterOperations(contestId: string): Promise<void> {
    await router.push({
      name: 'ContestOperations',
      params: { id: contestId },
    })
  }

  async function handleBackToContestDirectory(): Promise<void> {
    await router.push({
      name: 'ContestManage',
      query: { panel: 'list' },
    })
  }

  onMounted(() => {
    void loadContests()
  })

  return {
    loading,
    loadError,
    operableContests,
    runningContestCount,
    frozenContestCount,
    preferredContest,
    loadContests,
    handleEnterOperations,
    handleBackToContestDirectory,
  }
}
