export interface AwdHttpStandardPreset {
  id: string
  label: string
}

export interface AwdHttpActionDraft {
  method: string
  path: string
  expected_status: number | null
  body_template: string
  expected_substring: string
  headers_text: string
}

export interface AwdHttpActionSection {
  key: string
  title: string
  pathErrorKey?: string
  statusErrorKey: string
  headersErrorKey: string
}

export interface AwdLegacyProbeDraft {
  health_path: string
}

export interface AwdCheckerFormDraft {
  sla_score: number
  defense_score: number
}

export interface AwdPreviewFormDraft {
  access_url: string
  preview_flag: string
}

export interface AwdTcpCheckerStepDraft {
  send: string
  send_template: string
  send_hex: string
  expect_contains: string
  expect_regex: string
  timeout_ms: number
}

export interface AwdTcpStandardDraft {
  timeout_ms: number | null
  steps: AwdTcpCheckerStepDraft[]
}

export interface AwdScriptCheckerDraft {
  runtime: string
  output: string
  timeout_sec: number | null
  entry: string
  args_text: string
  env_text: string
}

export type AwdConfigFieldErrors = Record<string, string | undefined>
