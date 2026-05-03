import { computed, onMounted, ref, watch, type Ref } from 'vue'

import { deleteChallengeWriteup, getChallengeWriteup } from '@/api/admin/authoring'
import { getTeacherWriteupSubmissions } from '@/api/teacher'
import type { AdminChallengeWriteupData, TeacherSubmissionWriteupItemData } from '@/api/contracts'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'

export type WriteupDirectoryRow = {
  key: string
  source: 'official' | 'student'
  title: string
  preview?: string
  authorPrimary: string
  authorSecondary?: string
  authorTertiary?: string
  studentNo: string
  statusPrimary: string
  statusSecondary?: string
  updatedAt: string
}

interface UseChallengeWriteupManagementOptions {
  challengeId: Readonly<Ref<string>>
}

export function useChallengeWriteupManagement(options: UseChallengeWriteupManagementOptions) {
  const toast = useToast()
  const loading = ref(true)
  const deleting = ref(false)
  const submissionLoading = ref(true)
  const writeup = ref<AdminChallengeWriteupData | null>(null)
  const writeupSubmissions = ref<TeacherSubmissionWriteupItemData[]>([])
  const submissionPage = ref(1)
  const submissionPageSize = ref(6)
  const submissionTotal = ref(0)

  const submissionTotalPages = computed(() =>
    Math.max(1, Math.ceil(submissionTotal.value / Math.max(1, submissionPageSize.value)))
  )
  const officialWriteupCount = computed(() => (writeup.value ? 1 : 0))
  const hasAnyWriteups = computed(
    () => Boolean(writeup.value) || writeupSubmissions.value.length > 0
  )
  const directoryRows = computed<WriteupDirectoryRow[]>(() => {
    const rows: WriteupDirectoryRow[] = []

    if (writeup.value && submissionPage.value === 1) {
      rows.push({
        key: `official-${writeup.value.id}`,
        source: 'official',
        title: writeup.value.title,
        authorPrimary: '平台官方',
        authorSecondary: '独立查看 / 编辑入口',
        studentNo: '-',
        statusPrimary: writeup.value.visibility,
        statusSecondary: writeup.value.is_recommended ? '推荐题解' : '未推荐',
        updatedAt: formatDate(writeup.value.updated_at),
      })
    }

    rows.push(
      ...writeupSubmissions.value.map((item) => ({
        key: `student-${item.id}`,
        source: 'student' as const,
        title: item.title,
        preview: item.content_preview,
        authorPrimary: resolveAuthorName(item),
        authorSecondary: item.student_username,
        authorTertiary: resolveClassName(item),
        studentNo: resolveStudentNo(item),
        statusPrimary: submissionStatusLabel(item.submission_status),
        statusSecondary: visibilityStatusLabel(item.visibility_status),
        updatedAt: formatDate(item.updated_at),
      }))
    )

    return rows
  })

  async function loadWriteup() {
    if (!options.challengeId.value) {
      writeup.value = null
      loading.value = false
      return
    }

    loading.value = true
    try {
      writeup.value = await getChallengeWriteup(options.challengeId.value)
    } catch {
      toast.error('加载题解目录失败')
    } finally {
      loading.value = false
    }
  }

  async function loadWriteupSubmissions(targetPage = 1) {
    if (!options.challengeId.value) {
      writeupSubmissions.value = []
      submissionPage.value = 1
      submissionTotal.value = 0
      submissionLoading.value = false
      return
    }

    submissionLoading.value = true
    try {
      const payload = await getTeacherWriteupSubmissions({
        challenge_id: options.challengeId.value,
        page: targetPage,
        page_size: submissionPageSize.value,
      })
      writeupSubmissions.value = payload.list
      submissionPage.value = payload.page
      submissionPageSize.value = payload.page_size
      submissionTotal.value = payload.total
    } catch {
      toast.error('加载题解投稿失败')
    } finally {
      submissionLoading.value = false
    }
  }

  async function deleteOfficialWriteup(): Promise<boolean> {
    if (!options.challengeId.value || !writeup.value || deleting.value) {
      return false
    }

    const confirmed = await confirmDestructiveAction({
      message: '确定删除当前题解吗？删除后学员将无法继续查看。',
    })
    if (!confirmed) {
      return false
    }

    deleting.value = true
    try {
      await deleteChallengeWriteup(options.challengeId.value)
      writeup.value = null
      toast.success('题解已删除')
      return true
    } catch (error) {
      const message =
        error instanceof Error && error.message.trim() ? error.message : '删除题解失败'
      toast.error(message)
      return false
    } finally {
      deleting.value = false
    }
  }

  async function changeSubmissionPage(page: number) {
    if (
      page < 1 ||
      page === submissionPage.value ||
      submissionLoading.value ||
      !options.challengeId.value
    ) {
      return
    }

    await loadWriteupSubmissions(page)
  }

  watch(options.challengeId, () => {
    void loadWriteup()
    void loadWriteupSubmissions(1)
  })

  onMounted(() => {
    void loadWriteup()
    void loadWriteupSubmissions(1)
  })

  return {
    changeSubmissionPage,
    deleteOfficialWriteup,
    deleting,
    directoryRows,
    hasAnyWriteups,
    loading,
    loadWriteup,
    loadWriteupSubmissions,
    officialWriteupCount,
    submissionLoading,
    submissionPage,
    submissionPageSize,
    submissionTotal,
    submissionTotalPages,
    writeup,
  }
}

function formatDate(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString('zh-CN')
}

function submissionStatusLabel(status: TeacherSubmissionWriteupItemData['submission_status']): string {
  return status === 'draft' ? '草稿' : '已发布'
}

function visibilityStatusLabel(status: TeacherSubmissionWriteupItemData['visibility_status']): string {
  return status === 'hidden' ? '已隐藏' : '已公开'
}

function resolveAuthorName(item: TeacherSubmissionWriteupItemData): string {
  const name = item.student_name?.trim()
  return name || item.student_username
}

function resolveStudentNo(item: TeacherSubmissionWriteupItemData): string {
  const studentNo = item.student_no?.trim()
  return studentNo || '未设置学号'
}

function resolveClassName(item: TeacherSubmissionWriteupItemData): string {
  const className = item.class_name?.trim()
  return className || '未分班'
}
