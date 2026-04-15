import { ref, watch, type Ref } from 'vue'

import { getStudentReviewArchive } from '@/api/teacher'
import type { ReviewArchiveData } from '@/api/contracts'

export function useTeacherStudentReviewArchive(studentId: Readonly<Ref<string>>) {
  const archive = ref<ReviewArchiveData | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load(): Promise<void> {
    if (!studentId.value) {
      archive.value = null
      error.value = '缺少学生标识'
      return
    }

    loading.value = true
    error.value = null

    try {
      archive.value = await getStudentReviewArchive(studentId.value)
    } catch (err) {
      console.error('加载学生复盘归档失败:', err)
      archive.value = null
      error.value = '加载学生复盘归档失败，请稍后重试'
    } finally {
      loading.value = false
    }
  }

  watch(
    studentId,
    () => {
      void load()
    },
    { immediate: true }
  )

  return {
    archive,
    loading,
    error,
    reload: load,
  }
}
