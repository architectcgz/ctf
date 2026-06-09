import type {
  AWDCheckerBuildResult,
  AWDCheckerFieldErrorKey,
  AWDTCPStandardDraft,
} from './awdCheckerConfigSupport'

export function buildTCPStandardCheckerConfig(
  draft: AWDTCPStandardDraft,
  strict = true
): AWDCheckerBuildResult {
  const errors: Partial<Record<AWDCheckerFieldErrorKey, string>> = {}
  if (!Number.isInteger(draft.timeout_ms) || draft.timeout_ms <= 0 || draft.timeout_ms > 60000) {
    errors.tcp_timeout = '超时时间必须是 1-60000 毫秒'
  }

  const steps = draft.steps
    .map((step) => ({
      send: step.send,
      send_template: step.send_template,
      send_hex: step.send_hex.trim(),
      expect_contains: step.expect_contains,
      expect_regex: step.expect_regex,
      timeout_ms: step.timeout_ms,
    }))
    .filter(
      (step) =>
        step.send ||
        step.send_template ||
        step.send_hex ||
        step.expect_contains ||
        step.expect_regex
    )

  if (steps.length === 0 && strict) {
    errors.tcp_steps = '至少需要一个 TCP 步骤'
  }

  for (const step of steps) {
    const sendFieldCount = [step.send, step.send_template, step.send_hex].filter(Boolean).length
    const hasExpectation = Boolean(step.expect_contains || step.expect_regex)
    if (sendFieldCount === 0 && !hasExpectation && strict) {
      errors.tcp_steps = '每个步骤至少需要发送内容或期望结果'
    }
    if (step.send_hex && (step.send || step.send_template) && strict) {
      errors.tcp_steps = 'send_hex 不能与 send 或 send_template 同时填写'
    }
    if (step.send && step.send_template && strict) {
      errors.tcp_steps = 'send 不能与 send_template 同时填写'
    }
    if (
      !Number.isInteger(step.timeout_ms) ||
      step.timeout_ms < 0 ||
      step.timeout_ms > 60000
    ) {
      errors.tcp_steps = '步骤超时时间必须是 0-60000 毫秒'
    }
    if (step.expect_regex) {
      try {
        new RegExp(step.expect_regex)
      } catch {
        errors.tcp_steps = 'expect_regex 必须是合法正则'
      }
    }
  }

  return {
    config: {
      timeout_ms:
        Number.isInteger(draft.timeout_ms) && draft.timeout_ms > 0 ? draft.timeout_ms : 3000,
      steps: steps.map((step) => ({
        ...(step.send ? { send: step.send } : {}),
        ...(step.send_template ? { send_template: step.send_template } : {}),
        ...(step.send_hex ? { send_hex: step.send_hex } : {}),
        ...(step.expect_contains ? { expect_contains: step.expect_contains } : {}),
        ...(step.expect_regex ? { expect_regex: step.expect_regex } : {}),
        ...(Number.isInteger(step.timeout_ms) && step.timeout_ms > 0
          ? { timeout_ms: step.timeout_ms }
          : {}),
      })),
    },
    errors,
  }
}
