import { onMounted, reactive, ref } from 'vue'
import { ElMessageBox } from 'element-plus'

import {
  configureChallengeFlag,
  createChallenge,
  deleteChallenge,
  getChallengeDetail,
  getChallengeFlagConfig,
  getChallenges,
  getImages,
  publishChallenge,
  updateChallenge,
  type AdminChallengePayload,
} from '@/api/admin'
import type {
  AdminChallengeHint,
  AdminChallengeListItem,
  AdminImageListItem,
  ChallengeCategory,
  ChallengeDifficulty,
  ChallengeStatus,
} from '@/api/contracts'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'

type EditableDifficulty = Extract<
  ChallengeDifficulty,
  'beginner' | 'easy' | 'medium' | 'hard' | 'insane'
>

export interface ChallengeFormDraft {
  title: string
  category: ChallengeCategory
  difficulty: EditableDifficulty
  points: number
  description: string
  attachment_url: string
  image_id: string
  hints: AdminChallengeHint[]
  flag_type: 'static' | 'dynamic'
  flag: string
  flag_prefix: string
  current_status: ChallengeStatus
  publish_after_save: boolean
}

function createEmptyForm(): ChallengeFormDraft {
  return {
    title: '',
    category: 'web',
    difficulty: 'easy',
    points: 100,
    description: '',
    attachment_url: '',
    image_id: '',
    hints: [],
    flag_type: 'static',
    flag: '',
    flag_prefix: '',
    current_status: 'draft',
    publish_after_save: false,
  }
}

function normalizeAdminHints(hints?: AdminChallengeHint[]): AdminChallengeHint[] {
  return (hints ?? []).map((hint) => ({
    id: hint.id,
    level: hint.level,
    title: hint.title || '',
    cost_points: hint.cost_points,
    content: hint.content,
  }))
}

function normalizeDifficulty(difficulty: ChallengeDifficulty): EditableDifficulty {
  return difficulty
}

function applyFormDraft(target: ChallengeFormDraft, next: ChallengeFormDraft) {
  Object.assign(target, next)
}

export function useAdminChallenges() {
  const toast = useToast()
  const dialogVisible = ref(false)
  const saving = ref(false)
  const editingId = ref<string | null>(null)
  const images = ref<AdminImageListItem[]>([])
  const form = reactive<ChallengeFormDraft>(createEmptyForm())

  const pagination = usePagination(getChallenges)

  async function openDialog(row?: AdminChallengeListItem) {
    if (row) {
      editingId.value = row.id
      let detail = row
      try {
        detail = await getChallengeDetail(row.id)
      } catch {
        toast.error('加载挑战详情失败，已回退到基础信息')
      }

      applyFormDraft(form, {
        title: detail.title,
        category: detail.category,
        difficulty: normalizeDifficulty(detail.difficulty),
        points: detail.points,
        description: detail.description || '',
        attachment_url: detail.attachment_url || '',
        image_id: detail.image_id || '',
        hints: normalizeAdminHints(detail.hints),
        flag_type: detail.flag_config?.flag_type || 'static',
        flag: '',
        flag_prefix: detail.flag_config?.flag_prefix || '',
        current_status: detail.status,
        publish_after_save: false,
      })

      try {
        const flagConfig = await getChallengeFlagConfig(row.id)
        form.flag_type = flagConfig.flag_type
        form.flag_prefix = flagConfig.flag_prefix || ''
      } catch {
        form.flag_type = 'static'
        form.flag_prefix = ''
      }
    } else {
      editingId.value = null
      applyFormDraft(form, createEmptyForm())
    }

    dialogVisible.value = true
  }

  function addHint() {
    form.hints.push({
      level: form.hints.length + 1,
      title: '',
      cost_points: 0,
      content: '',
    })
  }

  function removeHint(index: number) {
    form.hints.splice(index, 1)
  }

  function validateHints() {
    const levels = new Set<number>()

    for (const hint of form.hints) {
      if (!hint.level || hint.level < 1) {
        toast.error('提示级别必须大于 0')
        return false
      }
      if (levels.has(hint.level)) {
        toast.error('提示级别不能重复')
        return false
      }
      levels.add(hint.level)
      if (!hint.content.trim()) {
        toast.error('提示内容不能为空')
        return false
      }
    }

    return true
  }

  async function saveChallenge() {
    if (!form.title.trim()) {
      toast.error('请填写标题')
      return
    }
    if (!validateHints()) {
      return
    }
    if (!editingId.value && form.flag_type === 'static' && !form.flag.trim()) {
      toast.error('静态 Flag 不能为空')
      return
    }
    const imageID = form.image_id ? Number(form.image_id) : 0
    if (Number.isNaN(imageID) || imageID < 0) {
      toast.error('镜像配置无效')
      return
    }

    saving.value = true
    try {
      const challengePayload: AdminChallengePayload = {
        title: form.title.trim(),
        description: form.description.trim() || undefined,
        category: form.category,
        difficulty: form.difficulty,
        points: form.points,
        image_id: imageID,
        attachment_url: form.attachment_url.trim() || undefined,
        hints: normalizeAdminHints(form.hints),
      }

      const challengeId = editingId.value
        ? editingId.value
        : (await createChallenge(challengePayload)).challenge.id

      if (editingId.value) {
        await updateChallenge(challengeId, challengePayload)
      }

      const shouldUpdateFlag =
        form.flag_type === 'dynamic' ||
        !editingId.value ||
        Boolean(form.flag.trim()) ||
        Boolean(form.flag_prefix.trim())

      if (shouldUpdateFlag) {
        await configureChallengeFlag(challengeId, {
          flag_type: form.flag_type,
          flag: form.flag_type === 'static' ? form.flag.trim() : undefined,
          flag_prefix: form.flag_prefix.trim() || undefined,
        })
      }

      if (form.publish_after_save && form.current_status !== 'published') {
        await publishChallenge(challengeId)
      }

      toast.success(form.publish_after_save ? '保存并发布成功' : '保存成功')
      dialogVisible.value = false
      await pagination.refresh()
    } catch {
      toast.error('保存失败')
    } finally {
      saving.value = false
    }
  }

  async function publish(row: AdminChallengeListItem) {
    try {
      await publishChallenge(row.id)
      toast.success('发布成功')
      await pagination.refresh()
    } catch {
      toast.error('发布失败，请先确认镜像和 Flag 已配置')
    }
  }

  async function remove(id: string) {
    try {
      await ElMessageBox.confirm('确定要删除此挑战吗？', '确认', { type: 'warning' })
      await deleteChallenge(id)
      toast.success('删除成功')
      await pagination.refresh()
    } catch (error) {
      if (error !== 'cancel') {
        toast.error('删除失败')
      }
    }
  }

  async function loadImages() {
    try {
      const res = await getImages({ page: 1, page_size: 100 })
      images.value = res.list.filter((img) => img.status === 'available')
    } catch {
      toast.error('加载镜像列表失败')
    }
  }

  onMounted(() => {
    void pagination.refresh()
    void loadImages()
  })

  return {
    ...pagination,
    dialogVisible,
    saving,
    editingId,
    images,
    form,
    openDialog,
    addHint,
    removeHint,
    saveChallenge,
    publish,
    remove,
  }
}
