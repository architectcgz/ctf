import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AWDDefenseFileWorkbench from '../AWDDefenseFileWorkbench.vue'
import awdDefenseFileWorkbenchSource from '../AWDDefenseFileWorkbench.vue?raw'

describe('AWDDefenseFileWorkbench', () => {
  it('展示左侧目录入口，并在查看区直接提供可编辑文件输入框', async () => {
    const wrapper = mount(AWDDefenseFileWorkbench, {
      props: {
        serviceTitle: 'Bank Portal',
        loading: false,
        error: '',
        editablePaths: ['app/app.py', 'app/utils/crypto.py'],
        directory: {
          path: 'app',
          entries: [
            { name: 'templates', path: 'app/templates', type: 'dir', size: 0 },
            { name: 'app.py', path: 'app/app.py', type: 'file', size: 13 },
          ],
        },
        file: {
          path: 'app/app.py',
          content: "print('vuln')",
          size: 13,
        },
      },
    })

    expect(wrapper.get('section').attributes('aria-label')).toBe('防守文件')
    expect(wrapper.text()).toContain('目录入口')
    expect(wrapper.text()).toContain('Bank Portal')
    expect(wrapper.text()).toContain('app')
    expect(wrapper.text()).toContain('templates')
    expect(wrapper.text()).toContain('app.py')
    expect(wrapper.text()).toContain('已保存')
    expect(wrapper.get('section').classes()).toEqual(
      expect.arrayContaining(['metric-panel-default-surface', 'metric-panel-workspace-surface'])
    )
    expect(wrapper.find('.defense-file-workbench__file--content').exists()).toBe(false)
    expect(wrapper.findAll('.defense-file-workbench__entries')).toHaveLength(1)
    expect(wrapper.find('.defense-file-workbench__editor-shell').exists()).toBe(true)
    expect(wrapper.find('.defense-file-workbench__editor-gutter').exists()).toBe(true)
    expect(wrapper.text()).toContain('Python')
    expect(wrapper.text()).toContain('app/app.py')
    expect(wrapper.text()).toContain('已保存')
    expect(wrapper.find('.defense-file-workbench__file--draft').exists()).toBe(false)

    const directoryButton = wrapper.findAll('button').find((node) => node.text().includes('templates'))
    const fileButton = wrapper.findAll('button').find((node) => node.text().includes('app.py'))
    const textarea = wrapper.find('textarea')

    expect(directoryButton).toBeTruthy()
    expect(fileButton).toBeTruthy()
    expect(directoryButton!.classes()).toEqual(expect.arrayContaining(['ui-btn', 'ui-btn--ghost']))
    expect(fileButton!.classes()).toEqual(expect.arrayContaining(['ui-btn', 'ui-btn--ghost']))
    expect(textarea.exists()).toBe(true)
    expect(textarea.classes()).toContain('defense-file-workbench__editor')
    expect((textarea.element as HTMLTextAreaElement).value).toContain("print('vuln')")

    await directoryButton!.trigger('click')
    await fileButton!.trigger('click')

    expect(wrapper.emitted('openDirectory')).toEqual([['app/templates']])
    expect(wrapper.emitted('openFile')).toEqual([['app/app.py']])
  })

  it('编辑可修改文件后按下 Ctrl+S 应触发保存事件', async () => {
    const wrapper = mount(AWDDefenseFileWorkbench, {
      props: {
        serviceTitle: 'Bank Portal',
        loading: false,
        error: '',
        editablePaths: ['app/app.py'],
        directory: {
          path: 'app',
          entries: [{ name: 'app.py', path: 'app/app.py', type: 'file', size: 13 }],
        },
        file: {
          path: 'app/app.py',
          content: "print('vuln')",
          size: 13,
        },
      },
    })

    const textarea = wrapper.get('textarea')

    await textarea.setValue("print('fixed')")
    expect(wrapper.text()).toContain('未保存')
    await textarea.trigger('keydown', { key: 's', ctrlKey: true })

    expect(wrapper.emitted('saveFile')).toEqual([
      ['app/app.py', "print('fixed')"],
    ])
  })

  it('焦点离开 textarea 后按下 Ctrl+S 仍应触发保存事件', async () => {
    const wrapper = mount(AWDDefenseFileWorkbench, {
      props: {
        serviceTitle: 'Bank Portal',
        loading: false,
        error: '',
        editablePaths: ['app/app.py'],
        directory: {
          path: 'app',
          entries: [{ name: 'app.py', path: 'app/app.py', type: 'file', size: 13 }],
        },
        file: {
          path: 'app/app.py',
          content: "print('vuln')",
          size: 13,
        },
      },
    })

    await wrapper.get('textarea').setValue("print('fixed-again')")
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 's', ctrlKey: true }))

    expect(wrapper.emitted('saveFile')).toEqual([
      ['app/app.py', "print('fixed-again')"],
    ])
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

  it('只读文件应继续显示内容预览，而不是进入编辑态', () => {
    const wrapper = mount(AWDDefenseFileWorkbench, {
      props: {
        serviceTitle: 'Bank Portal',
        loading: false,
        error: '',
        editablePaths: ['app/utils/crypto.py'],
        directory: {
          path: 'app',
          entries: [{ name: 'app.py', path: 'app/app.py', type: 'file', size: 13 }],
        },
        file: {
          path: 'app/app.py',
          content: "print('readonly')",
          size: 17,
        },
      },
    })

    expect(wrapper.text()).toContain('只读')
    expect(wrapper.find('textarea').exists()).toBe(false)
    expect(wrapper.find('.defense-file-workbench__editor-shell').exists()).toBe(true)
    expect(wrapper.find('pre').text()).toContain("print('readonly')")
  })

  it('桌面端文件预览区应固定在视口内，并通过内部滚动承载长内容', () => {
    expect(awdDefenseFileWorkbenchSource).toContain('--defense-file-workbench-preview-top')
    expect(awdDefenseFileWorkbenchSource).toContain('position: sticky;')
    expect(awdDefenseFileWorkbenchSource).toContain(
      'max-height: calc(100vh - var(--defense-file-workbench-preview-top) - var(--space-4));'
    )
    expect(awdDefenseFileWorkbenchSource).not.toContain('defense-file-workbench__file-head')
    expect(awdDefenseFileWorkbenchSource).not.toContain('showcase-panel-card showcase-panel-card--minimal-wire')
  })
})
