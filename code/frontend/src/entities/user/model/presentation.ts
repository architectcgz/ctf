interface UserIdentityInput {
  name?: string | null
  username?: string | null
}

function normalizeValue(value: string | null | undefined): string {
  return typeof value === 'string' ? value.trim() : ''
}

export function getUserDisplayName(
  user: UserIdentityInput | null | undefined,
  fallback = '--'
): string {
  const name = normalizeValue(user?.name)
  if (name) {
    return name
  }

  const username = normalizeValue(user?.username)
  if (username) {
    return username
  }

  return fallback
}

export function getUserName(
  user: UserIdentityInput | null | undefined,
  fallback = '--'
): string {
  const name = normalizeValue(user?.name)
  return name || fallback
}

export function getUserUsername(
  user: UserIdentityInput | null | undefined,
  fallback = '--'
): string {
  const username = normalizeValue(user?.username)
  return username || fallback
}

export function getUserUsernameHandle(
  user: UserIdentityInput | null | undefined,
  fallback = '--'
): string {
  const username = normalizeValue(user?.username)
  return username ? `@${username}` : fallback
}

export function getUserDisplayLabel(
  user: UserIdentityInput | null | undefined,
  fallback = '--'
): string {
  const displayName = getUserDisplayName(user, fallback)
  const username = normalizeValue(user?.username)

  if (!username || username === displayName) {
    return displayName
  }

  return `${displayName} (${username})`
}
