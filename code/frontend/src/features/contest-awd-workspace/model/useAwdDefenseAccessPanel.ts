import { computed, ref, toValue, type MaybeRefOrGetter } from 'vue'

import type {
  AWDDefenseSSHAccessData,
  ContestAWDWorkspaceServiceData,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'
import { getVSCodeSSHCommand } from './sshAccessPresentation'

interface UseAwdDefenseAccessPanelOptions {
  selectedServiceId: MaybeRefOrGetter<string>
  servicesByServiceId: MaybeRefOrGetter<Map<string, ContestAWDWorkspaceServiceData>>
  sshAccessByServiceId: MaybeRefOrGetter<Record<string, AWDDefenseSSHAccessData>>
  openService: (instanceId: string) => Promise<string | null>
}

export function useAwdDefenseAccessPanel(options: UseAwdDefenseAccessPanelOptions) {
  const toast = useToast()
  const copiedSSHCommandKey = ref('')
  const copiedSSHPasswordKey = ref('')

  const selectedDefenseAccess = computed(() => getSSHAccess(toValue(options.selectedServiceId)))
  const selectedDefenseCopiedCommand = computed(
    () => copiedSSHCommandKey.value === toValue(options.selectedServiceId)
  )
  const selectedDefenseCopiedPassword = computed(
    () => copiedSSHPasswordKey.value === toValue(options.selectedServiceId)
  )

  function getSSHAccess(serviceId?: string) {
    if (!serviceId) return undefined
    return toValue(options.sshAccessByServiceId)[serviceId]
  }

  function getSSHCommand(serviceId?: string): string {
    return getVSCodeSSHCommand(getSSHAccess(serviceId))
  }

  function openDefenseService(serviceId: string): void {
    const instanceId = toValue(options.servicesByServiceId).get(serviceId)?.instance_id
    if (!instanceId) return
    void options.openService(instanceId)
  }

  async function copyTextToClipboard(text: string, successMessage: string): Promise<boolean> {
    if (!text || typeof navigator === 'undefined' || !navigator.clipboard) {
      toast.error('复制失败，请手动选择文本')
      return false
    }

    try {
      await navigator.clipboard.writeText(text)
      toast.success(successMessage)
      return true
    } catch (err) {
      console.error(err)
      toast.error('复制失败，请手动选择文本')
      return false
    }
  }

  async function copySSHCommand(serviceId?: string): Promise<void> {
    if (!serviceId) return
    const copied = await copyTextToClipboard(getSSHCommand(serviceId), 'SSH 命令已复制')
    if (copied) {
      copiedSSHCommandKey.value = serviceId
    }
  }

  async function copySSHPassword(serviceId?: string): Promise<void> {
    if (!serviceId) return
    const password = getSSHAccess(serviceId)?.password || ''
    const copied = await copyTextToClipboard(password, 'SSH 密码已复制')
    if (copied) {
      copiedSSHPasswordKey.value = serviceId
    }
  }

  return {
    selectedDefenseAccess,
    selectedDefenseCopiedCommand,
    selectedDefenseCopiedPassword,
    getSSHCommand,
    openDefenseService,
    copySSHCommand,
    copySSHPassword,
  }
}
