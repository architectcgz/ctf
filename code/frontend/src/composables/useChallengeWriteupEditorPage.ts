import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessageBox } from 'element-plus'

import {
  deleteChallengeWriteup,
  getChallengeDetail,
  getChallengeWriteup,
  recommendChallengeWriteup,
  saveChallengeWriteup,
  unrecommendChallengeWriteup,
} from '@/api/admin'
import type {
  AdminChallengeListItem,
  AdminChallengeWriteupData,
  WriteupVisibility,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'

export function toLocalDateTimeInputValue(value?: string) {
  if (!value) {
    return ''
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000)
  return local.toISOString().slice(0, 16)
}

function fromLocalDateTimeInputValue(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? undefined : date.toISOString()
}

export function useChallengeWriteupEditorPage(challengeId: string) {
  const toast = useToast()

  const loading = ref(true)
  const saving = ref(false)
  const deleting = ref(false)
  const togglingRecommendation = ref(false)
  const challenge = ref<AdminChallengeListItem | null>(null)
  const writeup = ref<AdminChallengeWriteupData | null>(null)
  const form = reactive({
    title: '',
    content: '',
    visibility: 'private' as WriteupVisibility,
    releaseAt: '',
  })

  const hasWriteup = computed(() => writeup.value !== null)
  const visibilityLabel = computed(() => {
    switch (form.visibility) {
      case 'public':
        return '公开后，所有已发布挑战的学员都可查看'
      case 'scheduled':
        return '到达发布时间后自动公开'
      default:
        return '仅管理员可查看'
    }
  })

  function resetForm(item?: AdminChallengeWriteupData | null) {
    form.title = item?.title || ''
    form.content = item?.content || ''
    form.visibility = item?.visibility || 'private'
    form.releaseAt = toLocalDateTimeInputValue(item?.release_at)
  }

  async function loadPage() {
    loading.value = true
    try {
      const [challengeDetail, writeupDetail] = await Promise.all([
        getChallengeDetail(challengeId),
        getChallengeWriteup(challengeId),
      ])
      challenge.value = challengeDetail
      writeup.value = writeupDetail
      resetForm(writeupDetail)
    } catch {
      toast.error('加载题解管理页失败')
    } finally {
      loading.value = false
    }
  }

  function validateForm() {
    if (!form.title.trim()) {
      toast.error('题解标题不能为空')
      return false
    }
    if (!form.content.trim()) {
      toast.error('题解内容不能为空')
      return false
    }
    if (form.visibility === 'scheduled' && !form.releaseAt) {
      toast.error('定时公开必须设置发布时间')
      return false
    }
    return true
  }

  async function handleSave() {
    if (!validateForm()) {
      return
    }

    saving.value = true
    try {
      const saved = await saveChallengeWriteup(challengeId, {
        title: form.title.trim(),
        content: form.content.trim(),
        visibility: form.visibility,
        release_at:
          form.visibility === 'scheduled' ? fromLocalDateTimeInputValue(form.releaseAt) : undefined,
      })
      writeup.value = saved
      resetForm(saved)
      toast.success('题解已保存')
    } catch {
      toast.error('保存题解失败')
    } finally {
      saving.value = false
    }
  }

  async function handleDelete() {
    if (!writeup.value) {
      return
    }

    try {
      await ElMessageBox.confirm('确定删除当前题解吗？删除后学员将无法继续查看。', '确认删除', {
        type: 'warning',
      })
    } catch {
      return
    }

    deleting.value = true
    try {
      await deleteChallengeWriteup(challengeId)
      writeup.value = null
      resetForm(null)
      toast.success('题解已删除')
    } catch {
      toast.error('删除题解失败')
    } finally {
      deleting.value = false
    }
  }

  function restoreExistingWriteup() {
    resetForm(writeup.value)
  }

  async function handleToggleRecommendation() {
    if (!writeup.value) {
      return
    }

    togglingRecommendation.value = true
    try {
      writeup.value = writeup.value.is_recommended
        ? await unrecommendChallengeWriteup(challengeId)
        : await recommendChallengeWriteup(challengeId)
      toast.success(writeup.value.is_recommended ? '已设为推荐题解' : '已取消推荐题解')
    } catch {
      toast.error(writeup.value.is_recommended ? '取消推荐失败' : '设为推荐失败')
    } finally {
      togglingRecommendation.value = false
    }
  }

  onMounted(() => {
    void loadPage()
  })

  return {
    loading,
    saving,
    deleting,
    togglingRecommendation,
    challenge,
    writeup,
    form,
    hasWriteup,
    visibilityLabel,
    loadPage,
    handleSave,
    handleDelete,
    handleToggleRecommendation,
    restoreExistingWriteup,
  }
}
