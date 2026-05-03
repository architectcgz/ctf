import { nextTick, watch, type Ref } from 'vue'

interface UseInstanceWarningFocusOptions {
  showWarning: Ref<boolean>
  warningCloseButton: Ref<HTMLButtonElement | null>
}

export function useInstanceWarningFocus(options: UseInstanceWarningFocusOptions) {
  let previouslyFocusedElement: HTMLElement | null = null

  watch(options.showWarning, async (visible) => {
    if (visible) {
      previouslyFocusedElement =
        document.activeElement instanceof HTMLElement ? document.activeElement : null
      await nextTick()
      options.warningCloseButton.value?.focus()
      return
    }

    previouslyFocusedElement?.focus()
    previouslyFocusedElement = null
  })
}
