import type {
  AWDCheckerBuildResult,
  AWDCheckerFieldErrorKey,
  AWDScriptCheckerDraft,
} from './awdCheckerConfigSupport'

export function buildScriptCheckerConfig(
  draft: AWDScriptCheckerDraft,
  strict = true
): AWDCheckerBuildResult {
  const errors: Partial<Record<AWDCheckerFieldErrorKey, string>> = {}
  const entry = draft.entry.trim()
  if (!entry && strict) {
    errors.script_entry = '入口文件不能为空'
  }
  if (entry.startsWith('/') || entry.includes('..')) {
    errors.script_entry = '入口文件必须是题目包内相对路径'
  }
  if (!Number.isInteger(draft.timeout_sec) || draft.timeout_sec <= 0 || draft.timeout_sec > 60) {
    errors.script_timeout = '超时时间必须是 1-60 秒'
  }

  const args = draft.args_text
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean)

  let env: Record<string, string> = {}
  const envText = draft.env_text.trim()
  if (envText) {
    try {
      const parsed = JSON.parse(envText)
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
        errors.script_env_text = '环境变量必须是 JSON 对象'
      } else {
        env = Object.fromEntries(
          Object.entries(parsed as Record<string, unknown>).map(([key, value]) => [
            key,
            String(value),
          ])
        )
      }
    } catch {
      errors.script_env_text = '环境变量必须是合法 JSON'
    }
  }

  return {
    config: {
      runtime: draft.runtime.trim() || 'python3',
      entry,
      timeout_sec:
        Number.isInteger(draft.timeout_sec) && draft.timeout_sec > 0 ? draft.timeout_sec : 10,
      args,
      ...(Object.keys(env).length > 0 ? { env } : {}),
      output: draft.output,
    },
    errors,
  }
}
