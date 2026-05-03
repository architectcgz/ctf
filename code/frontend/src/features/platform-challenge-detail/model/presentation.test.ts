import { describe, expect, it } from 'vitest'

import {
  applyChallengeFlagDraftPatch,
  buildChallengeFlagDraftSummary,
  summarizeChallengeFlagConfig,
} from './presentation'

describe('platform challenge detail flag presentation', () => {
  it('应根据配置状态与类型返回判题配置摘要', () => {
    expect(summarizeChallengeFlagConfig()).toBe('未配置')
    expect(
      summarizeChallengeFlagConfig({
        configured: true,
        flag_type: 'static',
      })
    ).toBe('静态 Flag')
    expect(
      summarizeChallengeFlagConfig({
        configured: true,
        flag_type: 'dynamic',
        flag_prefix: 'ctf',
      })
    ).toBe('动态 Flag / 前缀 ctf')
    expect(
      summarizeChallengeFlagConfig({
        configured: true,
        flag_type: 'regex',
        flag_regex: '^flag',
      })
    ).toBe('正则匹配 / ^flag')
  })

  it('应根据草稿输入生成判题配置摘要并处理空白值', () => {
    expect(
      buildChallengeFlagDraftSummary({
        flagType: 'dynamic',
        flagPrefix: ' ',
        flagRegex: '',
      })
    ).toBe('动态 Flag / 前缀 flag')

    expect(
      buildChallengeFlagDraftSummary({
        flagType: 'regex',
        flagPrefix: 'ctf',
        flagRegex: '  ',
      })
    ).toBe('正则匹配 / 未填写')
  })

  it('应仅用 patch 覆盖指定字段并保留其余草稿值', () => {
    expect(
      applyChallengeFlagDraftPatch(
        {
          flagType: 'static',
          flagValue: 'flag{demo}',
          flagRegex: '^flag',
          flagPrefix: 'ctf',
        },
        {
          flagType: 'regex',
          flagRegex: '^ctf',
        }
      )
    ).toEqual({
      flagType: 'regex',
      flagValue: 'flag{demo}',
      flagRegex: '^ctf',
      flagPrefix: 'ctf',
    })
  })
})
