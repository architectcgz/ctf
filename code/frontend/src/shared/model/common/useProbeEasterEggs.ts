interface ProbeStorageState {
  version: 1
  counts: Record<string, number>
  activated: Record<string, boolean>
}

interface ProbeTrackResult {
  count: number
  unlocked: boolean
  activated: boolean
}

const STORAGE_KEY = 'ctf.probe-easter-eggs'

function createEmptyState(): ProbeStorageState {
  return {
    version: 1,
    counts: {},
    activated: {},
  }
}

function getSessionStorage(): Storage | null {
  try {
    return window.sessionStorage
  } catch {
    return null
  }
}

function readStoredState(): ProbeStorageState {
  const storage = getSessionStorage()
  if (!storage) return createEmptyState()

  try {
    const raw = storage.getItem(STORAGE_KEY)
    if (!raw) return createEmptyState()

    const parsed = JSON.parse(raw) as Partial<ProbeStorageState>
    return {
      version: 1,
      counts: parsed.counts && typeof parsed.counts === 'object' ? parsed.counts : {},
      activated:
        parsed.activated && typeof parsed.activated === 'object' ? parsed.activated : {},
    }
  } catch {
    return createEmptyState()
  }
}

function persistState(state: ProbeStorageState): void {
  const storage = getSessionStorage()
  if (!storage) return

  try {
    storage.setItem(STORAGE_KEY, JSON.stringify(state))
  } catch {
    // sessionStorage 不可用时退回当前实例内存态即可
  }
}

export function useProbeEasterEggs() {
  const state = readStoredState()

  function track(key: string, threshold: number): ProbeTrackResult {
    const nextCount = (state.counts[key] ?? 0) + 1
    state.counts[key] = nextCount

    const alreadyActivated = state.activated[key] === true
    const shouldUnlock = !alreadyActivated && nextCount >= threshold

    if (shouldUnlock) {
      state.activated[key] = true
    }

    persistState(state)

    return {
      count: nextCount,
      unlocked: shouldUnlock,
      activated: state.activated[key] === true,
    }
  }

  function isActivated(key: string): boolean {
    return state.activated[key] === true
  }

  return {
    track,
    isActivated,
  }
}
