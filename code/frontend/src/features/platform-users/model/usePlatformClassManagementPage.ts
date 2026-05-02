import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import type { TeacherClassItem } from '@/api/contracts'
import { getClasses } from '@/api/teacher'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'

export function usePlatformClassManagementPage() {
  const router = useRouter()
  const list = ref<TeacherClassItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(DEFAULT_PAGE_SIZE)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1))))

  async function loadClasses(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const data = await getClasses({
        page: page.value,
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
      console.error('加载班级列表失败:', err)
      error.value = '加载班级列表失败，请稍后重试'
      list.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  function handlePageChange(p: number): void {
    const normalizedPage = Math.max(1, Math.floor(p))
    if (normalizedPage === page.value || normalizedPage > totalPages.value) {
      return
    }

    page.value = normalizedPage
    void loadClasses()
  }

  function openClass(className: string): void {
    void router.push({
      name: 'PlatformClassStudents',
      params: { className },
    })
  }

  const totalStudents = computed(() =>
    list.value.reduce((sum, item) => sum + (item.student_count || 0), 0)
  )

  const rows = computed(() =>
    list.value.map((item, index) => ({
      id: item.name,
      name: item.name,
      student_count: item.student_count || 0,
      teacher_name: '--',
      created_at: '--',
      actions: '查看班级',
      rowIndex: index,
    }))
  )

  onMounted(() => {
    void loadClasses()
  })

  return {
    list,
    total,
    page,
    totalPages,
    loading,
    error,
    totalStudents,
    rows,
    loadClasses,
    handlePageChange,
    openClass,
  }
}
