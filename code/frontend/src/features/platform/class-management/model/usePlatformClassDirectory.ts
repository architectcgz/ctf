import { computed, onMounted, ref } from 'vue'

import type { ClassDirectoryItem } from '@/api/contracts'
import { getClasses } from '@/api/admin'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'
import { reportFrontendError } from '@/utils/reportFrontendError'

export function usePlatformClassDirectory() {
  const list = ref<ClassDirectoryItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(DEFAULT_PAGE_SIZE)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const totalPages = computed(() =>
    Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1)))
  )

  async function loadClasses(nextPage = page.value): Promise<void> {
    loading.value = true
    error.value = null

    try {
      const data = await getClasses({
        page: nextPage,
        page_size: pageSize.value,
      })
      if (Array.isArray(data)) {
        list.value = data
        total.value = data.length
        return
      }

      list.value = data.list
      total.value = data.total
      page.value = data.page
      pageSize.value = data.page_size
    } catch (err) {
      reportFrontendError('加载班级列表失败:', err)
      error.value = '加载班级列表失败，请稍后重试'
      list.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  function handlePageChange(nextPage: number): void {
    const normalizedPage = Math.max(1, Math.floor(nextPage))
    if (normalizedPage === page.value || normalizedPage > totalPages.value) {
      return
    }

    void loadClasses(normalizedPage)
  }

  onMounted(() => {
    void loadClasses()
  })

  return {
    list,
    total,
    page,
    pageSize,
    totalPages,
    loading,
    error,
    loadClasses,
    handlePageChange,
  }
}
