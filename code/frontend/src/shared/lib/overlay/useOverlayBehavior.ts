import { onBeforeUnmount, unref, watch } from 'vue'
import type { MaybeRef, Ref } from 'vue'

interface OverlayStackEntry {
  id: symbol
  close: () => void
  closeOnEscape: () => boolean
}

interface UseOverlayBehaviorOptions {
  open: Readonly<Ref<boolean>>
  close: () => void
  closeOnEscape?: MaybeRef<boolean>
  lockScroll?: MaybeRef<boolean>
}

const overlayStack: OverlayStackEntry[] = []
let scrollLockCount = 0
let originalBodyOverflow: string | null = null
let isGlobalKeydownListening = false

function handleGlobalKeydown(event: KeyboardEvent): void {
  if (event.key !== 'Escape') return

  const topOverlay = overlayStack.at(-1)
  if (!topOverlay?.closeOnEscape()) return

  event.stopPropagation()
  topOverlay.close()
}

function syncGlobalKeydownListener(): void {
  if (typeof window === 'undefined') return

  if (overlayStack.length > 0 && !isGlobalKeydownListening) {
    window.addEventListener('keydown', handleGlobalKeydown)
    isGlobalKeydownListening = true
    return
  }

  if (overlayStack.length === 0 && isGlobalKeydownListening) {
    window.removeEventListener('keydown', handleGlobalKeydown)
    isGlobalKeydownListening = false
  }
}

function registerOverlay(entry: OverlayStackEntry): void {
  if (overlayStack.some((item) => item.id === entry.id)) return
  overlayStack.push(entry)
  syncGlobalKeydownListener()
}

function unregisterOverlay(id: symbol): void {
  const index = overlayStack.findIndex((entry) => entry.id === id)
  if (index >= 0) {
    overlayStack.splice(index, 1)
  }
  syncGlobalKeydownListener()
}

function acquireScrollLock(): void {
  if (typeof document === 'undefined') return

  if (scrollLockCount === 0) {
    originalBodyOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
  }

  scrollLockCount += 1
}

function releaseScrollLock(): void {
  if (typeof document === 'undefined' || scrollLockCount === 0) return

  scrollLockCount -= 1
  if (scrollLockCount > 0) return

  document.body.style.overflow = originalBodyOverflow ?? ''
  originalBodyOverflow = null
}

export function useOverlayBehavior({
  open,
  close,
  closeOnEscape = true,
  lockScroll = true,
}: UseOverlayBehaviorOptions) {
  const overlayId = Symbol('overlay')
  const overlayEntry: OverlayStackEntry = {
    id: overlayId,
    close,
    closeOnEscape: () => unref(closeOnEscape),
  }
  let isOverlayRegistered = false
  let hasScrollLock = false

  watch(
    () => [open.value, unref(closeOnEscape), unref(lockScroll)] as const,
    ([isOpen, , shouldLockScroll]) => {
      if (isOpen && !isOverlayRegistered) {
        registerOverlay(overlayEntry)
        isOverlayRegistered = true
      }

      if (!isOpen && isOverlayRegistered) {
        unregisterOverlay(overlayId)
        isOverlayRegistered = false
      }

      if (isOpen && shouldLockScroll && !hasScrollLock) {
        acquireScrollLock()
        hasScrollLock = true
        return
      }

      if ((!isOpen || !shouldLockScroll) && hasScrollLock) {
        releaseScrollLock()
        hasScrollLock = false
      }
    },
    { immediate: true }
  )

  onBeforeUnmount(() => {
    if (isOverlayRegistered) {
      unregisterOverlay(overlayId)
      isOverlayRegistered = false
    }
    if (hasScrollLock) {
      releaseScrollLock()
      hasScrollLock = false
    }
  })
}
