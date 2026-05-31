import { ref, toValue, type MaybeRefOrGetter } from 'vue'

import { destroyManagedInstanceByRole } from '@/api/instances'
import {
  confirmDestructiveAction,
  type DestructiveConfirmOptions,
} from '@/shared/model/common/useDestructiveConfirm'
import type { UserRole } from '@/utils/constants'

interface ManagedInstanceDestroyTarget {
  id: string
}

interface UseManagedInstanceDestroyActionOptions<T extends ManagedInstanceDestroyTarget> {
  role: MaybeRefOrGetter<UserRole | null | undefined>
  resolveTarget: (id: string) => T | null | undefined
  buildConfirmOptions: (target: T) => DestructiveConfirmOptions
  onDestroyed: (context: { target: T }) => Promise<void> | void
  onDestroySuccess?: (context: { target: T }) => void
  onDestroyError?: (context: { target: T; error: unknown; message: string }) => void
  fallbackErrorMessage?: string
}

function resolveErrorMessage(error: unknown, fallback: string): string {
  return error instanceof Error && error.message.trim() ? error.message : fallback
}

export function useManagedInstanceDestroyAction<T extends ManagedInstanceDestroyTarget>(
  options: UseManagedInstanceDestroyActionOptions<T>
) {
  const destroyingId = ref('')
  let activeDestroyId = ''

  async function destroyManagedInstance(id: string): Promise<void> {
    const target = options.resolveTarget(id)
    if (!target || activeDestroyId) {
      return
    }

    activeDestroyId = target.id
    const confirmed = await confirmDestructiveAction(options.buildConfirmOptions(target))
    if (!confirmed) {
      activeDestroyId = ''
      return
    }

    destroyingId.value = target.id
    try {
      await destroyManagedInstanceByRole(toValue(options.role), target.id)
      await options.onDestroyed({ target })
      options.onDestroySuccess?.({ target })
    } catch (error) {
      const message = resolveErrorMessage(error, options.fallbackErrorMessage ?? '销毁实例失败')
      options.onDestroyError?.({ target, error, message })
    } finally {
      destroyingId.value = ''
      activeDestroyId = ''
    }
  }

  return {
    destroyingId,
    destroyManagedInstance,
  }
}
