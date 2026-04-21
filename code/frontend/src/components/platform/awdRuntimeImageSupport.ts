const PLACEHOLDER_REGISTRY_PREFIXES = ['registry.example.edu/']

function normalizeImageRef(value: string): string {
  return value.trim()
}

export function extractAwdRuntimeImageRef(runtimeConfig?: Record<string, unknown>): string {
  const imageRef = runtimeConfig?.image_ref
  return typeof imageRef === 'string' ? normalizeImageRef(imageRef) : ''
}

export function isPlaceholderAwdRuntimeImageRef(imageRef: string): boolean {
  const normalized = normalizeImageRef(imageRef)
  if (!normalized) {
    return false
  }
  return PLACEHOLDER_REGISTRY_PREFIXES.some((prefix) => normalized.startsWith(prefix))
}

export function buildAwdRuntimeImagePlaceholderWarning(imageRef: string): string {
  if (!isPlaceholderAwdRuntimeImageRef(imageRef)) {
    return ''
  }
  return 'runtime.image.ref 使用了示例占位 registry；导入后若需试跑，请先在当前环境准备同名镜像，或改成可直接拉取的真实镜像地址。'
}

function looksLikeImagePullFailure(message: string): boolean {
  const normalized = message.toLowerCase()
  return [
    'failed to resolve reference',
    'no such image',
    'pull access denied',
    'manifest unknown',
    'docker 镜像不存在或无法访问',
    'repository does not exist',
  ].some((pattern) => normalized.includes(pattern))
}

function compactRuntimeError(message: string): string {
  const normalized = message.trim().replace(/\s+/g, ' ')
  if (normalized.length <= 220) {
    return normalized
  }
  return `${normalized.slice(0, 217)}...`
}

function extractImageRefFromRuntimeError(message: string): string {
  const matched = message.match(/reference\s+"([^"]+)"/i)
  if (matched?.[1]) {
    return matched[1].trim()
  }
  return ''
}

export function formatAwdPreviewRuntimeError(rawMessage: string, imageRef: string): string {
  const normalized = rawMessage.trim()
  if (!normalized || !looksLikeImagePullFailure(normalized)) {
    return normalized || '试跑失败，请稍后重试。'
  }

  const resolvedImageRef = imageRef || extractImageRefFromRuntimeError(normalized)
  const parts = ['自动拉起预览实例失败：当前模板引用的运行镜像暂时无法拉取。']

  if (isPlaceholderAwdRuntimeImageRef(resolvedImageRef)) {
    parts.push('如果这是示例占位地址，请先在当前环境构建同名镜像，或把模板镜像改成可直接拉取的真实地址。')
  } else {
    parts.push('请确认平台所在机器能拉取这个镜像，或先把镜像推送到当前环境可访问的仓库。')
  }

  if (resolvedImageRef) {
    parts.push(`镜像引用：${resolvedImageRef}`)
  }
  parts.push(`原始错误：${compactRuntimeError(normalized)}`)
  return parts.join(' ')
}
