import { config } from '@vue/test-utils'

config.global.stubs = {
  ...(config.global.stubs || {}),
  ElButton: {
    template: '<button type="button"><slot /></button>',
  },
  ElTag: {
    template: '<span><slot /></span>',
  },
  ElCard: {
    template: '<div><slot /><slot name="header" /></div>',
  },
}
