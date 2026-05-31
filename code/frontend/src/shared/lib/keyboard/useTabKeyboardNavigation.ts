interface UseTabKeyboardNavigationOptions<T extends string> {
  orderedTabs: readonly T[]
  selectTab: (tab: T) => void | Promise<void>
}

interface UseTabKeyboardNavigationResult<T extends string> {
  setTabButtonRef: (tab: T, element: HTMLButtonElement | null) => void
  handleTabKeydown: (event: KeyboardEvent, index: number) => void
}

export function useTabKeyboardNavigation<T extends string>({
  orderedTabs,
  selectTab,
}: UseTabKeyboardNavigationOptions<T>): UseTabKeyboardNavigationResult<T> {
  const tabButtonRefs: Partial<Record<T, HTMLButtonElement | null>> = {}

  function setTabButtonRef(tab: T, element: HTMLButtonElement | null): void {
    tabButtonRefs[tab] = element
  }

  function focusTab(tab: T): void {
    tabButtonRefs[tab]?.focus()
  }

  function handleTabKeydown(event: KeyboardEvent, index: number): void {
    if (
      event.key !== 'ArrowRight' &&
      event.key !== 'ArrowLeft' &&
      event.key !== 'Home' &&
      event.key !== 'End'
    ) {
      return
    }

    event.preventDefault()

    if (event.key === 'Home') {
      const firstTab = orderedTabs[0]
      if (!firstTab) return
      void selectTab(firstTab)
      focusTab(firstTab)
      return
    }

    if (event.key === 'End') {
      const lastTab = orderedTabs[orderedTabs.length - 1]
      if (!lastTab) return
      void selectTab(lastTab)
      focusTab(lastTab)
      return
    }

    const direction = event.key === 'ArrowRight' ? 1 : -1
    const nextIndex = (index + direction + orderedTabs.length) % orderedTabs.length
    const nextTab = orderedTabs[nextIndex]
    if (!nextTab) return
    void selectTab(nextTab)
    focusTab(nextTab)
  }

  return {
    setTabButtonRef,
    handleTabKeydown,
  }
}
