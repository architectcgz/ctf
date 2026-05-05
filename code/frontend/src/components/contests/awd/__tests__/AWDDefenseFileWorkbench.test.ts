import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AWDDefenseFileWorkbench from '../AWDDefenseFileWorkbench.vue'

describe('AWDDefenseFileWorkbench', () => {
  it('展示目录、文件内容并上抛只读打开事件', async () => {
    const wrapper = mount(AWDDefenseFileWorkbench, {
      props: {
        serviceTitle: 'Bank Portal',
        loading: false,
        error: '',
        directory: {
          path: '.',
          entries: [
            { name: 'templates', path: 'templates', type: 'dir', size: 0 },
            { name: 'app.py', path: 'app.py', type: 'file', size: 13 },
          ],
        },
        file: {
          path: 'app.py',
          content: "print('vuln')",
          size: 13,
        },
      },
    })

    expect(wrapper.text()).toContain('防守文件')
    expect(wrapper.text()).toContain('Bank Portal')
    expect(wrapper.text()).toContain('templates')
    expect(wrapper.text()).toContain('app.py')
    expect(wrapper.text()).toContain("print('vuln')")

    const refreshButton = wrapper.findAll('button').find((node) => node.text().includes('刷新'))
    const directoryButton = wrapper.findAll('button').find((node) => node.text().includes('templates'))
    const fileButton = wrapper.findAll('button').find((node) => node.text().includes('app.py'))

    expect(refreshButton).toBeTruthy()
    expect(directoryButton).toBeTruthy()
    expect(fileButton).toBeTruthy()

    await refreshButton!.trigger('click')
    await directoryButton!.trigger('click')
    await fileButton!.trigger('click')

    expect(wrapper.emitted('refresh')).toEqual([[]])
    expect(wrapper.emitted('openDirectory')).toEqual([['templates']])
    expect(wrapper.emitted('openFile')).toEqual([['app.py']])
  })

  it('进入子目录后应允许返回根目录和上一级', async () => {
    const wrapper = mount(AWDDefenseFileWorkbench, {
      props: {
        serviceTitle: 'Bank Portal',
        loading: false,
        error: '',
        directory: {
          path: 'templates/emails',
          entries: [],
        },
        file: null,
      },
    })

    const rootButton = wrapper.findAll('button').find((node) => node.text() === '根目录')
    const parentButton = wrapper.findAll('button').find((node) => node.text() === '上一级')

    expect(rootButton).toBeTruthy()
    expect(parentButton).toBeTruthy()

    await rootButton!.trigger('click')
    await parentButton!.trigger('click')

    expect(wrapper.emitted('openDirectory')).toEqual([['.'], ['templates']])
  })

  it('显示后端拒绝或未启用状态', () => {
    const wrapper = mount(AWDDefenseFileWorkbench, {
      props: {
        serviceTitle: 'Bank Portal',
        loading: false,
        error: '文件工作台暂不可用。',
        directory: null,
        file: null,
      },
    })

    expect(wrapper.text()).toContain('文件工作台暂不可用。')
  })
})
