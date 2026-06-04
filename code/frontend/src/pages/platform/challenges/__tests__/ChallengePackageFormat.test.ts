import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ChallengePackageFormatRoutePage from '@/pages/platform/challenges/ChallengePackageFormatRoutePage.vue'
import challengePackageFormatSource from '@/pages/platform/challenges/ChallengePackageFormatRoutePage.vue?raw'
import challengePackageFormatPageSource from '@/features/platform/challenge-package-import/model/useChallengePackageFormatPage.ts?raw'

describe('ChallengePackageFormat', () => {
  function mountChallengePackageFormat() {
    return mount(ChallengePackageFormatRoutePage, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :data-to="JSON.stringify(to)"><slot /></a>',
          },
        },
      },
    })
  }

  it('应该展示题目包结构与 challenge.yml 示例', () => {
    const wrapper = mountChallengePackageFormat()

    expect(wrapper.text()).toContain('题目包示例')
    expect(wrapper.text()).toContain('challenge.yml')
    expect(wrapper.text()).toContain('statement.md')
    expect(wrapper.text()).toContain('Dockerfile')
    expect(wrapper.text()).toContain('app.py')
    expect(wrapper.text()).toContain('不要写 # 题目名')
    expect(wrapper.text()).toContain('不要写 ## 题目描述')
    expect(wrapper.text()).toContain('api_version: v1')
    expect(wrapper.text()).toContain('flag:')
    expect(wrapper.text()).toContain('checker:')
    expect(wrapper.text()).toContain('http_standard')
    expect(wrapper.text()).toContain('tcp_standard')
    expect(wrapper.text()).toContain('script_checker')
    expect(wrapper.text()).toContain('SET_FLAG {{FLAG}}')
    expect(wrapper.text()).toContain('docker/check/protocol.py')
    expect(wrapper.text()).toContain('X-AWD-Checker-Token')
  })

  it('路由页应仅负责组合，不直接耦合返回跳转细节', () => {
    expect(challengePackageFormatSource).toContain('useChallengePackageFormatPage')
    expect(challengePackageFormatSource).toContain(
      "ChallengePackageFormatGuidePanel,\n  useChallengePackageFormatPage,\n} from '@/features/platform/challenge-package-import'"
    )
    expect(challengePackageFormatSource).toContain('<RouterLink')
    expect(challengePackageFormatSource).not.toContain('@click="backToImportManage"')
    expect(challengePackageFormatSource).not.toContain('useRouter')
    expect(challengePackageFormatPageSource).toContain("name: 'PlatformChallengeImportManage'")
    expect(challengePackageFormatPageSource).not.toContain("from 'vue-router'")
  })
})
