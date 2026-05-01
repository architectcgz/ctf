import { computed, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type { AdminChallengeFlagPayload } from '@/api/admin/authoring'
import { configureChallengeFlag, getChallengeDetail } from '@/api/admin/authoring'
import { downloadAttachment as downloadChallengeAttachment } from '@/api/challenge'
import type { AdminChallengeListItem, FlagType } from '@/api/contracts'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useToast } from '@/composables/useToast'

function summarizeFlagConfig(config?: AdminChallengeListItem['flag_config']): string {
  if (!config?.configured) return '未配置'

  switch (config.flag_type) {
    case 'static':
      return '静态 Flag'
    case 'dynamic':
      return `动态 Flag / 前缀 ${config.flag_prefix || 'flag'}`
    case 'regex':
      return `正则匹配 / ${config.flag_regex || '未填写'}`
    case 'manual_review':
      return '人工审核'
    default:
      return '未配置'
  }
}

export function usePlatformChallengeDetailPage() {
  const route = useRoute()
  const router = useRouter()
  const toast = useToast()
  const { setBreadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()

  const loading = ref(true)
  const saving = ref(false)
  const downloadingAttachment = ref(false)
  const challenge = ref<AdminChallengeListItem | null>(null)
  const flagType = ref<FlagType>('static')
  const flagValue = ref('')
  const flagRegex = ref('')
  const flagPrefix = ref('')
  let redirectTimer: ReturnType<typeof setTimeout> | null = null

  const challengeId = computed(() => String(route.params.id || ''))
  const workspaceLabel = computed(() => challenge.value?.title || '题目管理')
  const flagConfigSummary = computed(() => summarizeFlagConfig(challenge.value?.flag_config))
  const isSharedInstanceChallenge = computed(() => challenge.value?.instance_sharing === 'shared')
  const flagDraftSummary = computed(() =>
    summarizeFlagConfig({
      configured: true,
      flag_type: flagType.value,
      flag_regex: flagRegex.value.trim() || undefined,
      flag_prefix: flagPrefix.value.trim() || undefined,
    })
  )

  function openTopology(): void {
    if (!challengeId.value) return
    void router.push(`/platform/challenges/${challengeId.value}/topology`)
  }

  function openChallengeList(): void {
    void router.push('/platform/challenges')
  }

  function clearRedirectTimer(): void {
    if (redirectTimer === null) {
      return
    }
    clearTimeout(redirectTimer)
    redirectTimer = null
  }

  async function downloadAttachment(): Promise<void> {
    const attachmentURL = challenge.value?.attachment_url?.trim()
    if (!attachmentURL) return

    try {
      const parsed = new URL(attachmentURL, window.location.origin)
      if (parsed.origin !== window.location.origin) {
        window.open(attachmentURL, '_blank', 'noopener')
        return
      }
    } catch {
      // fallback to axios download for relative urls
    }

    downloadingAttachment.value = true
    try {
      const { blob, filename } = await downloadChallengeAttachment(attachmentURL)
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = filename
      document.body.appendChild(link)
      link.click()
      link.remove()
      URL.revokeObjectURL(url)
    } catch {
      toast.error('下载附件失败')
    } finally {
      downloadingAttachment.value = false
    }
  }

  function hydrateFlagForm(item: AdminChallengeListItem | null): void {
    const config = item?.flag_config
    flagType.value = config?.flag_type ?? 'static'
    flagValue.value = ''
    flagRegex.value = config?.flag_regex ?? ''
    flagPrefix.value = config?.flag_prefix ?? ''
  }

  function setChallengeBreadcrumbTitle(title?: string): void {
    setBreadcrumbDetailTitle(title)
  }

  async function loadChallenge(id: string): Promise<void> {
    if (!id) {
      challenge.value = null
      setChallengeBreadcrumbTitle()
      loading.value = false
      return
    }

    try {
      setChallengeBreadcrumbTitle()
      challenge.value = await getChallengeDetail(id)
      setChallengeBreadcrumbTitle(challenge.value.title)
      hydrateFlagForm(challenge.value)
    } catch {
      challenge.value = null
      setChallengeBreadcrumbTitle()
      toast.error('加载失败')
      clearRedirectTimer()
      redirectTimer = setTimeout(() => {
        redirectTimer = null
        void router.push('/platform/challenges')
      }, 1500)
    } finally {
      loading.value = false
    }
  }

  async function saveFlagConfig(): Promise<void> {
    if (isSharedInstanceChallenge.value && flagType.value === 'dynamic') {
      toast.error(
        '共享实例只适用于无状态题，不支持动态 Flag；若需隔离答案，请使用 per_user 或 per_team'
      )
      return
    }

    const payload: AdminChallengeFlagPayload = {
      flag_type: flagType.value,
    }

    if (flagType.value === 'static') {
      if (!flagValue.value.trim()) {
        toast.error('请填写静态 Flag')
        return
      }
      payload.flag = flagValue.value.trim()
    }

    if (flagType.value === 'dynamic') {
      if (!flagPrefix.value.trim()) {
        toast.error('请填写动态 Flag 前缀')
        return
      }
      payload.flag_prefix = flagPrefix.value.trim()
    }

    if (flagType.value === 'regex') {
      if (!flagRegex.value.trim()) {
        toast.error('请填写正则表达式')
        return
      }
      payload.flag_regex = flagRegex.value.trim()
      if (flagPrefix.value.trim()) {
        payload.flag_prefix = flagPrefix.value.trim()
      }
    }

    saving.value = true
    try {
      await configureChallengeFlag(challengeId.value, payload)
      toast.success('Flag 配置已保存')
      loading.value = true
      await loadChallenge(challengeId.value)
    } catch {
      toast.error('保存 Flag 配置失败')
    } finally {
      saving.value = false
    }
  }

  watch(
    challengeId,
    (id) => {
      loading.value = true
      void loadChallenge(id)
    },
    { immediate: true }
  )

  onUnmounted(() => {
    clearRedirectTimer()
    setChallengeBreadcrumbTitle()
  })

  return {
    challenge,
    challengeId,
    downloadingAttachment,
    downloadAttachment,
    flagConfigSummary,
    flagDraftSummary,
    flagPrefix,
    flagRegex,
    flagType,
    flagValue,
    isSharedInstanceChallenge,
    loading,
    openChallengeList,
    openTopology,
    saveFlagConfig,
    saving,
    workspaceLabel,
  }
}
