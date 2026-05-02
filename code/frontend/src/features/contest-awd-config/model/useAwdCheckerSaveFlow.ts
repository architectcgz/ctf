import { ref, type Ref } from 'vue'

import { updateContestAWDService } from '@/api/admin/contests'
import type { AdminContestAWDServiceData, AWDCheckerType } from '@/api/contracts'
import { useToast } from '@/composables/useToast'

interface UseAwdCheckerSaveFlowOptions {
  contestId: Readonly<Ref<string>>
  selectedService: Readonly<Ref<AdminContestAWDServiceData | null>>
  selectedCheckerType: Readonly<Ref<AWDCheckerType | undefined>>
  canAttachPreviewToken: Readonly<Ref<boolean>>
  previewToken: Readonly<Ref<string>>
  form: {
    sla_score: number
    defense_score: number
  }
  validateConfig: () => boolean
  buildCurrentCheckerConfig: () => Record<string, unknown>
  reloadPage: () => Promise<void>
}

export function useAwdCheckerSaveFlow(options: UseAwdCheckerSaveFlowOptions) {
  const {
    contestId,
    selectedService,
    selectedCheckerType,
    canAttachPreviewToken,
    previewToken,
    form,
    validateConfig,
    buildCurrentCheckerConfig,
    reloadPage,
  } = options

  const toast = useToast()
  const saving = ref(false)

  async function handleSave() {
    if (
      saving.value ||
      !selectedService.value ||
      !selectedCheckerType.value ||
      !validateConfig()
    ) {
      return
    }
    saving.value = true
    try {
      await updateContestAWDService(contestId.value, selectedService.value.id, {
        checker_type: selectedCheckerType.value,
        checker_config: buildCurrentCheckerConfig(),
        awd_sla_score: form.sla_score,
        awd_defense_score: form.defense_score,
        ...(canAttachPreviewToken.value
          ? { awd_checker_preview_token: previewToken.value }
          : {}),
      })
      toast.success(canAttachPreviewToken.value ? '配置与试跑结果已保存' : '配置已保存')
      await reloadPage()
    } catch (error) {
      toast.error(error instanceof Error && error.message.trim() ? error.message : '保存 AWD 配置失败')
    } finally {
      saving.value = false
    }
  }

  return {
    handleSave,
    saving,
  }
}
